package handlers

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/repository"
)

// ReadyHandler exposes GET /api/ready. Unlike /api/health it is
// binary: it returns 200 only when the service can accept new work
// (MongoDB writable, Redis reachable, optional seed completed). 503
// responses are authoritative for the orchestrator — the pod should
// be cordoned off.
//
// The platform distinction between liveness and readiness (v2.0 §6.5):
//
//   - /api/health answers "is the process alive?" — used by Kubernetes
//     livenessProbe, fails only on deadlock or crash.
//   - /api/ready answers "can I take traffic?" — used by readinessProbe
//     and by the LB before pulling a pod into rotation.
type ReadyHandler struct {
	cfg       *config.Config
	mongo     *repository.MongoRepository
	redis     *repository.RedisRepository
	seedDone  *atomic.Bool
	startedAt time.Time
}

// NewReadyHandler constructs the handler. seedDone is an atomic flag
// flipped to true after the optional demo seed completes; in
// non-seeding deployments it is set to true immediately.
func NewReadyHandler(cfg *config.Config, mongo *repository.MongoRepository, redis *repository.RedisRepository, seedDone *atomic.Bool) *ReadyHandler {
	return &ReadyHandler{cfg: cfg, mongo: mongo, redis: redis, seedDone: seedDone, startedAt: time.Now().UTC()}
}

// Get performs the readiness check with a short deadline so a stalled
// dependency never blocks kube-proxy.
func (h *ReadyHandler) Get(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{
		"mongodb": "ok",
		"redis":   "ok",
		"seed":    "ok",
	}
	ready := true

	if h.mongo != nil {
		// Round-trip a trivial command to prove we can talk to a primary.
		if err := h.mongo.Client().Ping(ctx, nil); err != nil {
			checks["mongodb"] = "fail"
			ready = false
		}
	}
	if h.redis != nil {
		if err := h.redis.Client().Ping(ctx).Err(); err != nil {
			checks["redis"] = "fail"
			ready = false
		}
	}
	if h.cfg.Demo.SeedOnBoot && !h.seedDone.Load() {
		checks["seed"] = "pending"
		ready = false
	}

	status := http.StatusOK
	verdict := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		verdict = "not_ready"
	}

	c.JSON(status, gin.H{
		"status":         verdict,
		"service":        h.cfg.App.Name,
		"version":        h.cfg.App.Version,
		"uptime_seconds": int64(time.Since(h.startedAt).Seconds()),
		"checks":         checks,
	})
}
