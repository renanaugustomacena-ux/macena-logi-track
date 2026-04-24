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

// ETAHandler exposes GET /api/v1/shipments/:id/eta.
type ETAHandler struct {
	svc *services.ETAService
}

// NewETAHandler constructs the handler.
func NewETAHandler(svc *services.ETAService) *ETAHandler { return &ETAHandler{svc: svc} }

// Get returns the current ETA for a shipment.
func (h *ETAHandler) Get(c *gin.Context) {
	claims := middleware.MustClaims(c)
	id := c.Param("id")
	r, err := h.svc.ETA(c.Request.Context(), claims.TenantID, id)
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "shipment "+id+" not found")
		return
	}
	if errors.Is(err, services.ErrShipmentMissingCurrentPosition) {
		c.JSON(http.StatusAccepted, gin.H{"status": "awaiting_first_waypoint"})
		return
	}
	if err != nil {
		problem.Emit(c, http.StatusBadGateway, "eta_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, r)
}
