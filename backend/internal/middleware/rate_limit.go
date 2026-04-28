package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/problem"
)

// rateLimitEvictAfter and rateLimitSweepInterval mirror the WS
// handshake limiter: per-IP buckets idle for more than evictAfter are
// removed by a background sweeper. Caps memory at recent-peer
// working-set size; without this, an attacker rotating source IPs
// drives the limiter map unbounded.
const (
	rateLimitEvictAfter     = 30 * time.Minute
	rateLimitSweepInterval  = 5 * time.Minute
)

// RouteOverride describes a per-route rate-limit override. Paths match
// with strings.HasPrefix so a single rule can guard a whole subtree
// (e.g. "/api/v1/shipments" covers all shipment mutations).
type RouteOverride struct {
	PathPrefix string
	Method     string // empty matches any method
	RPS        int
	Burst      int
}

// DefaultOverrides encodes the baseline v2.0 §12 rate tiers: auth
// endpoints (5/min), mutations (60/min) and reads (600/min). The
// per-IP base limiter still enforces the global ceiling. These values
// are hand-translated to RPS (rounded up) because the limiter works in
// tokens/second, not requests/minute.
var DefaultOverrides = []RouteOverride{
	{PathPrefix: "/api/v1/auth", RPS: 1, Burst: 5},
	{PathPrefix: "/api/v1/shipments", Method: "POST", RPS: 1, Burst: 60},
	{PathPrefix: "/api/v1/shipments", Method: "GET", RPS: 10, Burst: 60},
}

// RateLimiter returns a per-IP token-bucket limiter suitable for a
// single-instance deployment. For multi-instance deployments the
// limiter should be backed by Redis (see docs/SECURITY-SELF-ASSESSMENT.md
// P-04); this implementation remains in-process.
//
// The middleware supports per-route overrides: if a route matches a
// rule in DefaultOverrides it uses a dedicated limiter keyed by
// (clientIP, rule-index). This implements v2.0 §12's stricter tier for
// auth and mutation endpoints.
func RateLimiter(cfg config.RateLimitConfig) gin.HandlerFunc {
	return RateLimiterWithOverrides(cfg, DefaultOverrides)
}

// RateLimiterWithOverrides builds the limiter with explicit overrides
// (used by tests and by integrators that need a bespoke policy).
func RateLimiterWithOverrides(cfg config.RateLimitConfig, overrides []RouteOverride) gin.HandlerFunc {
	base := newLimiterStore(rate.Limit(cfg.RPS), cfg.Burst)
	overrideStores := make([]*limiterStore, len(overrides))
	for i, o := range overrides {
		overrideStores[i] = newLimiterStore(rate.Limit(o.RPS), o.Burst)
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !base.get(ip).Allow() {
			problem.TooMany(c, "rate_limited", "global limit exceeded")
			return
		}
		for i, o := range overrides {
			if !matches(c, o) {
				continue
			}
			if !overrideStores[i].get(ip).Allow() {
				problem.TooMany(c, "rate_limited", "route limit exceeded")
				return
			}
		}
		c.Next()
	}
}

func matches(c *gin.Context, o RouteOverride) bool {
	if !strings.HasPrefix(c.FullPath(), o.PathPrefix) && !strings.HasPrefix(c.Request.URL.Path, o.PathPrefix) {
		return false
	}
	if o.Method != "" && !strings.EqualFold(c.Request.Method, o.Method) {
		return false
	}
	return true
}

type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	rps      rate.Limit
	burst    int
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newLimiterStore(rps rate.Limit, burst int) *limiterStore {
	s := &limiterStore{
		limiters: make(map[string]*limiterEntry),
		rps:      rps,
		burst:    burst,
	}
	go s.sweepLoop()
	return s
}

func (s *limiterStore) get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	e, ok := s.limiters[key]
	if !ok {
		e = &limiterEntry{
			limiter:  rate.NewLimiter(s.rps, s.burst),
			lastSeen: now,
		}
		s.limiters[key] = e
	} else {
		e.lastSeen = now
	}
	return e.limiter
}

func (s *limiterStore) sweep(cutoff time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for k, e := range s.limiters {
		if e.lastSeen.Before(cutoff) {
			delete(s.limiters, k)
			removed++
		}
	}
	return removed
}

func (s *limiterStore) sweepLoop() {
	t := time.NewTicker(rateLimitSweepInterval)
	defer t.Stop()
	for range t.C {
		s.sweep(time.Now().Add(-rateLimitEvictAfter))
	}
}
