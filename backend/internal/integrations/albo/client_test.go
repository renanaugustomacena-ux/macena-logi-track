package albo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotConfigured(t *testing.T) {
	c := New(Config{})
	_, err := c.Verify(context.Background(), "IT12345678901")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestVerifyOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != "k" {
			t.Fatalf("unexpected api key: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"vatNumber":"IT12345678901","status":"active","registrationNumber":"VR123"}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, APIKey: "k"})
	reg, err := c.Verify(context.Background(), "IT12345678901")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if reg.Status != "active" {
		t.Fatalf("status: %q", reg.Status)
	}
}

func TestVerifyFromCSV(t *testing.T) {
	data := "vat;reg_no;status;name\n" +
		"IT12345678901;VR123;active;Trasporti Alpha Srl\n" +
		"IT98765432109;MI456;suspended;Beta Logistica\n"
	reg, found, err := VerifyFromCSV(strings.NewReader(data), "IT12345678901")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !found || reg.Status != "active" {
		t.Fatalf("match fail: %+v", reg)
	}
	_, found, err = VerifyFromCSV(strings.NewReader(data), "IT00000000000")
	if err != nil || found {
		t.Fatalf("expected not found, got %v found=%v", err, found)
	}
}
