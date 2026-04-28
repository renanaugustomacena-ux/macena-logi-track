package tests

import (
	"testing"
	"time"

	"github.com/logitrack/backend/internal/modules/logistics"
	"github.com/logitrack/backend/internal/services"
)

// TestHashChain walks the canonical append/verify path: create a chain,
// mutate an interior record, and assert VerifyChain spots the breach.
func TestHashChain(t *testing.T) {
	now := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	records := buildChain(now)

	ok, badSeq, err := services.VerifyChain(records)
	if !ok || err != nil {
		t.Fatalf("expected clean chain, got ok=%v badSeq=%d err=%v", ok, badSeq, err)
	}

	// Tamper with the interior record's notes — hash should mismatch.
	tampered := append([]logistics.CustodyRecord(nil), records...)
	tampered[1].Notes = "tampered"
	ok2, _, err2 := services.VerifyChain(tampered)
	if ok2 || err2 == nil {
		t.Fatalf("expected tampered chain to fail verify; ok=%v err=%v", ok2, err2)
	}

	// Remove the head and verify we surface the mismatch at seq 2.
	shorter := append([]logistics.CustodyRecord(nil), records[1:]...)
	ok3, seq, err3 := services.VerifyChain(shorter)
	if ok3 || err3 == nil || seq != 2 {
		t.Fatalf("expected seq-2 break; ok=%v seq=%d err=%v", ok3, seq, err3)
	}
}

func buildChain(start time.Time) []logistics.CustodyRecord {
	base := []logistics.CustodyRecord{
		{
			Sequence:   1,
			ShipmentID: "S-1",
			Action:     logistics.CustodyCreated,
			OccurredAt: start,
			RecordedAt: start,
			Actor:      logistics.CustodyActor{Name: "system", Role: "creator"},
			Notes:      "initial",
		},
		{
			Sequence:   2,
			ShipmentID: "S-1",
			Action:     logistics.CustodyLoaded,
			OccurredAt: start.Add(1 * time.Hour),
			RecordedAt: start.Add(1 * time.Hour),
			Actor:      logistics.CustodyActor{Name: "driver", Role: "driver"},
			Notes:      "loaded at warehouse",
		},
		{
			Sequence:   3,
			ShipmentID: "S-1",
			Action:     logistics.CustodyHandover,
			OccurredAt: start.Add(3 * time.Hour),
			RecordedAt: start.Add(3 * time.Hour),
			Actor:      logistics.CustodyActor{Name: "hub", Role: "hub"},
			Notes:      "handover at hub",
		},
	}
	prev := ""
	for i := range base {
		base[i].PrevHash = prev
		base[i].Hash = services.ComputeHashForTest(&base[i])
		prev = base[i].Hash
	}
	return base
}
