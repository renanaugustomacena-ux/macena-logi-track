package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/modules/logistics"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
	"github.com/logitrack/backend/internal/services"
)

// ShipmentHandler glues the ShipmentService to Gin.
type ShipmentHandler struct {
	svc *services.ShipmentService
}

// NewShipmentHandler constructs the handler.
func NewShipmentHandler(svc *services.ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{svc: svc}
}

// Create handles POST /api/v1/shipments.
func (h *ShipmentHandler) Create(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var body logistics.Shipment
	if err := c.ShouldBindJSON(&body); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	body.TenantID = claims.TenantID
	if err := h.svc.CreateShipment(c.Request.Context(), &body); err != nil {
		problem.Unprocessable(c, "cannot_create", err.Error())
		return
	}
	c.JSON(http.StatusCreated, body)
}

// List handles GET /api/v1/shipments.
func (h *ShipmentHandler) List(c *gin.Context) {
	claims := middleware.MustClaims(c)
	filter := repository.ShipmentFilter{
		TenantID: claims.TenantID,
		Status:   c.Query("status"),
		Carrier:  c.Query("carrier"),
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = t
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Offset = n
		}
	}
	out, err := h.svc.ListShipments(c.Request.Context(), filter)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "limit": filter.Limit, "offset": filter.Offset})
}

// Get handles GET /api/v1/shipments/:id.
func (h *ShipmentHandler) Get(c *gin.Context) {
	claims := middleware.MustClaims(c)
	id := c.Param("id")
	s, err := h.svc.GetShipment(c.Request.Context(), claims.TenantID, id)
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "shipment "+id+" not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, s)
}

// AddWaypoint handles POST /api/v1/shipments/:id/waypoints.
func (h *ShipmentHandler) AddWaypoint(c *gin.Context) {
	claims := middleware.MustClaims(c)
	id := c.Param("id")
	var wp logistics.Waypoint
	if err := c.ShouldBindJSON(&wp); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	if err := h.svc.RecordWaypoint(c.Request.Context(), claims.TenantID, id, wp); err != nil {
		problem.Unprocessable(c, "cannot_record", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"accepted": true})
}

// Trace handles GET /api/v1/shipments/:id/trace.
func (h *ShipmentHandler) Trace(c *gin.Context) {
	claims := middleware.MustClaims(c)
	id := c.Param("id")
	records, err := h.svc.ChainOfCustody(c.Request.Context(), claims.TenantID, id)
	if err != nil {
		problem.Internal(c, "trace_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"shipmentId": id, "records": records})
}
