// Package services contains the domain behaviour that orchestrates
// repositories and external integrations. Each service exposes a
// narrow interface the handlers depend on, enabling easy substitution
// in unit tests.
package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/repository"
)

// ShipmentService coordinates persistence and event publication for
// shipment mutations. All public methods are context-cancellable.
//
// The service owns (optional) references to a RouteOptimizer — used to
// cache the planned polyline on first creation — and an ETAService —
// used to feed telemetry into the smoothed-speed filter on every
// ingested waypoint. Both references are nullable so the unit-test
// path does not need to stand up the whole dependency graph.
type ShipmentService struct {
	mongo  *repository.MongoRepository
	redis  *repository.RedisRepository
	routes RouteOptimizer
	eta    *ETAService
	log    *zap.Logger
	clock  func() time.Time
}

// NewShipmentService constructs a ShipmentService.
func NewShipmentService(mongo *repository.MongoRepository, redis *repository.RedisRepository, log *zap.Logger) *ShipmentService {
	return &ShipmentService{mongo: mongo, redis: redis, log: log, clock: func() time.Time { return time.Now().UTC() }}
}

// WithRouting attaches the RouteOptimizer used to pre-compute the
// planned polyline at shipment creation time. Returns the service for
// easy wiring at the composition root.
func (s *ShipmentService) WithRouting(opt RouteOptimizer) *ShipmentService { s.routes = opt; return s }

// WithETA attaches the ETAService so incoming waypoints feed into the
// moving-average smoother.
func (s *ShipmentService) WithETA(e *ETAService) *ShipmentService { s.eta = e; return s }

// CreateShipment validates and persists a new shipment, then writes
// the first entry of the chain-of-custody log. If a RouteOptimizer is
// wired, the planned polyline is pre-computed so the map shows the
// route before the first waypoint arrives.
func (s *ShipmentService) CreateShipment(ctx context.Context, shp *models.Shipment) error {
	if shp.ID == "" {
		shp.ID = uuid.NewString()
	}
	if err := shp.Validate(); err != nil {
		return err
	}
	if s.routes != nil && shp.RoutePolyline == "" {
		if r, err := s.routes.OptimiseRoute(ctx, RouteRequest{
			Waypoints: []models.GeoPoint{shp.Origin, shp.Destination},
			Vehicle:   "truck",
		}); err == nil {
			shp.RoutePolyline = r.Geometry
			if shp.ETA.IsZero() {
				shp.ETA = s.clock().Add(time.Duration(r.Duration * float64(time.Second)))
			}
		}
	}
	if err := s.mongo.InsertShipment(ctx, shp); err != nil {
		return fmt.Errorf("create shipment: %w", err)
	}
	genesis := &models.CustodyRecord{
		ID:         uuid.NewString(),
		TenantID:   shp.TenantID,
		ShipmentID: shp.ID,
		Sequence:   1,
		Action:     models.CustodyCreated,
		OccurredAt: s.clock(),
		RecordedAt: s.clock(),
		Actor:      models.CustodyActor{Name: "system", Role: "creator", Organisation: shp.Carrier, VATNumber: shp.Consignor.VATNumber},
		PrevHash:   "",
	}
	genesis.Hash = computeHash(genesis)
	if err := s.mongo.AppendCustody(ctx, genesis); err != nil {
		s.log.Warn("genesis custody failed", zap.String("shipment_id", shp.ID), zap.Error(err))
	}
	return nil
}

// GetShipment retrieves a shipment by id.
func (s *ShipmentService) GetShipment(ctx context.Context, tenantID, id string) (*models.Shipment, error) {
	return s.mongo.FindShipment(ctx, tenantID, id)
}

// ListShipments returns a page of shipments.
func (s *ShipmentService) ListShipments(ctx context.Context, f repository.ShipmentFilter) ([]models.Shipment, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	return s.mongo.ListShipments(ctx, f)
}

