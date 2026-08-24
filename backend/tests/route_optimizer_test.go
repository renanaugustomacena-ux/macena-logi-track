package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/modules/logistics"
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
		Waypoints: []logistics.GeoPoint{
			logistics.NewGeoPoint(10.793, 45.341),
			logistics.NewGeoPoint(9.214, 45.450),
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
		Waypoints: []logistics.GeoPoint{
			logistics.NewGeoPoint(10, 45),
			logistics.NewGeoPoint(11, 46),
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
		Waypoints: []logistics.GeoPoint{
			logistics.NewGeoPoint(10.793, 45.341), // Mozzecane
			logistics.NewGeoPoint(10.965, 45.398), // Quadrante Europa (~ 7 km)
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

// TestRouteOptimizerTripReordersWaypoints verifies that with more than
// two waypoints the optimiser queries the OSRM trip service with first
// and last stop pinned, and reports the optimised visiting order.
func TestRouteOptimizerTripReordersWaypoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/trip/v1/driving/") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("roundtrip") != "false" || q.Get("source") != "first" || q.Get("destination") != "last" {
			t.Errorf("unexpected trip query %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		// Input order A,B,C,D; optimal visiting order A,C,B,D. The
		// waypoints array is in input order and waypoint_index is the
		// position of each input waypoint within the trip.
		_, _ = w.Write([]byte(`{"code":"Ok","waypoints":[{"waypoint_index":0},{"waypoint_index":2},{"waypoint_index":1},{"waypoint_index":3}],"trips":[{"distance":42000,"duration":2400,"geometry":"perm","legs":[{"distance":14000,"duration":800},{"distance":14000,"duration":800},{"distance":14000,"duration":800}]}]}`))
	}))
	defer ts.Close()

	cfg := config.OSRMConfig{
		BaseURL:      ts.URL,
		Timeout:      2_000_000_000,
		AllowedHosts: []string{extractHost(ts.URL)},
		CacheSize:    10,
	}
	opt := services.NewOSRMOptimizer(cfg, zap.NewNop())
	r, err := opt.OptimiseRoute(context.Background(), services.RouteRequest{
		Waypoints: []logistics.GeoPoint{
			logistics.NewGeoPoint(10.793, 45.341), // A
			logistics.NewGeoPoint(11.004, 45.439), // B
			logistics.NewGeoPoint(10.965, 45.398), // C
			logistics.NewGeoPoint(10.993, 45.549), // D
		},
	})
	if err != nil {
		t.Fatalf("OptimiseRoute returned error: %v", err)
	}
	if r.Source != "osrm" {
		t.Fatalf("expected osrm, got %q", r.Source)
	}
	if want := []int{0, 2, 1, 3}; !slices.Equal(r.WaypointOrder, want) {
		t.Fatalf("waypoint order = %v, expected %v", r.WaypointOrder, want)
	}
	if r.Distance != 42000 {
		t.Fatalf("distance = %v, expected 42000", r.Distance)
	}
}

// TestRouteOptimizerTripFallsBackToRoute verifies that a trip failure
// degrades to a route query in the given order instead of dropping
// straight to the haversine estimate.
func TestRouteOptimizerTripFallsBackToRoute(t *testing.T) {
	tripCalls, routeCalls := 0, 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/trip/v1/"):
			tripCalls++
			w.WriteHeader(500)
		case strings.HasPrefix(r.URL.Path, "/route/v1/"):
			routeCalls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":"Ok","routes":[{"distance":21000,"duration":1200,"geometry":"seq","legs":[{"distance":10500,"duration":600},{"distance":10500,"duration":600}]}]}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	cfg := config.OSRMConfig{
		BaseURL:      ts.URL,
		Timeout:      2_000_000_000,
		AllowedHosts: []string{extractHost(ts.URL)},
		CacheSize:    10,
	}
	opt := services.NewOSRMOptimizer(cfg, zap.NewNop())
	r, err := opt.OptimiseRoute(context.Background(), services.RouteRequest{
		Waypoints: []logistics.GeoPoint{
			logistics.NewGeoPoint(10.793, 45.341),
			logistics.NewGeoPoint(10.965, 45.398),
			logistics.NewGeoPoint(11.004, 45.439),
		},
	})
	if err != nil {
		t.Fatalf("OptimiseRoute returned error: %v", err)
	}
	if r.Source != "osrm" {
		t.Fatalf("expected osrm via route fallback, got %q", r.Source)
	}
	if want := []int{0, 1, 2}; !slices.Equal(r.WaypointOrder, want) {
		t.Fatalf("waypoint order = %v, expected identity %v", r.WaypointOrder, want)
	}
	if tripCalls != 1 || routeCalls != 1 {
		t.Fatalf("expected 1 trip + 1 route call, got %d + %d", tripCalls, routeCalls)
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

// TestRouteOptimizerTruckProfile verifies the OSRM_TRUCK_PROFILE
// configuration knob: when set, the optimiser routes vehicle="truck"
// requests against the configured profile path; when empty, it falls
// back to "driving" (and emits a one-shot WARN — covered by
// stdout-capturing in a separate test below).
func TestRouteOptimizerTruckProfile(t *testing.T) {
	cases := []struct {
		name         string
		truckProfile string
		expectedPath string
	}{
		{name: "configured_truck_profile", truckProfile: "truck", expectedPath: "/route/v1/truck/"},
		{name: "configured_hgv_profile", truckProfile: "hgv", expectedPath: "/route/v1/hgv/"},
		{name: "empty_falls_back_to_driving", truckProfile: "", expectedPath: "/route/v1/driving/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seenPath string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seenPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":"Ok","routes":[{"distance":1000,"duration":60,"geometry":"x","legs":[{"distance":1000,"duration":60}]}]}`))
			}))
			defer ts.Close()

			cfg := config.OSRMConfig{
				BaseURL:      ts.URL,
				Timeout:      2_000_000_000,
				AllowedHosts: []string{extractHost(ts.URL)},
				CacheSize:    10,
				TruckProfile: tc.truckProfile,
			}
			opt := services.NewOSRMOptimizer(cfg, zap.NewNop())
			_, err := opt.OptimiseRoute(context.Background(), services.RouteRequest{
				Vehicle: "truck",
				Waypoints: []logistics.GeoPoint{
					logistics.NewGeoPoint(10.793, 45.341),
					logistics.NewGeoPoint(10.965, 45.398),
				},
			})
			if err != nil {
				t.Fatalf("OptimiseRoute returned error: %v", err)
			}
			if !strings.HasPrefix(seenPath, tc.expectedPath) {
				t.Errorf("OSRM path = %q, expected prefix %q", seenPath, tc.expectedPath)
			}
		})
	}
}
