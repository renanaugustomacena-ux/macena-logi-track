package middleware

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/problem"
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
	base := &limiterStore{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(cfg.RPS),
		burst:    cfg.Burst,
	}
	overrideStores := make([]*limiterStore, len(overrides))
	for i, o := range overrides {
		overrideStores[i] = &limiterStore{
			limiters: make(map[string]*rate.Limiter),
			rps:      rate.Limit(o.RPS),
			burst:    o.Burst,
		}
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
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func (s *limiterStore) get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	lim, ok := s.limiters[key]
	if !ok {
		lim = rate.NewLimiter(s.rps, s.burst)
		s.limiters[key] = lim
	}
	return lim
}
