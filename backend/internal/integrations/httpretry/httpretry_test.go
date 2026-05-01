package httpretry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestDoRetriesTransient503 confirms the helper retries on a 503
// and surfaces the eventual 200 to the caller.
func TestDoRetriesTransient503(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := Do(context.Background(), func() (*http.Response, error) {
		req, rerr := http.NewRequest(http.MethodGet, ts.URL, nil)
		if rerr != nil {
			return nil, rerr
		}
		return client.Do(req)
	})
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("expected 3 attempts (2 retries on 503 + 1 success), got %d", got)
	}
}

// TestDoStopsAtMaxAttempts confirms the helper gives up after
// MaxAttempts when every attempt is transient.
func TestDoStopsAtMaxAttempts(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "always busy", http.StatusBadGateway)
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := Do(context.Background(), func() (*http.Response, error) {
		req, rerr := http.NewRequest(http.MethodGet, ts.URL, nil)
		if rerr != nil {
			return nil, rerr
		}
		return client.Do(req)
	})
	if err == nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("expected an error after MaxAttempts of 502")
	}
	if got := calls.Load(); got != int32(MaxAttempts) {
		t.Fatalf("expected exactly %d attempts, got %d", MaxAttempts, got)
	}
}

// TestDoDoesNotRetry4xx: a 404 is a permanent client error and must
// be returned to the caller on the first attempt.
func TestDoDoesNotRetry4xx(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "no such mrn", http.StatusNotFound)
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := Do(context.Background(), func() (*http.Response, error) {
		req, rerr := http.NewRequest(http.MethodGet, ts.URL, nil)
		if rerr != nil {
			return nil, rerr
		}
		return client.Do(req)
	})
	if err != nil {
		t.Fatalf("404 should be returned not retried: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 surfaced, got %d", resp.StatusCode)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 attempt for a 404, got %d", got)
	}
}

// TestDoHonoursContextCancellation: a cancelled context aborts the
// retry loop without further attempts.
func TestDoHonoursContextCancellation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "busy", http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before the first attempt

	client := &http.Client{Timeout: 2 * time.Second}
	_, err := Do(ctx, func() (*http.Response, error) {
		req, rerr := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
		if rerr != nil {
			return nil, rerr
		}
		return client.Do(req)
	})
	if err == nil {
		t.Fatal("expected an error for a cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
