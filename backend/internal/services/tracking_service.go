package services

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/repository"
)

// TrackingService exposes read-side tracking queries and a very small
// analytical helper (distance-to-destination, delay detection) used by
// the dashboards and WebSocket broadcast loop.
type TrackingService struct {
	mongo *repository.MongoRepository
	redis *repository.RedisRepository
	log   *zap.Logger
}

// NewTrackingService constructs a TrackingService.
func NewTrackingService(mongo *repository.MongoRepository, redis *repository.RedisRepository, log *zap.Logger) *TrackingService {
	return &TrackingService{mongo: mongo, redis: redis, log: log}
}

// LatestPosition returns the last known position for a shipment.
// The Redis cache is the fast path; Mongo is consulted on cache miss.
func (t *TrackingService) LatestPosition(ctx context.Context, tenantID, shipmentID string) (*models.Waypoint, error) {
	if wp, err := t.redis.GetLatestPosition(ctx, shipmentID); err == nil {
		return wp, nil
	}
	shp, err := t.mongo.FindShipment(ctx, tenantID, shipmentID)
	if err != nil {
		return nil, err
	}
	if shp.CurrentPosition == nil {
		return nil, repository.ErrNotFound
	}
	wp := models.Waypoint{
		RecordedAt: shp.UpdatedAt,
		Position:   *shp.CurrentPosition,
		Source:     "mongo_fallback",
	}
	return &wp, nil
}

// DistanceRemainingKM returns the great-circle distance between the
// cached position and the destination. Good enough for UI summaries;
// the driving distance is resolved via OSRM where accuracy matters.
func (t *TrackingService) DistanceRemainingKM(current, destination models.GeoPoint) float64 {
	if len(current.Coordinates) != 2 || len(destination.Coordinates) != 2 {
		return math.NaN()
	}
	return haversineKm(current.Coordinates[1], current.Coordinates[0], destination.Coordinates[1], destination.Coordinates[0])
}

// IsDelayed reports whether a shipment is more than `threshold` past
// its planned ETA with no terminal event.
func (t *TrackingService) IsDelayed(s models.Shipment, threshold time.Duration, now time.Time) bool {
	if s.Status == models.StatusDelivered || s.Status == models.StatusCancelled {
		return false
	}
	if s.ETA.IsZero() {
		return false
	}
	return now.After(s.ETA.Add(threshold))
}

// haversineKm computes the great-circle distance between two points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}
