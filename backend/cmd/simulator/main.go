// Command simulator publishes GPS waypoints for the three demo shipments
// to the LogiTrack API at 1 Hz. It is used by the compose "demo"
// profile and the E2E test to validate the end-to-end stream.
//
// Behaviour:
//   - Mint a short-lived JWT matching JWT_SECRET / JWT_ISSUER (same
//     settings as the backend), so the simulator is just another
//     authenticated client — no magic internal path.
//   - For each demo route, interpolate between polyline vertices so the
//     per-second distance is ~1/3600 of the segment distance; publish a
//     waypoint at each tick. If LOOP is enabled, wrap to the start.
//   - Print a single-line status every 5 seconds (current coordinates,
//     shipment ID, publish latency) so "docker compose logs -f
//     logitrack-simulator" is friendly to operators.
//
// Golden-path contract:
//   - Within 30 seconds, every shipment has received at least 25
//     waypoints, the map marker has moved, and the ETA updates on each
//     WS broadcast.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/logitrack/backend/internal/demo"
)

type config struct {
	APIBase     string
	TenantID    string
	UserID      string
	JWTSecret   string
	JWTIssuer   string
	RateHz      float64
	Loop        bool
	DurationSec int
}

func mustEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envFloat(key string, fallback float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return fallback
}

func main() {
	cfg := config{
		APIBase:     mustEnv("LOGITRACK_API_BASE", "http://localhost:8080"),
		TenantID:    mustEnv("SIM_TENANT_ID", "demo-tenant"),
		UserID:      mustEnv("SIM_USER_ID", "demo-sim"),
		JWTSecret:   mustEnv("SIM_JWT_SECRET", "replace-me-with-a-real-256-bit-secret"),
		JWTIssuer:   mustEnv("SIM_JWT_ISSUER", "logitrack.it"),
		RateHz:      envFloat("SIM_RATE_HZ", 1),
		Loop:        envBool("SIM_LOOP", false),
		DurationSec: envInt("SIM_DURATION_SEC", 0),
	}
	token, err := mintToken(cfg)
	if err != nil {
		log.Fatalf("mint token: %v", err)
	}
	log.Printf("simulator starting; api=%s tenant=%s rate=%.1fHz loop=%v", cfg.APIBase, cfg.TenantID, cfg.RateHz, cfg.Loop)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Printf("simulator stopping on signal")
		cancel()
	}()
	if cfg.DurationSec > 0 {
		go func() {
			time.Sleep(time.Duration(cfg.DurationSec) * time.Second)
			cancel()
		}()
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Wait for the backend to accept the token. Retries with backoff
	// so the simulator can be started in parallel to the API.
	if err := waitReady(ctx, client, cfg.APIBase); err != nil {
		log.Fatalf("backend never became ready: %v", err)
	}

	var wg sync.WaitGroup
	for _, r := range demo.AllRoutes() {
		r := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			runRoute(ctx, cfg, client, token, r)
		}()
	}
	wg.Wait()
	log.Printf("simulator finished")
}

func runRoute(ctx context.Context, cfg config, client *http.Client, token string, r demo.Route) {
	period := time.Second
	if cfg.RateHz > 0 {
		period = time.Duration(float64(time.Second) / cfg.RateHz)
	}
	// Pre-interpolate the polyline so the per-tick move ≈ per-second
	// distance implied by the nominal speed (70 km/h ≈ 19.4 m/s; we
	// emit a point every 3–4 vertices assuming ~1 km spacing).
	points := interpolate(r.Polyline, 60) // 60 steps per original segment
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	idx := 0
	ticks := 0
	logEvery := int(5.0 / period.Seconds())
	if logEvery <= 0 {
		logEvery = 5
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if idx >= len(points) {
				if !cfg.Loop {
					log.Printf("[%s] arrived at destination", r.ID)
					return
				}
				idx = 0
			}
			p := points[idx]
			if err := postWaypoint(ctx, client, cfg, token, r.ID, p); err != nil {
				log.Printf("[%s] POST waypoint failed: %v", r.ID, err)
			} else if ticks%logEvery == 0 {
				log.Printf("[%s] idx=%d/%d pos=(%.4f,%.4f)", r.ID, idx, len(points), p.Lon, p.Lat)
			}
			idx++
			ticks++
		}
	}
}

type waypointBody struct {
	RecordedAt string `json:"recordedAt"`
	Position   struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"position"`
	SpeedKph   float64 `json:"speedKph"`
	HeadingDeg float64 `json:"headingDeg"`
	Source     string  `json:"source"`
}

func postWaypoint(ctx context.Context, client *http.Client, cfg config, token, shipmentID string, p demo.Waypoint) error {
	url := fmt.Sprintf("%s/api/v1/shipments/%s/waypoints", cfg.APIBase, shipmentID)
	body := waypointBody{
		RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
		SpeedKph:   70 + (float64(time.Now().UnixNano()%200)-100)/20, // 65–75 km/h jitter
		HeadingDeg: 0,
		Source:     "simulator",
	}
	body.Position.Type = "Point"
	body.Position.Coordinates = []float64{p.Lon, p.Lat}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func waitReady(ctx context.Context, client *http.Client, apiBase string) error {
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/api/health", nil)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for %s/api/health", apiBase)
}

func mintToken(cfg config) (string, error) {
	claims := jwt.MapClaims{
		"sub":      cfg.UserID,
		"tenantId": cfg.TenantID,
		"roles":    []string{"simulator", "operator"},
		"iss":      cfg.JWTIssuer,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(12 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.JWTSecret))
}

// interpolate adds `steps` intermediate points between successive
// polyline vertices so the simulator can emit smooth per-second
// updates without re-implementing a Bezier spline. The result
// preserves the starting and ending vertex exactly.
func interpolate(pts []demo.Waypoint, steps int) []demo.Waypoint {
	if steps < 1 {
		steps = 1
	}
	if len(pts) < 2 {
		return pts
	}
	out := make([]demo.Waypoint, 0, len(pts)*steps)
	for i := 0; i < len(pts)-1; i++ {
		a, b := pts[i], pts[i+1]
		for s := 0; s < steps; s++ {
			t := float64(s) / float64(steps)
			out = append(out, demo.Waypoint{
				Lon: a.Lon + (b.Lon-a.Lon)*t,
				Lat: a.Lat + (b.Lat-a.Lat)*t,
			})
		}
	}
	out = append(out, pts[len(pts)-1])
	return out
}
