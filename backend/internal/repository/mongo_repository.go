// Package repository isolates data-store concerns from the service
// layer. Each repository exposes a small, intention-revealing surface
// so switching the underlying engine (for example from MongoDB to
// CockroachDB JSONB) is possible without rewriting callers.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/modules/logistics"
)

// ErrNotFound is returned when a document cannot be located.
var ErrNotFound = errors.New("repository: not found")

// Collections names — single source of truth, referenced by services,
// indexers and migrations. Avoid string-literal duplication elsewhere.
const (
	CollectionShipments      = "shipments"
	CollectionTrackingEvents = "tracking_events"
	CollectionVehicles       = "vehicles"
	CollectionDrivers        = "drivers"
	CollectionGeofences      = "geofences"
	CollectionCustody        = "chain_of_custody"
	CollectionAuditLog       = "audit_log"
)

// MongoRepository wraps the Mongo client and exposes per-collection
// accessors. The repository is safe for concurrent use.
type MongoRepository struct {
	client *mongo.Client
	db     *mongo.Database
	log    *zap.Logger
}

// NewMongo constructs the repository and performs a connect + ping,
// returning a typed error if the database is unreachable.
func NewMongo(ctx context.Context, cfg config.MongoConfig, log *zap.Logger) (*MongoRepository, error) {
	opts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetAppName("logitrack-backend")

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	return &MongoRepository{
		client: client,
		db:     client.Database(cfg.Database),
		log:    log,
	}, nil
}

// Client returns the underlying client for health checks.
func (r *MongoRepository) Client() *mongo.Client { return r.client }

// DatabaseName returns the logical Mongo database name (so packages
// that need a distinct collection handle — for example the audit
// writer — don't have to plumb MongoConfig through).
func (r *MongoRepository) DatabaseName() string { return r.db.Name() }

// Disconnect closes the client connections.
func (r *MongoRepository) Disconnect(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

// EnsureIndexes creates the indexes required for performant queries.
// Idempotent — safe to invoke on every boot.
func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	shipmentIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "reference", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_reference_uniq")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("tenant_status_updated")},
		{Keys: bson.D{{Key: "carrier", Value: 1}, {Key: "etd", Value: 1}},
			Options: options.Index().SetName("carrier_etd")},
		{Keys: bson.D{{Key: "current_position", Value: "2dsphere"}},
			Options: options.Index().SetName("current_position_2dsphere")},
	}
	if _, err := r.db.Collection(CollectionShipments).Indexes().CreateMany(ctx, shipmentIdx); err != nil {
		return fmt.Errorf("shipments indexes: %w", err)
	}
	geoIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "polygon", Value: "2dsphere"}},
			Options: options.Index().SetName("polygon_2dsphere")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "type", Value: 1}},
			Options: options.Index().SetName("tenant_type")},
	}
	if _, err := r.db.Collection(CollectionGeofences).Indexes().CreateMany(ctx, geoIdx); err != nil {
		return fmt.Errorf("geofences indexes: %w", err)
	}
	custodyIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "shipment_id", Value: 1}, {Key: "sequence", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("shipment_sequence_uniq")},
	}
	if _, err := r.db.Collection(CollectionCustody).Indexes().CreateMany(ctx, custodyIdx); err != nil {
		return fmt.Errorf("custody indexes: %w", err)
	}
	if err := r.EnsureRifiutiIndexes(ctx); err != nil {
		return fmt.Errorf("rifiuti indexes: %w", err)
	}
	r.log.Info("mongo indexes ensured", zap.String("database", r.db.Name()))
	return nil
}

// --- Shipment accessors -------------------------------------------------

// InsertShipment persists a new shipment document.
func (r *MongoRepository) InsertShipment(ctx context.Context, s *logistics.Shipment) error {
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	_, err := r.db.Collection(CollectionShipments).InsertOne(ctx, s)
	return err
}

