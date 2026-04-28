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

	"github.com/logitrack/backend/internal/modules/logistics"
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
func (s *ShipmentService) CreateShipment(ctx context.Context, shp *logistics.Shipment) error {
	if shp.ID == "" {
		shp.ID = uuid.NewString()
	}
	if err := shp.Validate(); err != nil {
		return err
	}
	if s.routes != nil && shp.RoutePolyline == "" {
		if r, err := s.routes.OptimiseRoute(ctx, RouteRequest{
			Waypoints: []logistics.GeoPoint{shp.Origin, shp.Destination},
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
	genesis := &logistics.CustodyRecord{
		ID:         uuid.NewString(),
		TenantID:   shp.TenantID,
		ShipmentID: shp.ID,
		Sequence:   1,
		Action:     logistics.CustodyCreated,
		OccurredAt: s.clock(),
		RecordedAt: s.clock(),
		Actor:      logistics.CustodyActor{Name: "system", Role: "creator", Organisation: shp.Carrier, VATNumber: shp.Consignor.VATNumber},
		PrevHash:   "",
	}
	hash, err := computeHash(genesis)
	if err != nil {
		s.log.Warn("genesis custody hash failed", zap.String("shipment_id", shp.ID), zap.Error(err))
		return nil
	}
	genesis.Hash = hash
	if err := s.mongo.AppendCustody(ctx, genesis); err != nil {
		s.log.Warn("genesis custody failed", zap.String("shipment_id", shp.ID), zap.Error(err))
	}
	return nil
}

// GetShipment retrieves a shipment by id.
func (s *ShipmentService) GetShipment(ctx context.Context, tenantID, id string) (*logistics.Shipment, error) {
	return s.mongo.FindShipment(ctx, tenantID, id)
}

// ListShipments returns a page of shipments.
func (s *ShipmentService) ListShipments(ctx context.Context, f repository.ShipmentFilter) ([]logistics.Shipment, error) {
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
func (s *ShipmentService) RecordWaypoint(ctx context.Context, tenantID, shipmentID string, wp logistics.Waypoint) error {
	if wp.RecordedAt.IsZero() {
		wp.RecordedAt = s.clock()
	}
	if len(wp.Position.Coordinates) != 2 {
		return logistics.ErrInvalidGeoPoint
	}
	// Capture the previous cached position BEFORE writing the new one.
	var prev *logistics.Waypoint
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
	evt := logistics.TrackingEvent{
		ID:         uuid.NewString(),
		TenantID:   tenantID,
		ShipmentID: shipmentID,
		Type:       logistics.EventPositionUpdate,
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
func (s *ShipmentService) ChainOfCustody(ctx context.Context, tenantID, shipmentID string) ([]logistics.CustodyRecord, error) {
	return s.mongo.ListCustody(ctx, tenantID, shipmentID)
}

// AppendCustody records a new custody action, chaining hashes to the
// previous tail.
func (s *ShipmentService) AppendCustody(ctx context.Context, rec *logistics.CustodyRecord) error {
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
	hash, err := computeHash(rec)
	if err != nil {
		return fmt.Errorf("append custody: %w", err)
	}
	rec.Hash = hash
	return s.mongo.AppendCustody(ctx, rec)
}

// computeHash returns the SHA-256 of the canonical JSON encoding of
// the custody record minus the Hash field itself. The same algorithm
// must be implemented by auditors verifying chain integrity.
//
// Canonicalisation is essential because the record round-trips through
// MongoDB, which stores time.Time as BSON Date — millisecond precision,
// no monotonic-clock semantics, always UTC. If we hashed the
// pre-persistence record (with arbitrary precision and the local
// monotonic clock attached), the read-back hash would never match.
// Every time field is therefore normalised to UTC at millisecond
// precision before marshalling so the in-memory and post-Mongo forms
// produce byte-identical JSON.
func computeHash(rec *logistics.CustodyRecord) (string, error) {
	copyRec := *rec
	copyRec.Hash = ""
	copyRec.OccurredAt = canonicalTime(copyRec.OccurredAt)
	copyRec.RecordedAt = canonicalTime(copyRec.RecordedAt)
	if copyRec.Location != nil {
		// Defensive copy of the location pointer so callers don't see
		// any mutation. Currently no time on Location, but keeping the
		// pattern symmetric simplifies future field additions.
		loc := *copyRec.Location
		copyRec.Location = &loc
	}
	payload, err := json.Marshal(copyRec)
	if err != nil {
		return "", fmt.Errorf("custody hash: marshal sequence %d: %w", rec.Sequence, err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

// canonicalTime is the time-normalisation rule used by computeHash.
// It is exported only as documentation through the package — callers
// must always go through computeHash to keep one source of truth for
// the hash algorithm. Zero values are returned untouched so the JSON
// encoder can emit the canonical Go zero ("0001-01-01T00:00:00Z").
func canonicalTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.UTC().Truncate(time.Millisecond)
}

// VerifyChain recomputes every record's hash and confirms the PrevHash
// linkage walks back to the empty genesis pointer. It returns
// (valid, first-broken-sequence, error). A nil error and valid==true
// means the chain is intact. Exposed for the external auditor tool and
// the integration tests.
func VerifyChain(records []logistics.CustodyRecord) (bool, int64, error) {
	prev := ""
	for i := range records {
		r := records[i]
		if r.PrevHash != prev {
			return false, r.Sequence, fmt.Errorf("sequence %d: prev_hash mismatch", r.Sequence)
		}
		want, err := computeHash(&r)
		if err != nil {
			return false, r.Sequence, fmt.Errorf("sequence %d: %w", r.Sequence, err)
		}
		if r.Hash != want {
			return false, r.Sequence, fmt.Errorf("sequence %d: hash mismatch", r.Sequence)
		}
		prev = r.Hash
	}
	return true, 0, nil
}
