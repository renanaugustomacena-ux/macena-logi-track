// Package albo is the client for the Albo Nazionale degli
// Autotrasportatori, the national register of road-freight carriers
// maintained by the Italian Ministry of Infrastructure and Transport
// (MIT). The register must be consulted by operators before dispatching
// a load to verify the carrier is currently authorised (art. 1 L. 298/1974).
//
// The Albo publishes two endpoints: a machine-readable REST API (tier
// 1, flat-rate subscription) and a daily CSV bulletin (tier 0, free).
// LogiTrack supports both. If neither is configured the client returns
// ErrNotConfigured so the caller cannot accidentally mark a driver as
// verified against an empty dataset.
package albo

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNotConfigured names the env variables an operator must set.
var ErrNotConfigured = errors.New(
	"integrations/albo: not configured — set LOGITRACK_ALBO_API_BASE " +
		"(e.g. https://api.alboautotrasporto.it/v1) and LOGITRACK_ALBO_API_KEY " +
		"(issued by the Ministero Infrastrutture — Albo Autotrasportatori portal). " +
		"CSV bulletin verification via VerifyFromCSV does not require REST " +
		"credentials but does require a recent bulletin dump.")

// ErrUpstream wraps a non-2xx response from the Albo.
type ErrUpstream struct {
	StatusCode int
	Body       string
}

func (e *ErrUpstream) Error() string {
	return fmt.Sprintf("integrations/albo: upstream %d: %s", e.StatusCode, e.Body)
}

// Config resolves from process environment.
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client is the Albo Autotrasportatori HTTP client.
type Client struct {
	cfg  Config
	http *http.Client
}

// New builds a client. Empty config is legal at construction time; the
// error surfaces on the first Verify call.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 8 * time.Second
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// Configured reports whether both REST credentials are present.
func (c *Client) Configured() bool {
	return strings.TrimSpace(c.cfg.BaseURL) != "" && strings.TrimSpace(c.cfg.APIKey) != ""
}

// Registration is the subset of the Albo record LogiTrack surfaces.
type Registration struct {
	VATNumber       string    `json:"vatNumber"`
	CompanyName     string    `json:"companyName"`
	RegistrationNum string    `json:"registrationNumber"`
	Status          string    `json:"status"` // active, suspended, revoked
	RegisteredAt    time.Time `json:"registeredAt"`
	LastVerifiedAt  time.Time `json:"lastVerifiedAt"`
	Capabilities    []string  `json:"capabilities,omitempty"`
}

// Verify looks up a carrier by VAT number via the REST API. Returns
// ErrNotConfigured if no credentials are set.
func (c *Client) Verify(ctx context.Context, vatNumber string) (*Registration, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	if strings.TrimSpace(vatNumber) == "" {
		return nil, fmt.Errorf("integrations/albo: empty VAT number")
	}
	endpoint, err := url.JoinPath(c.cfg.BaseURL, "/carriers/", url.PathEscape(vatNumber))
	if err != nil {
		return nil, fmt.Errorf("integrations/albo: build url: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("integrations/albo: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("integrations/albo: vat %q not in register", vatNumber)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		return nil, &ErrUpstream{StatusCode: resp.StatusCode, Body: string(buf[:n])}
	}
	var reg Registration
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return nil, fmt.Errorf("integrations/albo: decode: %w", err)
	}
	reg.LastVerifiedAt = time.Now().UTC()
	return &reg, nil
}

// VerifyFromCSV verifies a VAT number against the free daily bulletin
// CSV. The Albo publishes the dump in Latin-1 with a `;` separator; the
// columns of interest are vat_number, registration_number, status. An
// absent VAT yields ("", false, nil); a match yields the record and
// (..., true, nil). Errors reflect parser problems only.
func VerifyFromCSV(r io.Reader, vatNumber string) (*Registration, bool, error) {
	reader := csv.NewReader(bufio.NewReader(r))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, false, fmt.Errorf("integrations/albo: csv: %w", err)
	}
	want := strings.TrimSpace(strings.ToUpper(vatNumber))
	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(row[0])) != want {
			continue
		}
		reg := &Registration{
			VATNumber:       strings.TrimSpace(row[0]),
			RegistrationNum: strings.TrimSpace(row[1]),
			Status:          strings.ToLower(strings.TrimSpace(row[2])),
			LastVerifiedAt:  time.Now().UTC(),
		}
		if len(row) > 3 {
			reg.CompanyName = strings.TrimSpace(row[3])
		}
		return reg, true, nil
	}
	return nil, false, nil
}
