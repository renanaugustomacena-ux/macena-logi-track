// Package aida contains the client for the Italian Customs Agency
// (Agenzia delle Dogane e dei Monopoli — AIDA) declaration and MRN
// lookup API. The public AIDA interface is the "Servizio Telematico
// Doganale" (STD) which issues NCTS T1/T2 transit declarations for
// cross-border road and rail freight.
//
// This package is intentionally a thin typed facade over the HTTP
// service. LogiTrack customers configure their AIDA credentials through
// the environment variables declared in `internal/config`. If those
// variables are empty the client refuses to silently return empty
// responses: it returns ErrNotConfigured so the caller can surface a
// diagnostic that names the missing variable. This matches the
// remediation brief rule "no silent stubs".
package aida

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/logitrack/backend/internal/integrations/httpretry"
)

// ErrNotConfigured is returned when the AIDA integration is invoked
// without the required credentials. The error message names the env
// variables the operator must set so the failure is self-explanatory
// in logs and the /api/ready endpoint.
var ErrNotConfigured = errors.New(
	"integrations/aida: not configured — set LOGITRACK_AIDA_API_BASE " +
		"(URL of the Agenzia delle Dogane STD endpoint) and LOGITRACK_AIDA_API_KEY " +
		"(API key issued by the Agenzia). Without both the AIDA client must not be used.")

// ErrUpstream wraps any non-2xx HTTP response from AIDA.
type ErrUpstream struct {
	StatusCode int
	Body       string
}

func (e *ErrUpstream) Error() string {
	return fmt.Sprintf("integrations/aida: upstream returned %d: %s", e.StatusCode, e.Body)
}

// Config captures the runtime parameters resolved from the process
// environment. BaseURL is the API base (e.g. https://api.adm.gov.it/std)
// and APIKey is the bearer secret.
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client is the AIDA HTTP client. It is safe for concurrent use.
type Client struct {
	cfg  Config
	http *http.Client
}

// New builds the client. Passing an empty BaseURL is legal — the
// resulting client will return ErrNotConfigured on every call. This
// allows the composition root to instantiate the dependency
// unconditionally and defer the "is it configured?" decision to the
// service layer.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Configured reports whether the client has both a base URL and an
// API key. Callers may use this for /api/ready checks to surface
// "AIDA integration not configured" distinctly from a transient
// network error.
func (c *Client) Configured() bool {
	return strings.TrimSpace(c.cfg.BaseURL) != "" && strings.TrimSpace(c.cfg.APIKey) != ""
}

// DeclarationStatus is the subset of AIDA's T1/T2 document lifecycle
// LogiTrack surfaces to the user. We intentionally pick the fields that
// the consignment-custody view needs rather than mirroring the full
// NCTS schema.
type DeclarationStatus struct {
	MRN            string    `json:"mrn"`
	State          string    `json:"state"` // accepted, in_transit, released, discharged
	DeclaredAt     time.Time `json:"declaredAt"`
	ClearancePoint string    `json:"clearancePoint,omitempty"`
	HSCode         string    `json:"hsCode,omitempty"`
}

// LookupMRN fetches the current status of a customs declaration by its
// Movement Reference Number. It returns ErrNotConfigured if the client
// was built without credentials.
func (c *Client) LookupMRN(ctx context.Context, mrn string) (*DeclarationStatus, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	if strings.TrimSpace(mrn) == "" {
		return nil, fmt.Errorf("integrations/aida: empty MRN")
	}
	endpoint, err := url.JoinPath(c.cfg.BaseURL, "/ncts/movements/", url.PathEscape(mrn))
	if err != nil {
		return nil, fmt.Errorf("integrations/aida: build url: %w", err)
	}
	resp, err := httpretry.Do(ctx, func() (*http.Response, error) {
		req, rerr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if rerr != nil {
			return nil, rerr
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		req.Header.Set("Accept", "application/json")
		return c.http.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("integrations/aida: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		return nil, &ErrUpstream{StatusCode: resp.StatusCode, Body: string(buf[:n])}
	}
	var out DeclarationStatus
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("integrations/aida: decode: %w", err)
	}
	return &out, nil
}
