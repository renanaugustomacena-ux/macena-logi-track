// Fleet handlers cover vehicles, drivers and geofences: the three
// collections defined under v2.0 §22.4 beyond the shipments + tracking
// core. They all follow the same shape — a thin Gin adapter over the
// repository CRUD — because the business rules for fleet-master data
// live with the operator, not with the platform.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
)

// FleetHandler groups the three fleet resources because they share
// the repository singleton and the same middleware surface.
type FleetHandler struct {
	repo *repository.MongoRepository
}

// NewFleetHandler constructs the handler.
func NewFleetHandler(repo *repository.MongoRepository) *FleetHandler {
	return &FleetHandler{repo: repo}
}

// --- Vehicles ----------------------------------------------------------

// CreateVehicle handles POST /api/v1/vehicles.
func (h *FleetHandler) CreateVehicle(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var v models.Vehicle
	if err := c.ShouldBindJSON(&v); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	v.TenantID = claims.TenantID
	if v.Plate == "" {
		problem.Unprocessable(c, "missing_plate", "plate is required")
		return
	}
	if err := models.ValidatePlate(v.Plate); err != nil {
		problem.Unprocessable(c, "invalid_plate", err.Error())
		return
	}
	v.Plate = models.NormalisePlate(v.Plate)
	if err := h.repo.InsertVehicle(c.Request.Context(), &v); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GetVehicle handles GET /api/v1/vehicles/:id.
func (h *FleetHandler) GetVehicle(c *gin.Context) {
	claims := middleware.MustClaims(c)
	v, err := h.repo.GetVehicle(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "vehicle not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, v)
}

// ListVehicles handles GET /api/v1/vehicles.
func (h *FleetHandler) ListVehicles(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	active := c.Query("active") == "true"
	items, err := h.repo.ListVehicles(c.Request.Context(), claims.TenantID, active, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// --- Drivers -----------------------------------------------------------

// CreateDriver handles POST /api/v1/drivers.
func (h *FleetHandler) CreateDriver(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var d models.Driver
	if err := c.ShouldBindJSON(&d); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	d.TenantID = claims.TenantID
	if d.Name == "" || d.Carrier == "" {
		problem.Unprocessable(c, "missing_fields", "name and carrier are required")
		return
	}
	if err := h.repo.InsertDriver(c.Request.Context(), &d); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, d)
}

// GetDriver handles GET /api/v1/drivers/:id.
func (h *FleetHandler) GetDriver(c *gin.Context) {
	claims := middleware.MustClaims(c)
	d, err := h.repo.GetDriver(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "driver not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, d)
}

// ListDrivers handles GET /api/v1/drivers.
func (h *FleetHandler) ListDrivers(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	active := c.Query("active") == "true"
	items, err := h.repo.ListDrivers(c.Request.Context(), claims.TenantID, active, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// --- Geofences ---------------------------------------------------------

// CreateGeofence handles POST /api/v1/geofences.
func (h *FleetHandler) CreateGeofence(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var g models.Geofence
	if err := c.ShouldBindJSON(&g); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	g.TenantID = claims.TenantID
	if g.Name == "" {
		problem.Unprocessable(c, "missing_name", "name is required")
		return
	}
	if g.Polygon.Type != "Polygon" || len(g.Polygon.Coordinates) == 0 {
		problem.Unprocessable(c, "invalid_polygon", "polygon must be a GeoJSON Polygon with at least one ring")
		return
	}
	if err := h.repo.InsertGeofence(c.Request.Context(), &g); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, g)
}

// GetGeofence handles GET /api/v1/geofences/:id.
func (h *FleetHandler) GetGeofence(c *gin.Context) {
	claims := middleware.MustClaims(c)
	g, err := h.repo.GetGeofence(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "geofence not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, g)
}

// ListGeofences handles GET /api/v1/geofences.
func (h *FleetHandler) ListGeofences(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	active := c.Query("active") == "true"
	kind := models.GeofenceType(c.Query("type"))
	items, err := h.repo.ListGeofences(c.Request.Context(), claims.TenantID, active, kind, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}
