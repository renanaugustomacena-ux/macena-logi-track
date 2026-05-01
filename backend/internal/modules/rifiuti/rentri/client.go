package rentri

import (
	"context"
	"errors"
	"time"
)

// VidimazioneRequest carries the pre-vidimazione FIR payload that
// the producer submits to RENTRI to obtain a unique numero +
// vidimazione virtuale + QR code, as required by art. 5 D.M.
// 59/2023. The payload is the canonical xFIR body minus the
// vidimazione fields RENTRI fills in on accept.
type VidimazioneRequest struct {
	TenantID       string
	XFIRPayload    []byte // canonical XML conforming to rentri-formulario-1.0.xsd
	IdempotencyKey string // stable per logical FIR; RENTRI 4xx on retry without it
}

// VidimazioneResponse carries the fields RENTRI returns on a
// successful vidimazione. NumeroRENTRI is the alphanumeric unique
// identifier of the FIR; QRCodePayload is the URL the QR encodes
// (for verification by inspectors with a phone camera).
type VidimazioneResponse struct {
	NumeroRENTRI  string
	VidimatoAt    time.Time
	QRCodePayload string
	XFIRSignedXML []byte // server-signed XAdES envelope
}

// MovimentoRequest is the per-movement record sent to RENTRI to
// append a carico / scarico line on the producer's, carrier's, or
// destinatario's registro cronologico. Submitted as XML conforming
// to rentri-movimenti-1.0.xsd.
type MovimentoRequest struct {
	TenantID       string
	XMLPayload     []byte
	IdempotencyKey string
}

// MovimentoResponse is the receipt RENTRI returns after a
// successful append. ProgressivoRENTRI is the unique number RENTRI
// assigns to the entry; the local platform stores both this and
// the operator-local progressive number for cross-reference.
type MovimentoResponse struct {
	ProgressivoRENTRI string
	AcceptedAt        time.Time
}

// FIRSummary is the lightweight projection RENTRI returns on a GET
// or list call. The full canonical XML is fetched on demand via
// GetFIRXML when an inspector or the producer requests it.
type FIRSummary struct {
	NumeroRENTRI    string
	VidimatoAt      time.Time
	State           string // RENTRI's state vocabulary, mirrors but does not equal FIRState
	ProduttoreCF    string
	TrasportatoreCF string
	DestinatarioCF  string
	CER             string
	QuantitaKg      float64
}

// Client is the interface every RENTRI integration must implement.
// Kept narrow on purpose: handlers + service layer talk to this,
// adapters (sandbox HTTP, queued stub, in-memory test double)
// implement it. Adding a new transport (mTLS vs OAuth2) is a new
// adapter, not a new method.
//
// All methods take a context.Context as the first argument so
// callers can apply deadlines and cancellation; idempotency keys
// are mandatory because RENTRI rejects retries that lack one.
type Client interface {
	// VidimaFIR submits a draft FIR for vidimazione and returns the
	// numero + QR + signed envelope.
	VidimaFIR(ctx context.Context, req VidimazioneRequest) (*VidimazioneResponse, error)

	// TrasmettiMovimento appends a carico/scarico entry to the
	// caller's registro cronologico for the supplied FIR.
	TrasmettiMovimento(ctx context.Context, req MovimentoRequest) (*MovimentoResponse, error)

	// GetFIR fetches the lightweight projection by numero.
	GetFIR(ctx context.Context, numero string) (*FIRSummary, error)

	// GetFIRXML fetches the canonical signed XML envelope for a FIR.
	GetFIRXML(ctx context.Context, numero string) ([]byte, error)

	// ListFIRByCounterparty returns the FIRs where the supplied
	// codice fiscale appears as produttore, trasportatore, or
	// destinatario, bounded by the supplied date window.
	ListFIRByCounterparty(ctx context.Context, codiceFiscale string, since, until time.Time) ([]FIRSummary, error)
}

// Common error sentinels every adapter must surface for the same
// upstream conditions. Service layer maps these to RFC 7807
// problem types in HTTP responses.
var (
	// ErrNotConfigured is returned when an adapter is constructed
	// without the credentials required to talk to RENTRI. The
	// integration boots and the rest of the platform stays usable;
	// only RENTRI calls fail with this error.
	ErrNotConfigured = errors.New("rentri: client not configured")

	// ErrUpstream wraps any non-2xx response from RENTRI that the
	// adapter cannot recover from automatically.
	ErrUpstream = errors.New("rentri: upstream error")

	// ErrValidation indicates the xFIR/movement payload was rejected
	// by server-side XSD or business validation. The adapter
	// preserves the original RENTRI problem-detail body so callers
	// can surface the specific reason.
	ErrValidation = errors.New("rentri: payload rejected by RENTRI validation")

	// ErrIdempotencyRequired is returned when a request reaches an
	// adapter without an IdempotencyKey. RENTRI mandates the header
	// on every state-changing call.
	ErrIdempotencyRequired = errors.New("rentri: idempotency key required")

	// ErrFIRNotFound is returned by GetFIR / GetFIRXML for unknown
	// numero values.
	ErrFIRNotFound = errors.New("rentri: FIR not found")
)
