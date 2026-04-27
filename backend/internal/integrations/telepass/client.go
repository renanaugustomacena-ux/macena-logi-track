// Package telepass is the client for the ViaCard/Telepass toll-fusion
// service. Telepass S.p.A. distributes the on-board transponders used
// on every Italian motorway; its B2B "Telepass Business" portal
// publishes per-vehicle toll events that LogiTrack fuses with
// telematics waypoints to reconstruct the chain-of-custody of a
// shipment.
//
// Telepass exposes two primary ingestion paths: a pull REST API
// (rate-limited, used for hourly reconciliation) and a daily CSV drop
// via SFTP. This package implements the REST path directly and accepts
// a CSV import buffer for back-fill of historical data. If the REST
// credentials are missing, the client returns ErrNotConfigured and the
// CSV importer still works — legitimate for customers on the SFTP-only
// tier.
package telepass

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
	"strconv"
	"strings"
	"time"
)

// ErrNotConfigured is returned from REST calls when the Telepass
// Business credentials are not set.
var ErrNotConfigured = errors.New(
	"integrations/telepass: not configured — set LOGITRACK_TELEPASS_API_BASE " +
		"(e.g. https://api.telepass.com/business/v1), LOGITRACK_TELEPASS_API_KEY " +
		"(issued by the Telepass Business portal) and LOGITRACK_TELEPASS_CONTRACT_ID " +
		"(numeric contract id). CSV import still works via ImportCSV without " +
		"credentials; the REST pull does not.")

// ErrUpstream wraps non-2xx responses.
type ErrUpstream struct {
	StatusCode int
	Body       string
}

func (e *ErrUpstream) Error() string {
	return fmt.Sprintf("integrations/telepass: upstream %d: %s", e.StatusCode, e.Body)
}

// Config resolves from process environment.
type Config struct {
	BaseURL    string
	APIKey     string
	ContractID string
	Timeout    time.Duration
}

// Client is the Telepass Business HTTP client.
type Client struct {
	cfg  Config
	http *http.Client
}

// New constructs the client. Empty config is legal and produces a
// client that returns ErrNotConfigured on every REST call.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// Configured reports whether the minimum REST credentials are present.
func (c *Client) Configured() bool {
	return strings.TrimSpace(c.cfg.BaseURL) != "" &&
		strings.TrimSpace(c.cfg.APIKey) != "" &&
		strings.TrimSpace(c.cfg.ContractID) != ""
}

// TollEvent is a single transit recorded by a Telepass gantry. Only
// the subset LogiTrack correlates with shipments is retained; the raw
// Telepass payload carries an order of magnitude more fields (fiscal
// nominal, contract id, class multiplier, etc.).
type TollEvent struct {
	EventID        string    `json:"eventId"`
	VehiclePlate   string    `json:"vehiclePlate"`
	TransponderSN  string    `json:"transponderSerial"`
	GantryID       string    `json:"gantryId"`
	GantryName     string    `json:"gantryName"`
	OccurredAt     time.Time `json:"occurredAt"`
	NetAmountCents int64     `json:"netAmountCents"`
}

// FetchEvents pulls the toll events for a vehicle plate during a time
// window. Returns ErrNotConfigured if REST credentials are missing.
func (c *Client) FetchEvents(ctx context.Context, plate string, from, to time.Time) ([]TollEvent, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	endpoint, err := url.JoinPath(c.cfg.BaseURL, "/transits")
	if err != nil {
		return nil, fmt.Errorf("integrations/telepass: build url: %w", err)
	}
	q := url.Values{}
	q.Set("contractId", c.cfg.ContractID)
	q.Set("plate", plate)
	q.Set("from", from.UTC().Format(time.RFC3339))
	q.Set("to", to.UTC().Format(time.RFC3339))
	endpoint += "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("integrations/telepass: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		return nil, &ErrUpstream{StatusCode: resp.StatusCode, Body: string(buf[:n])}
	}
	var out struct {
		Events []TollEvent `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("integrations/telepass: decode: %w", err)
	}
	return out.Events, nil
}

// ImportCSV parses the daily Telepass SFTP drop. The expected header
// row is:
//
//	event_id,plate,transponder,gantry_id,gantry_name,occurred_at,net_cents
//
// with occurred_at in RFC3339 and net_cents as a signed integer.
// This path does NOT require REST credentials.
func ImportCSV(r io.Reader) ([]TollEvent, error) {
	reader := csv.NewReader(bufio.NewReader(r))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("integrations/telepass: csv: %w", err)
	}
	if len(rows) < 1 {
		return nil, fmt.Errorf("integrations/telepass: empty csv")
	}
	// Header tolerance: we accept either the English or Italian header
	// that Telepass publishes depending on the tenant locale.
	headerLen := len(rows[0])
	if headerLen < 7 {
		return nil, fmt.Errorf("integrations/telepass: csv header has %d cols, want >=7", headerLen)
	}
	out := make([]TollEvent, 0, len(rows)-1)
	for i, row := range rows[1:] {
		if len(row) < 7 {
			continue
		}
		at, err := time.Parse(time.RFC3339, strings.TrimSpace(row[5]))
		if err != nil {
			return nil, fmt.Errorf("integrations/telepass: row %d time: %w", i+2, err)
		}
		net, err := strconv.ParseInt(strings.TrimSpace(row[6]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("integrations/telepass: row %d amount: %w", i+2, err)
		}
		out = append(out, TollEvent{
			EventID:        strings.TrimSpace(row[0]),
			VehiclePlate:   strings.ToUpper(strings.TrimSpace(row[1])),
			TransponderSN:  strings.TrimSpace(row[2]),
			GantryID:       strings.TrimSpace(row[3]),
			GantryName:     strings.TrimSpace(row[4]),
			OccurredAt:     at,
			NetAmountCents: net,
		})
	}
	return out, nil
}
