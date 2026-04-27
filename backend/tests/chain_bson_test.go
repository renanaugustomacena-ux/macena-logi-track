package tests

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/services"
)

// TestHashChainSurvivesBSONRoundTrip is the regression test for the
// chain-of-custody integrity defect identified during the
// 2026-04-27 audit. The defect mechanic:
//
//   1. computeHash json.Marshals a CustodyRecord that includes
//      time.Time fields with monotonic-clock readings and nanosecond
//      precision.
//   2. Mongo persists time.Time as BSON Date, which has millisecond
//      precision and no monotonic-clock semantics.
//   3. After read-back, the same record decodes to the same wall time
//      truncated to the millisecond. Recomputing the hash on the
//      decoded record produces a different SHA-256 because the marshal
//      output differs by sub-millisecond bytes.
//   4. VerifyChain compares the stored hash to the recomputed hash and
//      sees a mismatch — the chain is reported broken on every read,
//      even when nobody tampered with it.
//
// The fix in computeHash canonicalises every time field to UTC at
// millisecond precision before marshalling. Because BSON also stores
// times in UTC at millisecond precision, the in-memory canonical form
// and the post-round-trip canonical form are byte-identical.
//
// This test exercises the round-trip through bson.Marshal /
// bson.Unmarshal — the same encoding the official mongo-driver uses.
func TestHashChainSurvivesBSONRoundTrip(t *testing.T) {
	// time.Now() captures both monotonic and wall clocks at nanosecond
	// precision — the worst case for the defect. We deliberately do
	// NOT round it here; the fix must absorb arbitrary precision.
	now := time.Now()
	records := buildChain(now)

	// Sanity: in-memory chain must verify before we touch BSON.
	ok, badSeq, err := services.VerifyChain(records)
	if !ok || err != nil {
		t.Fatalf("pre-roundtrip verify failed: ok=%v badSeq=%d err=%v", ok, badSeq, err)
	}

	// Round-trip every record through BSON, the same way Mongo would
	// persist and re-fetch it. After this loop, monotonic-clock data
	// is stripped and sub-millisecond precision is gone.
	roundtripped := make([]models.CustodyRecord, len(records))
	for i, r := range records {
		raw, err := bson.Marshal(r)
		if err != nil {
			t.Fatalf("bson marshal seq=%d: %v", r.Sequence, err)
		}
		var decoded models.CustodyRecord
		if err := bson.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("bson unmarshal seq=%d: %v", r.Sequence, err)
		}
		roundtripped[i] = decoded
	}

	// The fix's contract: VerifyChain must still pass. Without the fix
	// we expect a hash mismatch on the very first record because the
	// pre-roundtrip Hash was computed against a nanosecond-precision
	// OccurredAt and the post-roundtrip Hash is recomputed against a
	// millisecond-precision OccurredAt.
	ok2, badSeq2, err2 := services.VerifyChain(roundtripped)
	if !ok2 || err2 != nil {
		t.Fatalf("post-roundtrip verify failed (this is the audit defect): ok=%v badSeq=%d err=%v", ok2, badSeq2, err2)
	}
}
