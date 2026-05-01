package rentri

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// QueuedStub is the default Client adapter used in environments
// where the RENTRI digital certificate is not yet provisioned. Every
// call is accepted, deterministic stub identifiers are returned, and
// the payload is appended to an in-memory FIFO so the platform's
// regulatory dashboard can show the operator exactly what would
// have been transmitted to RENTRI.
//
// The stub is a deliberate design choice for Phase 0 of the
// RENTRI rollout: it lets the rest of the platform write production
// FIR data, exercise the full state machine, and keep all the
// downstream audit + chain-of-custody logic green. The day the
// design-partner trasportatore provides Entratel/SPID delegation
// and the RENTRI certificate, swapping this for an HTTP adapter is
// a one-line constructor change in cmd/server/main.go.
//
// QueuedStub is safe for concurrent use; the internal mutex
// protects both the in-memory store and the FIFO position counter.
type QueuedStub struct {
	mu          sync.Mutex
	now         func() time.Time
	pending     []QueuedItem
	firByNum    map[string]FIRSummary
	xmlByNum    map[string][]byte
	progressivo int64
}

// QueuedItem is one queued, unsent operation. The platform exposes
// the slice via Drain so an operator UI can show what is waiting.
type QueuedItem struct {
	Kind           QueuedKind
	TenantID       string
	Numero         string
	IdempotencyKey string
	Payload        []byte
	EnqueuedAt     time.Time
}

// QueuedKind discriminates the operation a queued item represents.
type QueuedKind string

const (
	KindVidimazione QueuedKind = "vidimazione"
	KindMovimento   QueuedKind = "movimento"
)

// NewQueuedStub returns a QueuedStub with the default time source.
// Tests may inject a deterministic clock via WithClock.
func NewQueuedStub() *QueuedStub {
	return &QueuedStub{
		now:      time.Now,
		firByNum: make(map[string]FIRSummary),
		xmlByNum: make(map[string][]byte),
	}
}

// WithClock returns a QueuedStub that reads the current time from
// the supplied function. Used by tests to assert deterministic
// vidimazione timestamps.
func (s *QueuedStub) WithClock(now func() time.Time) *QueuedStub {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now != nil {
		s.now = now
	}
	return s
}

// Drain returns and clears the queued items. Callers typically
// invoke it from a background worker that, once a real RENTRI
// adapter is configured, replays the items in order.
func (s *QueuedStub) Drain() []QueuedItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.pending
	s.pending = nil
	return out
}

// Pending returns a snapshot of the current queue without clearing.
// Useful for the operator dashboard.
func (s *QueuedStub) Pending() []QueuedItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]QueuedItem, len(s.pending))
	copy(out, s.pending)
	return out
}

// VidimaFIR returns a deterministic stub numero derived from the
// idempotency key + tenant id + payload SHA-256. Repeated calls
// with the same idempotency key return the same numero (RENTRI
// behaviour).
func (s *QueuedStub) VidimaFIR(_ context.Context, req VidimazioneRequest) (*VidimazioneResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyRequired
	}
	if len(req.XFIRPayload) == 0 {
		return nil, fmt.Errorf("%w: empty xFIR payload", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	numero := stubNumeroFor(req.TenantID, req.IdempotencyKey, req.XFIRPayload)
	now := s.now().UTC()
	resp := &VidimazioneResponse{
		NumeroRENTRI:  numero,
		VidimatoAt:    now,
		QRCodePayload: fmt.Sprintf("%s/fir/%s", SandboxPortalBase, numero),
		XFIRSignedXML: req.XFIRPayload, // stub: echo unsigned
	}
	s.firByNum[numero] = FIRSummary{
		NumeroRENTRI: numero,
		VidimatoAt:   now,
		State:        "vidimato",
	}
	s.xmlByNum[numero] = req.XFIRPayload
	s.pending = append(s.pending, QueuedItem{
		Kind:           KindVidimazione,
		TenantID:       req.TenantID,
		Numero:         numero,
		IdempotencyKey: req.IdempotencyKey,
		Payload:        req.XFIRPayload,
		EnqueuedAt:     now,
	})
	return resp, nil
}

// TrasmettiMovimento appends a movement entry to the queue and
// returns a deterministic progressive identifier.
func (s *QueuedStub) TrasmettiMovimento(_ context.Context, req MovimentoRequest) (*MovimentoResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyRequired
	}
	if len(req.XMLPayload) == 0 {
		return nil, fmt.Errorf("%w: empty movement payload", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progressivo++
	now := s.now().UTC()
	progressivo := fmt.Sprintf("STUB-%010d", s.progressivo)
	s.pending = append(s.pending, QueuedItem{
		Kind:           KindMovimento,
		TenantID:       req.TenantID,
		Numero:         progressivo,
		IdempotencyKey: req.IdempotencyKey,
		Payload:        req.XMLPayload,
		EnqueuedAt:     now,
	})
	return &MovimentoResponse{
		ProgressivoRENTRI: progressivo,
		AcceptedAt:        now,
	}, nil
}

// GetFIR returns a stored summary by numero, or ErrFIRNotFound.
func (s *QueuedStub) GetFIR(_ context.Context, numero string) (*FIRSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	summary, ok := s.firByNum[numero]
	if !ok {
		return nil, ErrFIRNotFound
	}
	return &summary, nil
}

// GetFIRXML returns the stored canonical XML by numero.
func (s *QueuedStub) GetFIRXML(_ context.Context, numero string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	xml, ok := s.xmlByNum[numero]
	if !ok {
		return nil, ErrFIRNotFound
	}
	out := make([]byte, len(xml))
	copy(out, xml)
	return out, nil
}

// ListFIRByCounterparty is a no-op in the stub because the stub
// stores no counterparty index. Returns an empty slice and a
// sentinel error so callers do not silently treat the empty
// response as authoritative.
func (s *QueuedStub) ListFIRByCounterparty(_ context.Context, _ string, _, _ time.Time) ([]FIRSummary, error) {
	return nil, errors.New("rentri: list-by-counterparty not implemented in queued stub")
}

// stubNumeroFor produces a deterministic alphanumeric identifier of
// the same shape RENTRI is expected to issue (~16 chars). The hash
// is truncated to 16 hex chars and uppercased so it is visually
// distinguishable from a real RENTRI number in operator dashboards.
func stubNumeroFor(tenantID, key string, payload []byte) string {
	h := sha256.New()
	h.Write([]byte(tenantID))
	h.Write([]byte{0})
	h.Write([]byte(key))
	h.Write([]byte{0})
	h.Write(payload)
	sum := h.Sum(nil)
	encoded := hex.EncodeToString(sum[:8])
	return "STUB-" + encoded // STUB- prefix prevents any chance of being mistaken for a live numero
}
