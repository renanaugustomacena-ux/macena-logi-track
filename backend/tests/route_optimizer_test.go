package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/services"
)

// TestRouteOptimizerCache verifies that a second identical request is
// served from the in-memory LRU cache rather than hitting OSRM again.
func TestRouteOptimizerCache(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/route/v1/driving/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"Ok","routes":[{"distance":160000,"duration":9000,"geometry":"abc","legs":[{"distance":160000,"duration":9000}]}]}`))
	}))
	defer ts.Close()

	cfg := config.OSRMConfig{
		BaseURL:      ts.URL,
		Timeout:      2_000_000_000, // 2s
		AllowedHosts: []string{extractHost(ts.URL)},
		CacheSize:    10,
	}
	opt := services.NewOSRMOptimizer(cfg, zap.NewNop())

	req := services.RouteRequest{
		Waypoints: []models.GeoPoint{
			models.NewGeoPoint(10.793, 45.341),
			models.NewGeoPoint(9.214, 45.450),
		},
	}
	r1, err := opt.OptimiseRoute(context.Background(), req)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if r1.Source != "osrm" {
		t.Fatalf("expected osrm on first call, got %q", r1.Source)
	}
	r2, err := opt.OptimiseRoute(context.Background(), req)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if r2.Source != "cache" {
		t.Fatalf("expected cache on second call, got %q", r2.Source)
	}
	if calls != 1 {
		t.Fatalf("expected 1 OSRM call, got %d", calls)
	}
}

// TestRouteOptimizerAllowList rejects hosts not on the allow-list.
func TestRouteOptimizerAllowList(t *testing.T) {
	cfg := config.OSRMConfig{
		BaseURL:      "https://attacker.example.com",
		Timeout:      2_000_000_000,
		AllowedHosts: []string{"router.project-osrm.org"},
		CacheSize:    10,
	}
	opt := services.NewOSRMOptimizer(cfg, zap.NewNop())
	_, err := opt.OptimiseRoute(context.Background(), services.RouteRequest{
		Waypoints: []models.GeoPoint{
			models.NewGeoPoint(10, 45),
			models.NewGeoPoint(11, 46),
		},
	})
	if err == nil {
		t.Fatalf("expected allow-list rejection")
	}
}

// TestRouteOptimizerFallback verifies the straight-line + 70 km/h
// estimator kicks in when OSRM is unreachable.
func TestRouteOptimizerFallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always fail so the optimiser falls back.
		w.WriteHeader(500)
	}))
	defer ts.Close()
	cfg := config.OSRMConfig{
		BaseURL:      ts.URL,
		Timeout:      1_000_000_000,
		AllowedHosts: []string{extractHost(ts.URL)},
		CacheSize:    10,
	}
	opt := services.NewOSRMOptimizer(cfg, zap.NewNop())
	r, err := opt.OptimiseRoute(context.Background(), services.RouteRequest{
		Waypoints: []models.GeoPoint{
			models.NewGeoPoint(10.793, 45.341), // Mozzecane
			models.NewGeoPoint(10.965, 45.398), // Quadrante Europa (~ 7 km)
		},
	})
	if err != nil {
		t.Fatalf("fallback returned error: %v", err)
	}
	if r.Source != "fallback" {
		t.Fatalf("expected fallback, got %q", r.Source)
	}
	if r.Distance <= 0 || r.Duration <= 0 {
		t.Fatalf("fallback produced zero distance/duration: %+v", r)
	}
}

func extractHost(raw string) string {
	// httptest URLs look like "http://127.0.0.1:xxxx"; strip scheme and port.
	s := strings.TrimPrefix(raw, "http://")
	s = strings.TrimPrefix(s, "https://")
	if i := strings.Index(s, ":"); i >= 0 {
		return s[:i]
	}
	return s
}
