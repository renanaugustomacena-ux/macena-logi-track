package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/modules/itops"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
)

// ITOpsHandler groups the IT asset, incident, and monitoring endpoints.
type ITOpsHandler struct {
	repo *repository.MongoRepository
}

// NewITOpsHandler constructs the handler.
func NewITOpsHandler(repo *repository.MongoRepository) *ITOpsHandler {
	return &ITOpsHandler{repo: repo}
}

// --- IT Assets -----------------------------------------------------------

func (h *ITOpsHandler) CreateAsset(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var a itops.ITAsset
	if err := c.ShouldBindJSON(&a); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	a.TenantID = claims.TenantID
	if err := a.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.InsertITAsset(c.Request.Context(), &a); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *ITOpsHandler) GetAsset(c *gin.Context) {
	claims := middleware.MustClaims(c)
	a, err := h.repo.GetITAsset(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "asset not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ITOpsHandler) ListAssets(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	f := repository.AssetFilter{
		TenantID: claims.TenantID,
		Kind:     c.Query("kind"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
		Limit:    limit,
		Offset:   offset,
	}
	items, total, err := h.repo.ListITAssets(c.Request.Context(), f)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (h *ITOpsHandler) UpdateAsset(c *gin.Context) {
	claims := middleware.MustClaims(c)
	existing, err := h.repo.GetITAsset(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "asset not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	existing.TenantID = claims.TenantID
	if err := existing.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.UpdateITAsset(c.Request.Context(), existing); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *ITOpsHandler) AssetStats(c *gin.Context) {
	claims := middleware.MustClaims(c)
	counts, err := h.repo.CountITAssetsByKind(c.Request.Context(), claims.TenantID)
	if err != nil {
		problem.Internal(c, "stats_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"by_kind": counts})
}

// --- Incidents -----------------------------------------------------------

func (h *ITOpsHandler) CreateIncident(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var inc itops.Incident
	if err := c.ShouldBindJSON(&inc); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	inc.TenantID = claims.TenantID
	inc.Status = itops.IncidentOpen
	now := time.Now().UTC()
	inc.OpenedAt = now
	inc.SLATarget = now.Add(time.Duration(inc.Priority.SLAHours()) * time.Hour)
	if inc.Reporter == "" {
		inc.Reporter = claims.UserID
	}
	if err := inc.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.InsertIncident(c.Request.Context(), &inc); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	if inc.Reference == "" {
		inc.Reference = fmt.Sprintf("INC-%d-%s", now.Year(), inc.ID[:8])
		_ = h.repo.UpdateIncident(c.Request.Context(), &inc)
	}
	c.JSON(http.StatusCreated, inc)
}

func (h *ITOpsHandler) GetIncident(c *gin.Context) {
	claims := middleware.MustClaims(c)
	inc, err := h.repo.GetIncident(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "incident not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, inc)
}

func (h *ITOpsHandler) ListIncidents(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	f := repository.IncidentFilter{
		TenantID: claims.TenantID,
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Category: c.Query("category"),
		Limit:    limit,
		Offset:   offset,
	}
	items, total, err := h.repo.ListIncidents(c.Request.Context(), f)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (h *ITOpsHandler) TransitionIncident(c *gin.Context) {
	claims := middleware.MustClaims(c)
	inc, err := h.repo.GetIncident(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "incident not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}

	var body struct {
		Target     itops.IncidentStatus `json:"target"`
		AssignedTo string               `json:"assignedTo,omitempty"`
		Resolution string               `json:"resolution,omitempty"`
		Note       string               `json:"note,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	if !itops.CanTransition(inc.Status, body.Target) {
		problem.Unprocessable(c, "invalid_transition",
			fmt.Sprintf("cannot transition from %s to %s", inc.Status, body.Target))
		return
	}

	now := time.Now().UTC()
	inc.Status = body.Target
	if body.AssignedTo != "" {
		inc.AssignedTo = body.AssignedTo
	}
	if body.Resolution != "" {
		inc.Resolution = body.Resolution
	}
	if body.Target == itops.IncidentResolved {
		inc.ResolvedAt = &now
	}
	if body.Target == itops.IncidentClosed {
		inc.ClosedAt = &now
	}
	inc.WorkLog = append(inc.WorkLog, itops.WorkLogEntry{
		Author:    claims.UserID,
		Action:    fmt.Sprintf("transition to %s", body.Target),
		Note:      body.Note,
		Timestamp: now,
	})
	if err := h.repo.UpdateIncident(c.Request.Context(), inc); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, inc)
}

func (h *ITOpsHandler) AddWorkLog(c *gin.Context) {
	claims := middleware.MustClaims(c)
	inc, err := h.repo.GetIncident(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "incident not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}

	var entry itops.WorkLogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	entry.Author = claims.UserID
	entry.Timestamp = time.Now().UTC()
	inc.WorkLog = append(inc.WorkLog, entry)
	if err := h.repo.UpdateIncident(c.Request.Context(), inc); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, inc)
}

func (h *ITOpsHandler) IncidentStats(c *gin.Context) {
	claims := middleware.MustClaims(c)
	byStatus, err := h.repo.CountIncidentsByStatus(c.Request.Context(), claims.TenantID)
	if err != nil {
		problem.Internal(c, "stats_failed", err.Error())
		return
	}
	byPriority, err := h.repo.CountIncidentsByPriority(c.Request.Context(), claims.TenantID)
	if err != nil {
		problem.Internal(c, "stats_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"by_status": byStatus, "open_by_priority": byPriority})
}

// --- Monitoring Alerts ---------------------------------------------------

func (h *ITOpsHandler) ListAlerts(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, err := h.repo.ListMonitoringAlerts(c.Request.Context(), claims.TenantID, c.Query("status"), limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

func (h *ITOpsHandler) IngestAlert(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var a itops.MonitoringAlert
	if err := c.ShouldBindJSON(&a); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	a.TenantID = claims.TenantID
	if a.Title == "" {
		problem.Unprocessable(c, "missing_title", "alert title is required")
		return
	}
	if a.OccurredAt.IsZero() {
		a.OccurredAt = time.Now().UTC()
	}
	if a.Status == "" {
		a.Status = itops.AlertOpen
	}
	if err := h.repo.InsertMonitoringAlert(c.Request.Context(), &a); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *ITOpsHandler) AckAlert(c *gin.Context) {
	claims := middleware.MustClaims(c)
	found, err := h.repo.GetMonitoringAlert(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "alert not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	now := time.Now().UTC()
	found.Status = itops.AlertAcknowledged
	found.AckedAt = &now
	found.AckedBy = claims.UserID
	if err := h.repo.UpdateMonitoringAlert(c.Request.Context(), found); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, found)
}

func (h *ITOpsHandler) ResolveAlert(c *gin.Context) {
	claims := middleware.MustClaims(c)
	found, err := h.repo.GetMonitoringAlert(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "alert not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	now := time.Now().UTC()
	found.Status = itops.AlertResolved
	found.ResolvedAt = &now
	found.ResolvedBy = claims.UserID
	if err := h.repo.UpdateMonitoringAlert(c.Request.Context(), found); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, found)
}

func (h *ITOpsHandler) AlertStats(c *gin.Context) {
	claims := middleware.MustClaims(c)
	counts, err := h.repo.CountAlertsByStatus(c.Request.Context(), claims.TenantID)
	if err != nil {
		problem.Internal(c, "stats_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"by_status": counts})
}

// --- Monitoring Checks ---------------------------------------------------

func (h *ITOpsHandler) UpsertCheck(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var chk itops.MonitoringCheck
	if err := c.ShouldBindJSON(&chk); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	chk.TenantID = claims.TenantID
	if chk.CheckName == "" || chk.AssetID == "" {
		problem.Unprocessable(c, "missing_fields", "check_name and asset_id are required")
		return
	}
	if chk.LastChecked.IsZero() {
		chk.LastChecked = time.Now().UTC()
	}
	if chk.Status == "" {
		chk.Status = itops.CheckUnknown
	}
	if err := h.repo.UpsertMonitoringCheck(c.Request.Context(), &chk); err != nil {
		problem.Internal(c, "upsert_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, chk)
}

func (h *ITOpsHandler) ListChecks(c *gin.Context) {
	claims := middleware.MustClaims(c)
	items, err := h.repo.ListMonitoringChecks(c.Request.Context(), claims.TenantID, c.Query("asset_id"), c.Query("status"))
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// --- Licenses ------------------------------------------------------------

func (h *ITOpsHandler) CreateLicense(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var l itops.License
	if err := c.ShouldBindJSON(&l); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	l.TenantID = claims.TenantID
	if err := l.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.InsertLicense(c.Request.Context(), &l); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, l)
}

func (h *ITOpsHandler) GetLicense(c *gin.Context) {
	claims := middleware.MustClaims(c)
	l, err := h.repo.GetLicense(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "license not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *ITOpsHandler) ListLicenses(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.repo.ListLicenses(c.Request.Context(), repository.LicenseFilter{
		TenantID: claims.TenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (h *ITOpsHandler) UpdateLicense(c *gin.Context) {
	claims := middleware.MustClaims(c)
	existing, err := h.repo.GetLicense(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "license not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	existing.TenantID = claims.TenantID
	if err := existing.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.UpdateLicense(c.Request.Context(), existing); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *ITOpsHandler) LicenseCompliance(c *gin.Context) {
	claims := middleware.MustClaims(c)
	total, over, expiring, err := h.repo.LicenseComplianceSummary(c.Request.Context(), claims.TenantID)
	if err != nil {
		problem.Internal(c, "compliance_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total":          total,
		"over_deployed":  over,
		"expiring_soon":  expiring,
	})
}

// --- Runbooks ------------------------------------------------------------

func (h *ITOpsHandler) CreateRunbook(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var rb itops.Runbook
	if err := c.ShouldBindJSON(&rb); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	rb.TenantID = claims.TenantID
	if err := rb.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.InsertRunbook(c.Request.Context(), &rb); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, rb)
}

func (h *ITOpsHandler) GetRunbook(c *gin.Context) {
	claims := middleware.MustClaims(c)
	rb, err := h.repo.GetRunbook(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "runbook not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, rb)
}

func (h *ITOpsHandler) ListRunbooks(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.repo.ListRunbooks(c.Request.Context(), repository.RunbookFilter{
		TenantID: claims.TenantID,
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (h *ITOpsHandler) UpdateRunbook(c *gin.Context) {
	claims := middleware.MustClaims(c)
	existing, err := h.repo.GetRunbook(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "runbook not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	existing.TenantID = claims.TenantID
	existing.Version++
	if err := existing.Validate(); err != nil {
		problem.Unprocessable(c, "validation_failed", err.Error())
		return
	}
	if err := h.repo.UpdateRunbook(c.Request.Context(), existing); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}

// --- Dashboard Scadenze --------------------------------------------------

func (h *ITOpsHandler) DashboardScadenze(c *gin.Context) {
	claims := middleware.MustClaims(c)
	now := time.Now().UTC()
	ctx := c.Request.Context()

	var items []itops.Scadenza

	assets, err := h.repo.ListITAssetsWithWarrantyExpiry(ctx, claims.TenantID)
	if err != nil {
		problem.Internal(c, "scadenze_failed", err.Error())
		return
	}
	for _, a := range assets {
		if a.WarrantyExpiry == nil {
			continue
		}
		days := int(math.Ceil(a.WarrantyExpiry.Sub(now).Hours() / 24))
		items = append(items, itops.Scadenza{
			ID:       a.ID,
			Kind:     "warranty",
			Severity: itops.ComputeSeverity(days),
			Title:    fmt.Sprintf("%s — %s", a.Name, a.Manufacturer),
			Detail:   fmt.Sprintf("Garanzia scade il %s", a.WarrantyExpiry.Format("02/01/2006")),
			DueDate:  *a.WarrantyExpiry,
			DaysLeft: days,
		})
	}

	licenses, err := h.repo.ListLicensesWithExpiry(ctx, claims.TenantID)
	if err != nil {
		problem.Internal(c, "scadenze_failed", err.Error())
		return
	}
	for _, l := range licenses {
		if l.ExpiryDate == nil {
			continue
		}
		days := int(math.Ceil(l.ExpiryDate.Sub(now).Hours() / 24))
		items = append(items, itops.Scadenza{
			ID:       l.ID,
			Kind:     "license",
			Severity: itops.ComputeSeverity(days),
			Title:    fmt.Sprintf("%s (%s)", l.Software, l.Vendor),
			Detail:   fmt.Sprintf("Licenza scade il %s", l.ExpiryDate.Format("02/01/2006")),
			DueDate:  *l.ExpiryDate,
			DaysLeft: days,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *ITOpsHandler) DashboardOverview(c *gin.Context) {
	claims := middleware.MustClaims(c)
	ctx := c.Request.Context()
	tid := claims.TenantID

	assetsByKind, err := h.repo.CountITAssetsByKind(ctx, tid)
	if err != nil {
		problem.Internal(c, "overview_failed", err.Error())
		return
	}
	incByStatus, err := h.repo.CountIncidentsByStatus(ctx, tid)
	if err != nil {
		problem.Internal(c, "overview_failed", err.Error())
		return
	}
	incByPriority, err := h.repo.CountIncidentsByPriority(ctx, tid)
	if err != nil {
		problem.Internal(c, "overview_failed", err.Error())
		return
	}
	alertsByStatus, err := h.repo.CountAlertsByStatus(ctx, tid)
	if err != nil {
		problem.Internal(c, "overview_failed", err.Error())
		return
	}
	licTotal, licOver, licExpiring, err := h.repo.LicenseComplianceSummary(ctx, tid)
	if err != nil {
		problem.Internal(c, "overview_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets": gin.H{
			"by_kind": assetsByKind,
		},
		"incidents": gin.H{
			"by_status":        incByStatus,
			"open_by_priority": incByPriority,
		},
		"alerts": gin.H{
			"by_status": alertsByStatus,
		},
		"licenses": gin.H{
			"total":         licTotal,
			"over_deployed": licOver,
			"expiring_soon": licExpiring,
		},
	})
}
