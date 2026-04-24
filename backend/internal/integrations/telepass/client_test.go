package telepass

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNotConfiguredFetch(t *testing.T) {
	c := New(Config{})
	_, err := c.FetchEvents(context.Background(), "EB123XY", time.Now(), time.Now())
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestFetchConfigured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"events":[{"eventId":"e1","vehiclePlate":"EB123XY","netAmountCents":512}]}`))
	}))
	defer srv.Close()
	c := New(Config{BaseURL: srv.URL, APIKey: "k", ContractID: "42"})
	evs, err := c.FetchEvents(context.Background(), "EB123XY", time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(evs) != 1 || evs[0].NetAmountCents != 512 {
		t.Fatalf("unexpected: %+v", evs)
	}
}

func TestImportCSV(t *testing.T) {
	data := "event_id,plate,transponder,gantry_id,gantry_name,occurred_at,net_cents\n" +
		"ev1,eb123xy,TSN-01,G1,Verona Nord,2026-04-01T08:30:00Z,550\n" +
		"ev2,eb123xy,TSN-01,G2,Vicenza Est,2026-04-01T09:05:00Z,320\n"
	got, err := ImportCSV(strings.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 rows, got %d", len(got))
	}
	if got[0].VehiclePlate != "EB123XY" {
		t.Fatalf("plate uppercased: %q", got[0].VehiclePlate)
	}
	if got[1].NetAmountCents != 320 {
		t.Fatalf("amount: %d", got[1].NetAmountCents)
	}
}
