package handlers

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestIPRateLimiterRace hammers the limiter from many goroutines
// simultaneously with overlapping and distinct IPs. Run with -race
// (the suite default in CI). The previous map-without-mutex
// implementation panicked the runtime under this load.
func TestIPRateLimiterRace(t *testing.T) {
	t.Parallel()
	lim := newIPRateLimiter(1000, 1000)

	const goroutines = 64
	const callsPerG = 500

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		g := g
		go func() {
			defer wg.Done()
			for i := 0; i < callsPerG; i++ {
				// Mix of shared (10 hot IPs) and per-goroutine IPs so
				// the test exercises both contention on the same
				// bucket and concurrent insertion of new buckets.
				ip := fmt.Sprintf("10.0.0.%d", i%10)
				_ = lim.allow(ip)
				ip2 := fmt.Sprintf("10.%d.%d.%d", g, i%256, (i*g)%256)
				_ = lim.allow(ip2)
			}
		}()
	}
	wg.Wait()

	// Sanity: at least the 10 hot IPs are tracked; at most every
	// distinct IP we asked about (bounded above by what we generated).
	if got := lim.size(); got < 10 {
		t.Fatalf("size after race: %d, expected at least 10 hot IPs", got)
	}
}

// TestIPRateLimiterSweepEvictsIdle proves the eviction sweep
// removes buckets older than the cutoff while keeping fresh ones.
func TestIPRateLimiterSweepEvictsIdle(t *testing.T) {
	t.Parallel()
	lim := newIPRateLimiter(10, 20)

	// Insert two buckets; one will be marked stale by direct mutation
	// of lastSeen (we cannot wait 30 minutes in a unit test).
	if !lim.allow("1.1.1.1") {
		t.Fatal("first allow should succeed")
	}
	if !lim.allow("2.2.2.2") {
		t.Fatal("second allow should succeed")
	}
	if lim.size() != 2 {
		t.Fatalf("expected 2 buckets, got %d", lim.size())
	}

	// Backdate one bucket by an hour.
	lim.mu.Lock()
	lim.buckets["1.1.1.1"].lastSeen = time.Now().Add(-1 * time.Hour)
	lim.mu.Unlock()

	// Sweep with cutoff "30 minutes ago" — only the backdated bucket
	// should disappear.
	removed := lim.sweep(time.Now().Add(-30 * time.Minute))
	if removed != 1 {
		t.Fatalf("expected 1 bucket evicted, got %d", removed)
	}
	if lim.size() != 1 {
		t.Fatalf("expected 1 bucket remaining, got %d", lim.size())
	}
}

// TestIPRateLimiterAllowEnforcesBurst confirms the burst cap is
// honoured per IP — a single source cannot drive more than `burst`
// requests through faster than the bucket refills.
func TestIPRateLimiterAllowEnforcesBurst(t *testing.T) {
	t.Parallel()
	// 1 RPS, burst 3. The first 3 calls should succeed; the 4th
	// should be denied because the bucket has not refilled.
	lim := newIPRateLimiter(1, 3)
	for i := 0; i < 3; i++ {
		if !lim.allow("9.9.9.9") {
			t.Fatalf("call %d should have been allowed within burst", i+1)
		}
	}
	if lim.allow("9.9.9.9") {
		t.Fatal("call 4 should have been denied beyond burst")
	}
}
