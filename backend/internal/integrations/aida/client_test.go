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

func TestClientUpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, APIKey: "k"})
	_, err := c.LookupMRN(context.Background(), "m")
	var up *ErrUpstream
	if !errors.As(err, &up) {
		t.Fatalf("want ErrUpstream, got %v", err)
	}
}
