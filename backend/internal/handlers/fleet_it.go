package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/modules/fleet_it"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
)

// FleetITHandler groups the driver device and telematics unit endpoints.
type FleetITHandler struct {
	repo *repository.MongoRepository
}

// NewFleetITHandler constructs the handler.
func NewFleetITHandler(repo *repository.MongoRepository) *FleetITHandler {
	return &FleetITHandler{repo: repo}
}

// --- Driver Devices ------------------------------------------------------

func (h *FleetITHandler) CreateDevice(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var d fleet_it.DriverDevice
	if err := c.ShouldBindJSON(&d); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	d.TenantID = claims.TenantID
	if d.Status == "" {
		d.Status = fleet_it.DeviceSpare
	}
	if err := d.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.InsertDriverDevice(c.Request.Context(), &d); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *FleetITHandler) GetDevice(c *gin.Context) {
	claims := middleware.MustClaims(c)
	d, err := h.repo.GetDriverDevice(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "device not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *FleetITHandler) ListDevices(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, offset := clampPage(c.DefaultQuery("limit", "100"), c.DefaultQuery("offset", "0"))
	items, err := h.repo.ListDriverDevices(c.Request.Context(), claims.TenantID, c.Query("driver_id"), c.Query("status"), limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

func (h *FleetITHandler) UpdateDevice(c *gin.Context) {
	claims := middleware.MustClaims(c)
	existing, err := h.repo.GetDriverDevice(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "device not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	origID, origCreated := existing.ID, existing.CreatedAt
	if err := c.ShouldBindJSON(existing); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	existing.ID = origID
	existing.CreatedAt = origCreated
	existing.TenantID = claims.TenantID
	if err := existing.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.UpdateDriverDevice(c.Request.Context(), existing); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}

// --- Telematics Units ----------------------------------------------------

func (h *FleetITHandler) CreateTelematicsUnit(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var t fleet_it.TelematicsUnit
	if err := c.ShouldBindJSON(&t); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	t.TenantID = claims.TenantID
	if t.Status == "" {
		t.Status = fleet_it.TelematicsOffline
	}
	if err := t.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.InsertTelematicsUnit(c.Request.Context(), &t); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *FleetITHandler) GetTelematicsUnit(c *gin.Context) {
	claims := middleware.MustClaims(c)
	t, err := h.repo.GetTelematicsUnit(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "telematics unit not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *FleetITHandler) ListTelematicsUnits(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, offset := clampPage(c.DefaultQuery("limit", "100"), c.DefaultQuery("offset", "0"))
	items, err := h.repo.ListTelematicsUnits(c.Request.Context(), claims.TenantID, c.Query("vehicle_id"), limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

func (h *FleetITHandler) UpdateTelematicsUnit(c *gin.Context) {
	claims := middleware.MustClaims(c)
	existing, err := h.repo.GetTelematicsUnit(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "telematics unit not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	origID, origCreated := existing.ID, existing.CreatedAt
	if err := c.ShouldBindJSON(existing); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	existing.ID = origID
	existing.CreatedAt = origCreated
	existing.TenantID = claims.TenantID
	if err := existing.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.UpdateTelematicsUnit(c.Request.Context(), existing); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}
