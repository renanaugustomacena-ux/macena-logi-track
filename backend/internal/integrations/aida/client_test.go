package aida

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientNotConfigured(t *testing.T) {
	c := New(Config{})
	if c.Configured() {
		t.Fatalf("expected not configured")
	}
	_, err := c.LookupMRN(context.Background(), "24IT0000001")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestClientConfigured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mrn":"24IT0000001","state":"in_transit"}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, APIKey: "test-key"})
	if !c.Configured() {
		t.Fatalf("expected configured")
	}
	got, err := c.LookupMRN(context.Background(), "24IT0000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MRN != "24IT0000001" || got.State != "in_transit" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

// TestClientUpstreamError covers a permanent (non-transient) upstream
// error: 500 is "server is broken", not "server is busy", so the
// httpretry helper does NOT retry it and the client returns an
// ErrUpstream straight through.
func TestClientUpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, APIKey: "k"})
	_, err := c.LookupMRN(context.Background(), "m")
	var up *ErrUpstream
	if !errors.As(err, &up) {
		t.Fatalf("want ErrUpstream, got %v", err)
	}
	if up.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 surfaced as ErrUpstream, got %d", up.StatusCode)
	}
}

// TestClientRetriesTransient confirms the httpretry wrapper kicks
// in on a 503: the upstream returns 503 twice, then 200, and the
// caller sees the eventual success.
func TestClientRetriesTransient(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mrn":"24IT0000999","state":"released"}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, APIKey: "k"})
	got, err := c.LookupMRN(context.Background(), "24IT0000999")
	if err != nil {
		t.Fatalf("expected eventual success after 503 retries, got %v", err)
	}
	if got.MRN != "24IT0000999" || got.State != "released" {
		t.Fatalf("unexpected payload: %+v", got)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts (2 retries on 503 + 1 success), got %d", calls)
	}
}
