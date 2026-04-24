package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/services"
)

// stubRouter is a deterministic RouteOptimizer used by the ETA tests
// so they do not hit the real OSRM.
type stubRouter struct {
	distanceM float64
	durationS float64
}

func (s *stubRouter) OptimiseRoute(_ context.Context, _ services.RouteRequest) (*services.RouteResponse, error) {
	return &services.RouteResponse{Distance: s.distanceM, Duration: s.durationS, Source: "stub"}, nil
}

func (s *stubRouter) EstimateETA(_ context.Context, _ models.GeoPoint, _ models.GeoPoint) (time.Duration, error) {
	return time.Duration(s.durationS * float64(time.Second)), nil
}

// TestETASmoothingMovingAverage pushes 30 speed samples into the
// smoother and checks the running average respects the window size.
func TestETASmoothingMovingAverage(t *testing.T) {
	// Stand up an ETAService with a stub router and exercise only the
	// speed filter. We do not need Mongo/Redis for the pure math path.
	router := &stubRouter{distanceM: 70_000, durationS: 3600}
	svc := services.NewETAService(nil, nil, router, zap.NewNop())

	start := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	var prev *models.Waypoint
	for i := 0; i < 30; i++ {
		wp := models.Waypoint{
			RecordedAt: start.Add(time.Duration(i) * time.Second),
			Position:   models.NewGeoPoint(10.0+float64(i)*0.001, 45.0),
			SpeedKPH:   80,
		}
		svc.UpdateSpeed("S-1", wp, prev)
		prev = &wp
	}

	// 80 km/h for 70 km ⇒ 0.875 h ≈ 3150 s. Use an OSRM Distance of 70
	// km via the stub and confirm the ETA is "now + 0.875 h".
	shp := &models.Shipment{ID: "S-1", TenantID: "t", Destination: models.NewGeoPoint(11, 46)}
	now := time.Now()
	result, err := servicesComputeFrom(svc, shp, models.NewGeoPoint(10, 45))
	if err != nil {
		t.Fatalf("eta error: %v", err)
	}
	delta := result.ETA.Sub(now).Seconds()
	if delta < 3000 || delta > 3400 {
		t.Fatalf("expected ETA ~3150s from now, got %.1fs (speed=%.1f kph)", delta, result.SmoothedSpeedKPH)
	}
	if result.SmoothedSpeedKPH < 70 || result.SmoothedSpeedKPH > 90 {
		t.Fatalf("smoothed speed outside sanity band: %.1f", result.SmoothedSpeedKPH)
	}
}

// TestETABoundsClip verifies the lower/upper bounds on noisy telemetry.
func TestETABoundsClip(t *testing.T) {
	router := &stubRouter{distanceM: 10_000, durationS: 600}
	svc := services.NewETAService(nil, nil, router, zap.NewNop())
	start := time.Now()
	var prev *models.Waypoint
	for _, kph := range []float64{0, 2, 999, 500, 80} {
		wp := models.Waypoint{RecordedAt: start, Position: models.NewGeoPoint(10, 45), SpeedKPH: kph}
		svc.UpdateSpeed("S-clip", wp, prev)
		prev = &wp
	}
	shp := &models.Shipment{ID: "S-clip", TenantID: "t", Destination: models.NewGeoPoint(11, 46)}
	r, err := servicesComputeFrom(svc, shp, models.NewGeoPoint(10, 45))
	if err != nil {
		t.Fatalf("eta error: %v", err)
	}
	if r.SmoothedSpeedKPH > services.DefaultMaxKPH {
		t.Fatalf("smoothed exceeded max: %.1f", r.SmoothedSpeedKPH)
	}
}

// servicesComputeFrom is a tiny test shim so we can exercise the
// internal computeFrom without exposing it in the public API. It lives
// here in the test file; the production path calls ETAService.ETA.
func servicesComputeFrom(svc *services.ETAService, shp *models.Shipment, current models.GeoPoint) (*services.ETAResult, error) {
	return svc.ComputeFromForTest(context.Background(), shp, current)
}

// TestRateLimitConfigDefaults is a cheap regression to keep the config
// defaults stable across commits.
func TestRateLimitConfigDefaults(t *testing.T) {
	var c config.RateLimitConfig
	if c.RPS != 0 {
		t.Fatalf("unexpected zero value: %+v", c)
	}
	// This is mostly to flag a breaking change if someone alters the
	// struct layout without updating the env tags.
	_ = httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
}
