package rfi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientNotConfigured(t *testing.T) {
	c := New(Config{})
	if c.Configured() {
		t.Fatalf("expected not configured")
	}
	_, err := c.ListAvailableSlots(context.Background(), "VR", "MU", time.Now())
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestClientConfigured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slots":[{"id":"slot-1","terminal":"VR-QE","status":"confirmed"}]}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, ClientID: "id", ClientSecret: "s"})
	slots, err := c.ListAvailableSlots(context.Background(), "VR", "MU", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 1 || slots[0].ID != "slot-1" {
		t.Fatalf("unexpected payload: %+v", slots)
	}
}
