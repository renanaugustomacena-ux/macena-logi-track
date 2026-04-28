// Rifiuti handlers expose the rifiuti speciali domain (Produttori,
// Trasportatori, Destinatari, FIR) over HTTP. They follow the same
// thin-Gin-adapter shape as the fleet handlers because the
// regulatory rules sit in the rifiuti package — this file is just
// transport.
package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/modules/rifiuti"
	"github.com/logitrack/backend/internal/modules/rifiuti/rentri"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
)

// RifiutoHandler bundles the rifiuti CRUD + state-machine endpoints.
type RifiutoHandler struct {
	repo   *repository.MongoRepository
	rentri rentri.Client
}

// NewRifiutoHandler constructs the handler.
func NewRifiutoHandler(repo *repository.MongoRepository, client rentri.Client) *RifiutoHandler {
	return &RifiutoHandler{repo: repo, rentri: client}
}

// --- Produttori --------------------------------------------------------

// CreateProduttore handles POST /api/v1/rifiuti/produttori.
func (h *RifiutoHandler) CreateProduttore(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var p rifiuti.Produttore
	if err := c.ShouldBindJSON(&p); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	p.TenantID = claims.TenantID
	if p.RagioneSociale == "" || p.CodiceFiscale == "" {
		problem.Unprocessable(c, "missing_fields", "ragione_sociale and codice_fiscale are required")
		return
	}
	if err := h.repo.InsertProduttore(c.Request.Context(), &p); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ListProduttori handles GET /api/v1/rifiuti/produttori.
func (h *RifiutoHandler) ListProduttori(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, err := h.repo.ListProduttori(c.Request.Context(), claims.TenantID, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// --- Trasportatori -----------------------------------------------------

// CreateTrasportatore handles POST /api/v1/rifiuti/trasportatori.
func (h *RifiutoHandler) CreateTrasportatore(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var t rifiuti.Trasportatore
	if err := c.ShouldBindJSON(&t); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	t.TenantID = claims.TenantID
	if t.RagioneSociale == "" || t.CodiceFiscale == "" {
		problem.Unprocessable(c, "missing_fields", "ragione_sociale and codice_fiscale are required")
		return
	}
	if err := rifiuti.ValidateAlboCategoria(t.AlboCategoria); err != nil {
		problem.Unprocessable(c, "invalid_albo_categoria", err.Error())
		return
	}
	if err := h.repo.InsertTrasportatore(c.Request.Context(), &t); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, t)
}

// ListTrasportatori handles GET /api/v1/rifiuti/trasportatori.
func (h *RifiutoHandler) ListTrasportatori(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, err := h.repo.ListTrasportatori(c.Request.Context(), claims.TenantID, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// --- Destinatari -------------------------------------------------------

// CreateDestinatario handles POST /api/v1/rifiuti/destinatari.
func (h *RifiutoHandler) CreateDestinatario(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var d rifiuti.Destinatario
	if err := c.ShouldBindJSON(&d); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	d.TenantID = claims.TenantID
	if d.RagioneSociale == "" || d.CodiceFiscale == "" {
		problem.Unprocessable(c, "missing_fields", "ragione_sociale and codice_fiscale are required")
		return
	}
	if err := h.repo.InsertDestinatario(c.Request.Context(), &d); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, d)
}

// ListDestinatari handles GET /api/v1/rifiuti/destinatari.
func (h *RifiutoHandler) ListDestinatari(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, err := h.repo.ListDestinatari(c.Request.Context(), claims.TenantID, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// --- FIR ---------------------------------------------------------------

// CreateFIR handles POST /api/v1/rifiuti/fir. Validates the FIR
// shape, runs the Albo + autorizzazione cross-checks against the
// referenced trasportatore + destinatario, and persists in
// FIRDraft state. Returns 201 with the persisted document.
func (h *RifiutoHandler) CreateFIR(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var f rifiuti.FIR
	if err := c.ShouldBindJSON(&f); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	f.TenantID = claims.TenantID
	if err := f.Validate(); err != nil {
		problem.Unprocessable(c, "fir_invalid", err.Error())
		return
	}
	now := time.Now().UTC()
	tr, err := h.repo.GetTrasportatore(c.Request.Context(), claims.TenantID, f.TrasportatoreID)
	if errors.Is(err, repository.ErrNotFound) {
		problem.Unprocessable(c, "trasportatore_not_found", "trasportatore not registered for this tenant")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	if err := tr.CanCarry(f.CER, now); err != nil {
		problem.Unprocessable(c, "trasportatore_cannot_carry", err.Error())
		return
	}
	dest, err := h.repo.GetDestinatario(c.Request.Context(), claims.TenantID, f.DestinatarioID)
	if errors.Is(err, repository.ErrNotFound) {
		problem.Unprocessable(c, "destinatario_not_found", "destinatario not registered for this tenant")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	if err := dest.CanReceive(f.CER, f.OperazioneDestino, now); err != nil {
		problem.Unprocessable(c, "destinatario_cannot_receive", err.Error())
		return
	}
	if _, err := h.repo.GetProduttore(c.Request.Context(), claims.TenantID, f.ProduttoreID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			problem.Unprocessable(c, "produttore_not_found", "produttore not registered for this tenant")
			return
		}
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	f.State = rifiuti.FIRDraft
	if err := h.repo.InsertFIR(c.Request.Context(), &f); err != nil {
		problem.Internal(c, "create_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, f)
}

// GetFIR handles GET /api/v1/rifiuti/fir/:id.
func (h *RifiutoHandler) GetFIR(c *gin.Context) {
	claims := middleware.MustClaims(c)
	f, err := h.repo.GetFIR(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "FIR not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, f)
}

// ListFIR handles GET /api/v1/rifiuti/fir, optionally filtered by
// ?state=...
func (h *RifiutoHandler) ListFIR(c *gin.Context) {
	claims := middleware.MustClaims(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	state := rifiuti.FIRState(c.Query("state"))
	items, err := h.repo.ListFIR(c.Request.Context(), claims.TenantID, state, limit, offset)
	if err != nil {
		problem.Internal(c, "list_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// transitionRequest is the body shape for state transitions.
type transitionRequest struct {
	To string `json:"to"`
}

// TransitionFIR handles POST /api/v1/rifiuti/fir/:id/transition.
// Body: {"to": "<FIRState>"}. The state machine guard in
// FIR.AdvanceState rejects illegal edges; signature timestamps for
// the relevant role are stamped automatically.
func (h *RifiutoHandler) TransitionFIR(c *gin.Context) {
	claims := middleware.MustClaims(c)
	var req transitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	to := rifiuti.FIRState(req.To)
	f, err := h.repo.GetFIR(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "FIR not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	now := time.Now().UTC()
	if err := f.AdvanceState(to, now); err != nil {
		problem.Unprocessable(c, "invalid_transition", err.Error())
		return
	}
	switch to {
	case rifiuti.FIRConsegnatoTrasportatore:
		f.FirmaTrasportatoreAt = now
	case rifiuti.FIRConsegnatoDestinatario:
		f.FirmaDestinatarioAt = now
	}
	if err := h.repo.UpdateFIR(c.Request.Context(), f); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, f)
}

// VidimaFIR handles POST /api/v1/rifiuti/fir/:id/vidima. Submits the
// FIR to the RENTRI client (queued stub by default), captures the
// numero RENTRI + QR + signed XML, and advances the state to
// FIRVidimato. Idempotency-key derived deterministically from
// tenant + FIR id so retries on the same FIR yield the same numero.
func (h *RifiutoHandler) VidimaFIR(c *gin.Context) {
	claims := middleware.MustClaims(c)
	f, err := h.repo.GetFIR(c.Request.Context(), claims.TenantID, c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		problem.NotFound(c, "not_found", "FIR not found")
		return
	}
	if err != nil {
		problem.Internal(c, "lookup_failed", err.Error())
		return
	}
	if f.State != rifiuti.FIRDraft {
		problem.Unprocessable(c, "fir_not_draft", "vidimazione may only be requested on a FIRDraft")
		return
	}
	hash := sha256.Sum256([]byte(claims.TenantID + ":" + f.ID))
	idempotency := "FIR-" + hex.EncodeToString(hash[:8])
	// Placeholder xFIR payload until the RENTRI v1.0 XSD encoder lands.
	// Built via encoding/xml so user-controlled fields are escaped —
	// switching to the schema-driven encoder later will be a drop-in
	// replacement of the marshalled type.
	xfir, err := xml.Marshal(struct {
		XMLName xml.Name `xml:"formulario"`
		CER     string   `xml:"cer"`
	}{CER: string(f.CER)})
	if err != nil {
		problem.Internal(c, "xfir_marshal_failed", err.Error())
		return
	}
	resp, err := h.rentri.VidimaFIR(c.Request.Context(), rentri.VidimazioneRequest{
		TenantID:       claims.TenantID,
		IdempotencyKey: idempotency,
		XFIRPayload:    xfir,
	})
	if err != nil {
		problem.Internal(c, "rentri_failed", err.Error())
		return
	}
	now := time.Now().UTC()
	f.NumeroRENTRI = resp.NumeroRENTRI
	f.VidimatoAt = resp.VidimatoAt
	f.FirmaProduttoreAt = now
	if err := f.AdvanceState(rifiuti.FIRVidimato, now); err != nil {
		problem.Unprocessable(c, "invalid_transition", err.Error())
		return
	}
	if err := h.repo.UpdateFIR(c.Request.Context(), f); err != nil {
		problem.Internal(c, "update_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"fir":             f,
		"numero_rentri":   resp.NumeroRENTRI,
		"vidimato_at":     resp.VidimatoAt,
		"qr_code_payload": resp.QRCodePayload,
	})
}

// CERCheck handles GET /api/v1/rifiuti/cer/:code. Returns whether
// the supplied CER code is well-formed, whether it is pericoloso,
// and the chapter prefix. Useful for the operator UI to validate
// inline as the user types.
func (h *RifiutoHandler) CERCheck(c *gin.Context) {
	code := c.Param("code")
	if err := rifiuti.ValidateCER(code); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"input":      code,
			"valid":      false,
			"error":      err.Error(),
			"normalised": rifiuti.NormaliseCER(code),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"input":      code,
		"valid":      true,
		"normalised": rifiuti.NormaliseCER(code),
		"pericoloso": rifiuti.IsCERPericoloso(code),
		"chapter":    rifiuti.CERChapter(code),
	})
}