// FindShipment retrieves a shipment by id scoped to a tenant.
func (r *MongoRepository) FindShipment(ctx context.Context, tenantID, id string) (*logistics.Shipment, error) {
	var s logistics.Shipment
	filter := bson.M{"_id": id, "tenant_id": tenantID}
	err := r.db.Collection(CollectionShipments).FindOne(ctx, filter).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListShipments returns a page of shipments filtered by the supplied
// predicates. Pagination follows the cursor-free offset model; for
// high-volume tenants the service should migrate to keyset pagination
// on (updated_at, _id).
func (r *MongoRepository) ListShipments(ctx context.Context, f ShipmentFilter) ([]logistics.Shipment, error) {
	filter := bson.M{"tenant_id": f.TenantID}
	if f.Status != "" {
		filter["status"] = f.Status
	}
	if f.Carrier != "" {
		filter["carrier"] = f.Carrier
	}
	if !f.From.IsZero() || !f.To.IsZero() {
		rng := bson.M{}
		if !f.From.IsZero() {
			rng["$gte"] = f.From
		}
		if !f.To.IsZero() {
			rng["$lte"] = f.To
		}
		filter["etd"] = rng
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetLimit(int64(f.Limit)).
		SetSkip(int64(f.Offset))

	cur, err := r.db.Collection(CollectionShipments).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []logistics.Shipment
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AppendWaypoint adds a waypoint to a shipment and refreshes the
// current position in one atomic operation.
func (r *MongoRepository) AppendWaypoint(ctx context.Context, tenantID, shipmentID string, wp logistics.Waypoint) error {
	filter := bson.M{"_id": shipmentID, "tenant_id": tenantID}
	update := bson.M{
		"$push": bson.M{"waypoints": wp},
		"$set": bson.M{
			"current_position": wp.Position,
			"updated_at":       time.Now().UTC(),
		},
	}
	res, err := r.db.Collection(CollectionShipments).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Tracking event accessors ------------------------------------------

// InsertTrackingEvent persists an event. The caller has already
// computed the sequence and hash fields.
func (r *MongoRepository) InsertTrackingEvent(ctx context.Context, e *logistics.TrackingEvent) error {
	_, err := r.db.Collection(CollectionTrackingEvents).InsertOne(ctx, e)
	return err
}

// --- Chain of custody --------------------------------------------------

// AppendCustody persists a custody record. The service layer is
// responsible for sequencing and hash chaining.
func (r *MongoRepository) AppendCustody(ctx context.Context, rec *logistics.CustodyRecord) error {
	_, err := r.db.Collection(CollectionCustody).InsertOne(ctx, rec)
	return err
}

// ListCustody returns the append-only log for a shipment, ordered by
// sequence.
func (r *MongoRepository) ListCustody(ctx context.Context, tenantID, shipmentID string) ([]logistics.CustodyRecord, error) {
	filter := bson.M{"tenant_id": tenantID, "shipment_id": shipmentID}
	opts := options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}})
	cur, err := r.db.Collection(CollectionCustody).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []logistics.CustodyRecord
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LatestCustodySequence returns the current tail of the custody log
// for a shipment, or 0 if there are no records yet.
func (r *MongoRepository) LatestCustodySequence(ctx context.Context, tenantID, shipmentID string) (int64, string, error) {
	filter := bson.M{"tenant_id": tenantID, "shipment_id": shipmentID}
	opts := options.FindOne().SetSort(bson.D{{Key: "sequence", Value: -1}})
	var rec logistics.CustodyRecord
	err := r.db.Collection(CollectionCustody).FindOne(ctx, filter, opts).Decode(&rec)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return rec.Sequence, rec.Hash, nil
}

// ShipmentFilter describes the parameters for ListShipments.
type ShipmentFilter struct {
	TenantID string
	Status   string
	Carrier  string
	From     time.Time
	To       time.Time
	Limit    int
	Offset   int
}
