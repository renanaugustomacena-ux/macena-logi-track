// Package rfi is the client for the FERTRAM/RFI intermodal rail-slot
// reservation API. Rete Ferroviaria Italiana (RFI) is the state-owned
// infrastructure manager for the Italian rail network; FERTRAM is the
// freight booking portal used by operators at the Quadrante Europa
// intermodal terminal.
//
// The public API is protected by mutual-TLS and a rotating client
// certificate. LogiTrack customers must supply the certificate bundle
// and the booking portal endpoint through environment variables. When
// any of the required inputs is missing the client returns
// ErrNotConfigured so the failure is diagnosable; it never returns
// silently empty slot lists.
package rfi

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

// ErrNotConfigured is returned when the RFI/FERTRAM integration is
// invoked without credentials. The message names every required input.
var ErrNotConfigured = errors.New(
	"integrations/rfi: not configured — set LOGITRACK_RFI_API_BASE " +
		"(e.g. https://fertram.rfi.it/api/v2), LOGITRACK_RFI_CLIENT_ID " +
		"(FERTRAM operator id) and LOGITRACK_RFI_CLIENT_SECRET (OAuth2 secret). " +
		"An mTLS certificate bundle may additionally be required via " +
		"LOGITRACK_RFI_MTLS_CERT_FILE / LOGITRACK_RFI_MTLS_KEY_FILE.")

// ErrUpstream is a non-2xx response from the FERTRAM portal.
type ErrUpstream struct {
	StatusCode int
	Body       string
}

func (e *ErrUpstream) Error() string {
	return fmt.Sprintf("integrations/rfi: upstream %d: %s", e.StatusCode, e.Body)
}

// Config resolves from process environment.
type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	MTLSCertFile string
	MTLSKeyFile  string
	Timeout      time.Duration
}

// Client is the FERTRAM/RFI HTTP client.
type Client struct {
	cfg  Config
	http *http.Client
}

// New constructs the client. Missing inputs are not an error at
// construction time; the error surfaces on the first call.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// Configured reports whether the minimal credential set is present.
// The mTLS bundle is optional and FERTRAM may additionally require a
// certificate for production; operators see a distinct warning in that
// case logged at startup.
func (c *Client) Configured() bool {
	return strings.TrimSpace(c.cfg.BaseURL) != "" &&
		strings.TrimSpace(c.cfg.ClientID) != "" &&
		strings.TrimSpace(c.cfg.ClientSecret) != ""
}

// RailSlot is a train-path reservation at the Quadrante Europa terminal
// or any other RFI-managed intermodal facility. Only the fields
// LogiTrack surfaces to the consignor dashboard are retained.
type RailSlot struct {
	ID        string    `json:"id"`
	Terminal  string    `json:"terminal"`
	TrainPath string    `json:"trainPath"`
	Departure time.Time `json:"departure"`
	Arrival   time.Time `json:"arrival"`
	Capacity  int       `json:"capacity"`
	Available int       `json:"available"`
	Status    string    `json:"status"` // proposed, confirmed, cancelled
}

// ListAvailableSlots returns rail-slot candidates on a route for the
// requested time window. Returns ErrNotConfigured if credentials are
// missing.
func (c *Client) ListAvailableSlots(ctx context.Context, from, to string, after time.Time) ([]RailSlot, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	q.Set("after", after.UTC().Format(time.RFC3339))
	endpoint, err := url.JoinPath(c.cfg.BaseURL, "/slots")
	if err != nil {
		return nil, fmt.Errorf("integrations/rfi: build url: %w", err)
	}
	endpoint += "?" + q.Encode()
	resp, err := httpretry.Do(ctx, func() (*http.Response, error) {
		req, rerr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if rerr != nil {
			return nil, rerr
		}
		req.Header.Set("X-Client-Id", c.cfg.ClientID)
		req.Header.Set("X-Client-Secret", c.cfg.ClientSecret)
		req.Header.Set("Accept", "application/json")
		return c.http.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("integrations/rfi: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		return nil, &ErrUpstream{StatusCode: resp.StatusCode, Body: string(buf[:n])}
	}
	var out struct {
		Slots []RailSlot `json:"slots"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("integrations/rfi: decode: %w", err)
	}
	return out.Slots, nil
}
