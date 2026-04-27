// Package httpretry provides a shared bounded-exponential retry
// wrapper used by every Italian-government / private-API
// integration client (AIDA, Albo, RFI, Telepass). The audit found
// that all four clients propagated transient upstream failures
// (502 / 503 / 504, network errors) directly to the LogiTrack
// caller as a 502 — turning a 30-second AIDA hiccup into a
// 30-second outage of every shipment that needs a customs lookup.
//
// Policy:
//   - Retry on network errors and on HTTP 502 / 503 / 504. Body of
//     the failed response is drained + closed before retry so the
//     transport's connection pool can reuse the underlying TCP.
//   - Do NOT retry on 4xx (client error — sending the same body
//     again will fail the same way) or on 5xx values outside the
//     transient set (500 is "server is broken", not "server is
//     busy"; 501 is "not implemented"; 505 is "unsupported HTTP
//     version" — none are retryable).
//   - Do NOT retry when the caller's context is already cancelled.
//   - Retry budget: 3 attempts total, exponential backoff
//     (200ms → 400ms → 800ms cap 5s) with full jitter so multiple
//     in-flight callers do not synchronise on the same retry tick.
//
// Idempotency: this helper is unconditional about retrying — the
// caller is expected to only wrap idempotent operations (GET,
// PUT, DELETE). Wrapping a non-idempotent POST (e.g. AIDA T1
// declaration submission) will need a per-call idempotency key
// that the upstream API supports. The current client surfaces
// only GETs, so this is safe.
package httpretry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// MaxAttempts is the total number of HTTP attempts (initial + 2 retries).
const MaxAttempts = 3

// transientStatuses is the set of HTTP status codes the helper
// considers transient and retryable.
var transientStatuses = map[int]struct{}{
	http.StatusBadGateway:         {}, // 502
	http.StatusServiceUnavailable: {}, // 503
	http.StatusGatewayTimeout:     {}, // 504
}

// Do wraps a single HTTP call with the retry policy described in
// the package comment. The `do` callback should rebuild the
// *http.Request on every invocation (because http.Request bodies
// are read-once); a typical caller pattern is:
//
//	resp, err := httpretry.Do(ctx, func() (*http.Response, error) {
//	    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
//	    if err != nil { return nil, err }
//	    req.Header.Set("Authorization", "Bearer "+key)
//	    return client.Do(req)
//	})
//
// On success the caller gets the final *http.Response (already-
// open Body) — closing the body remains the caller's responsibility,
// matching the contract of net/http.
func Do(ctx context.Context, do func() (*http.Response, error)) (*http.Response, error) {
	if ctx == nil {
		return nil, errors.New("httpretry: nil context")
	}
	op := func() (*http.Response, error) {
		// Surface caller cancellation immediately. backoff/v5 does
		// honour the context too, but checking here keeps the error
		// path a typed context.Cancelled rather than the generic
		// "max attempts reached" wrapper.
		if err := ctx.Err(); err != nil {
			return nil, backoff.Permanent(err)
		}
		resp, err := do()
		if err != nil {
			// Network / dial / TLS errors are transient. backoff will
			// retry until MaxAttempts.
			return nil, err
		}
		if _, ok := transientStatuses[resp.StatusCode]; ok {
			// Drain + close the body so the connection can be reused
			// for the retry. Discarding up to 4 KiB is enough to free
			// most Italian-API error envelopes; anything larger can
			// be re-fetched on the retry's success.
			_, _ = io.CopyN(io.Discard, resp.Body, 4096)
			_ = resp.Body.Close()
			return nil, fmt.Errorf("transient upstream HTTP %d", resp.StatusCode)
		}
		// 2xx / 3xx / 4xx and "permanent" 5xx are returned as-is.
		// Wrapping them in backoff.Permanent stops the retry loop.
		return resp, nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 200 * time.Millisecond
	bo.MaxInterval = 5 * time.Second
	bo.RandomizationFactor = 0.5

	return backoff.Retry(
		ctx,
		op,
		backoff.WithBackOff(bo),
		backoff.WithMaxTries(MaxAttempts),
	)
}
