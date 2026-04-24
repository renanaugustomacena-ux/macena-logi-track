package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
	"github.com/logitrack/backend/internal/services"
)

// TrackingHandler wraps TrackingService.
type TrackingHandler struct {
	svc *services.TrackingService
}

// NewTrackingHandler constructs the handler.
func NewTrackingHandler(svc *services.TrackingService) *TrackingHandler {
	return &TrackingHandler{svc: svc}
}

// LatestPosition handles GET /api/v1/tracking/:shipment/position.
func (h *TrackingHandler) LatestPosition(c *gin.Context) {
	claims := middleware.MustClaims(c)
	id := c.Param("id")
	wp, err := h.svc.LatestPosition(c.Request.Context(), claims.TenantID, id)
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "no_position", "no position cached for "+id)
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, wp)
}

// OptimiseRoute handles POST /api/v1/routes/optimize. The orchestrator
// is passed in at handler-construction time.
type RouteHandler struct {
	optimiser services.RouteOptimizer
}

// NewRouteHandler constructs the handler.
func NewRouteHandler(opt services.RouteOptimizer) *RouteHandler {
	return &RouteHandler{optimiser: opt}
}

// Optimize handles POST /api/v1/routes/optimize.
func (h *RouteHandler) Optimize(c *gin.Context) {
	var req services.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	res, err := h.optimiser.OptimiseRoute(c.Request.Context(), req)
	if err != nil {
		problem.Emit(c, http.StatusBadGateway, "optimise_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}