// RecordWaypoint ingests a telematics position update. It is called
// from the HTTP handler and, in the future, from the Kafka consumer
// for direct provider webhooks. If an ETAService is wired, the
// waypoint feeds into the smoothing filter so the next ETA query
// reflects current traffic.
func (s *ShipmentService) RecordWaypoint(ctx context.Context, tenantID, shipmentID string, wp models.Waypoint) error {
	if wp.RecordedAt.IsZero() {
		wp.RecordedAt = s.clock()
	}
	if len(wp.Position.Coordinates) != 2 {
		return models.ErrInvalidGeoPoint
	}
	// Capture the previous cached position BEFORE writing the new one.
	var prev *models.Waypoint
	if cur, err := s.redis.GetLatestPosition(ctx, shipmentID); err == nil {
		prev = cur
	}
	if err := s.mongo.AppendWaypoint(ctx, tenantID, shipmentID, wp); err != nil {
		return err
	}
	if err := s.redis.CacheLatestPosition(ctx, shipmentID, wp); err != nil {
		s.log.Warn("redis cache position failed", zap.String("shipment_id", shipmentID), zap.Error(err))
	}
	if s.eta != nil {
		s.eta.UpdateSpeed(shipmentID, wp, prev)
	}
	evt := models.TrackingEvent{
		ID:         uuid.NewString(),
		TenantID:   tenantID,
		ShipmentID: shipmentID,
		Type:       models.EventPositionUpdate,
		Sequence:   wp.RecordedAt.UnixNano(),
		OccurredAt: wp.RecordedAt,
		RecordedAt: s.clock(),
		Position:   &wp.Position,
		Source:     wp.Source,
	}
	if err := s.mongo.InsertTrackingEvent(ctx, &evt); err != nil {
		s.log.Warn("mongo insert event failed", zap.Error(err))
	}
	if err := s.redis.PublishTrackingEvent(ctx, evt); err != nil {
		s.log.Warn("redis publish failed", zap.Error(err))
	}
	return nil
}

// ChainOfCustody returns the immutable custody log for a shipment.
func (s *ShipmentService) ChainOfCustody(ctx context.Context, tenantID, shipmentID string) ([]models.CustodyRecord, error) {
	return s.mongo.ListCustody(ctx, tenantID, shipmentID)
}

// AppendCustody records a new custody action, chaining hashes to the
// previous tail.
func (s *ShipmentService) AppendCustody(ctx context.Context, rec *models.CustodyRecord) error {
	seq, prev, err := s.mongo.LatestCustodySequence(ctx, rec.TenantID, rec.ShipmentID)
	if err != nil {
		return err
	}
	rec.Sequence = seq + 1
	rec.PrevHash = prev
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	if rec.RecordedAt.IsZero() {
		rec.RecordedAt = s.clock()
	}
	rec.Hash = computeHash(rec)
	return s.mongo.AppendCustody(ctx, rec)
}

// computeHash returns the SHA-256 of the canonical JSON encoding of
// the custody record minus the Hash field itself. The same algorithm
// must be implemented by auditors verifying chain integrity.
func computeHash(rec *models.CustodyRecord) string {
	copyRec := *rec
	copyRec.Hash = ""
	payload, _ := json.Marshal(copyRec)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// VerifyChain recomputes every record's hash and confirms the PrevHash
// linkage walks back to the empty genesis pointer. It returns
// (valid, first-broken-sequence, error). A nil error and valid==true
// means the chain is intact. Exposed for the external auditor tool and
// the integration tests.
func VerifyChain(records []models.CustodyRecord) (bool, int64, error) {
	prev := ""
	for i := range records {
		r := records[i]
		if r.PrevHash != prev {
			return false, r.Sequence, fmt.Errorf("sequence %d: prev_hash mismatch", r.Sequence)
		}
		want := computeHash(&r)
		if r.Hash != want {
			return false, r.Sequence, fmt.Errorf("sequence %d: hash mismatch", r.Sequence)
		}
		prev = r.Hash
	}
	return true, 0, nil
}
