// Package handlers holds the Gin HTTP handlers. Each file is a single
// logical resource or concern (health, shipments, tracking, stream)
// and wires directly to a service method — handlers contain no
// business logic beyond request validation.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/repository"
)

// HealthHandler exposes /api/health. The response shape follows the
// platform-wide contract documented in docs/API.md so every
// LogiTrack-family project answers the same question identically.
type HealthHandler struct {
	cfg       *config.Config
	mongo     *repository.MongoRepository
	redis     *repository.RedisRepository
	startedAt time.Time
}

// NewHealthHandler constructs the handler.
func NewHealthHandler(cfg *config.Config, mongo *repository.MongoRepository, redis *repository.RedisRepository) *HealthHandler {
	return &HealthHandler{cfg: cfg, mongo: mongo, redis: redis, startedAt: time.Now().UTC()}
}

// Get reports liveness plus per-dependency status.
func (h *HealthHandler) Get(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	deps := map[string]string{
		"mongodb": "unknown",
		"redis":   "unknown",
	}

	if h.mongo != nil {
		if err := h.mongo.Client().Ping(ctx, nil); err == nil {
			deps["mongodb"] = "ok"
		} else {
			deps["mongodb"] = "degraded"
		}
	}
	if h.redis != nil {
		if err := h.redis.Client().Ping(ctx).Err(); err == nil {
			deps["redis"] = "ok"
		} else {
			deps["redis"] = "degraded"
		}
	}

	overall := "ok"
	for _, v := range deps {
		if v != "ok" {
			overall = "degraded"
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          overall,
		"service":         h.cfg.App.Name,
		"version":         h.cfg.App.Version,
		"uptime_seconds":  int64(time.Since(h.startedAt).Seconds()),
		"time":            time.Now().UTC().Format(time.RFC3339Nano),
		"dependencies":    deps,
	})
}
