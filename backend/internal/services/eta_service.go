package services

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/repository"
)

// ETAService computes an estimated-time-of-arrival from a shipment's
// current position, the remaining polyline and a moving-average
// smoothing filter over the last N speed samples.
//
// Design:
//   - Given the shipment origin → destination, the service maintains a
//     "planned polyline" (from the route optimiser, cached).
//   - Every waypoint update is fed into the speed filter. If the
//     waypoint carries a SpeedKPH value we trust it; otherwise we
//     derive speed from delta-time × great-circle distance.
//   - ETA = now + remaining_km / smoothed_speed_kph.
//   - Smoothing uses a simple unweighted moving average over
//     DefaultMASamples entries; one sample is dropped from the head
//     once the window is full (FIFO).
//   - A lower-bound speed (DefaultMinKPH) prevents division by zero
//     when the vehicle is stationary at customs or in heavy traffic.
//   - An upper-bound speed (DefaultMaxKPH) clips telematics noise.
type ETAService struct {
	mongo   *repository.MongoRepository
	redis   *repository.RedisRepository
	routes  RouteOptimizer
	log     *zap.Logger
	mu      sync.Mutex
	filters map[string]*speedFilter // key = shipmentID
	clock   func() time.Time
}

// Default tuning constants. Values are chosen for Italian road freight
// in the Verona/Brennero corridor. A production deployment wires them
// to the tenant configuration.
const (
	DefaultMASamples = 20
	DefaultMinKPH    = 5.0
	DefaultMaxKPH    = 120.0
)

// NewETAService constructs an ETAService.
func NewETAService(mongo *repository.MongoRepository, redis *repository.RedisRepository, routes RouteOptimizer, log *zap.Logger) *ETAService {
	return &ETAService{
		mongo:   mongo,
		redis:   redis,
		routes:  routes,
		log:     log,
		filters: make(map[string]*speedFilter),
		clock:   func() time.Time { return time.Now().UTC() },
	}
}

// ErrShipmentMissingCurrentPosition is returned when ETA cannot be
// computed because no waypoint has been ingested yet.
var ErrShipmentMissingCurrentPosition = errors.New("eta: shipment has no current position")

// ETA returns the predicted arrival time and remaining distance for a
// shipment given its latest known position. The returned distance is
// in kilometres; duration is seconds.
//
// Deterministic in unit tests via clock injection.
func (e *ETAService) ETA(ctx context.Context, tenantID, shipmentID string) (*ETAResult, error) {
	shp, err := e.mongo.FindShipment(ctx, tenantID, shipmentID)
	if err != nil {
		return nil, err
	}
	current := shp.CurrentPosition
	if current == nil {
		return nil, ErrShipmentMissingCurrentPosition
	}
	return e.computeFrom(ctx, shp, *current)
}

// UpdateSpeed feeds a waypoint into the smoothing filter. Call this on
// every inbound telematics event.
func (e *ETAService) UpdateSpeed(shipmentID string, wp models.Waypoint, prev *models.Waypoint) {
	var kph float64
	switch {
	case wp.SpeedKPH > 0:
		kph = wp.SpeedKPH
	case prev != nil:
		dt := wp.RecordedAt.Sub(prev.RecordedAt).Seconds()
		if dt > 0 && len(wp.Position.Coordinates) == 2 && len(prev.Position.Coordinates) == 2 {
			km := haversineKm(
				prev.Position.Coordinates[1], prev.Position.Coordinates[0],
				wp.Position.Coordinates[1], wp.Position.Coordinates[0],
			)
			kph = (km / dt) * 3600
		}
	}
	if kph <= 0 {
		return
	}
	if kph < DefaultMinKPH {
		kph = DefaultMinKPH
	}
	if kph > DefaultMaxKPH {
		kph = DefaultMaxKPH
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	f, ok := e.filters[shipmentID]
	if !ok {
		f = newSpeedFilter(DefaultMASamples)
		e.filters[shipmentID] = f
	}
	f.add(kph)
}

// computeFrom is the pure computation path, useful from tests with a
// known current position.
func (e *ETAService) computeFrom(ctx context.Context, shp *models.Shipment, current models.GeoPoint) (*ETAResult, error) {
	rr, err := e.routes.OptimiseRoute(ctx, RouteRequest{
		Waypoints: []models.GeoPoint{current, shp.Destination},
		Vehicle:   "truck",
	})
	if err != nil {
		return nil, err
	}
	smoothed := e.smoothedSpeedKPH(shp.ID)
	if smoothed <= 0 {
		// No telematics samples yet — trust the OSRM duration.
		eta := e.clock().Add(time.Duration(rr.Duration * float64(time.Second)))
		return &ETAResult{
			ShipmentID:       shp.ID,
			RemainingMeters:  rr.Distance,
			SmoothedSpeedKPH: DefaultAverageSpeedKPH,
			ETA:              eta,
			Source:           rr.Source,
			Geometry:         rr.Geometry,
		}, nil
	}
	km := rr.Distance / 1000
	hours := km / smoothed
	etaTime := e.clock().Add(time.Duration(hours * float64(time.Hour)))
	return &ETAResult{
		ShipmentID:       shp.ID,
		RemainingMeters:  rr.Distance,
		SmoothedSpeedKPH: smoothed,
		ETA:              etaTime,
		Source:           rr.Source,
		Geometry:         rr.Geometry,
	}, nil
}

func (e *ETAService) smoothedSpeedKPH(shipmentID string) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	f, ok := e.filters[shipmentID]
	if !ok {
		return 0
	}
	return f.average()
}

// ETAResult is the response envelope returned by ETAService.ETA.
type ETAResult struct {
	ShipmentID       string    `json:"shipmentId"`
	RemainingMeters  float64   `json:"remainingMeters"`
	SmoothedSpeedKPH float64   `json:"smoothedSpeedKph"`
	ETA              time.Time `json:"eta"`
	Source           string    `json:"source"`
	Geometry         string    `json:"geometry,omitempty"`
}

// --- speedFilter -----------------------------------------------------------

type speedFilter struct {
	size   int
	values []float64
	sum    float64
}

func newSpeedFilter(size int) *speedFilter {
	return &speedFilter{size: size, values: make([]float64, 0, size)}
}

func (f *speedFilter) add(kph float64) {
	if len(f.values) >= f.size {
		f.sum -= f.values[0]
		f.values = f.values[1:]
	}
	f.values = append(f.values, kph)
	f.sum += kph
}

func (f *speedFilter) average() float64 {
	if len(f.values) == 0 {
		return 0
	}
	return f.sum / float64(len(f.values))
}

// Forces the package to import math so staticcheck stays green if we
// revisit the file for bounded-exponential smoothing later.
var _ = math.Pi
