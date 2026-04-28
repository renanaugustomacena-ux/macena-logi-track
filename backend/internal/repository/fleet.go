package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/logitrack/backend/internal/modules/logistics"
)

// --- Vehicles ----------------------------------------------------------

// InsertVehicle persists a new vehicle. CreatedAt/UpdatedAt and the
// document id are stamped here so callers don't have to remember and
// so the response carries the assigned id back to the client.
func (r *MongoRepository) InsertVehicle(ctx context.Context, v *logistics.Vehicle) error {
	if v.ID == "" {
		v.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if v.CreatedAt.IsZero() {
		v.CreatedAt = now
	}
	v.UpdatedAt = now
	_, err := r.db.Collection(CollectionVehicles).InsertOne(ctx, v)
	return err
}

// GetVehicle retrieves a vehicle by id, scoped to a tenant.
func (r *MongoRepository) GetVehicle(ctx context.Context, tenantID, id string) (*logistics.Vehicle, error) {
	var v logistics.Vehicle
	err := r.db.Collection(CollectionVehicles).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&v)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ListVehicles returns vehicles for a tenant (no filters beyond tenant +
// active flag, on purpose — richer queries live in the service layer
// when customer requirements settle).
func (r *MongoRepository) ListVehicles(ctx context.Context, tenantID string, activeOnly bool, limit, offset int) ([]logistics.Vehicle, error) {
	filter := bson.M{"tenant_id": tenantID}
	if activeOnly {
		filter["active"] = true
	}
	opts := options.Find().SetSort(bson.D{{Key: "plate", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.db.Collection(CollectionVehicles).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []logistics.Vehicle
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Drivers -----------------------------------------------------------

// InsertDriver persists a new driver. ID is generated when blank so
// the response carries the assigned identifier back to the caller.
func (r *MongoRepository) InsertDriver(ctx context.Context, d *logistics.Driver) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	_, err := r.db.Collection(CollectionDrivers).InsertOne(ctx, d)
	return err
}

// GetDriver retrieves a driver by id, scoped to a tenant.
func (r *MongoRepository) GetDriver(ctx context.Context, tenantID, id string) (*logistics.Driver, error) {
	var d logistics.Driver
	err := r.db.Collection(CollectionDrivers).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDrivers lists drivers for a tenant.
func (r *MongoRepository) ListDrivers(ctx context.Context, tenantID string, activeOnly bool, limit, offset int) ([]logistics.Driver, error) {
	filter := bson.M{"tenant_id": tenantID}
	if activeOnly {
		filter["active"] = true
	}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.db.Collection(CollectionDrivers).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []logistics.Driver
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Geofences ---------------------------------------------------------

// InsertGeofence persists a new geofence polygon. ID is generated
// when blank so the response carries the assigned identifier back
// to the caller.
func (r *MongoRepository) InsertGeofence(ctx context.Context, g *logistics.Geofence) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	g.UpdatedAt = now
	_, err := r.db.Collection(CollectionGeofences).InsertOne(ctx, g)
	return err
}

// GetGeofence retrieves a geofence by id, scoped to a tenant.
func (r *MongoRepository) GetGeofence(ctx context.Context, tenantID, id string) (*logistics.Geofence, error) {
	var g logistics.Geofence
	err := r.db.Collection(CollectionGeofences).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&g)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// ListGeofences returns geofences for a tenant.
func (r *MongoRepository) ListGeofences(ctx context.Context, tenantID string, activeOnly bool, kind logistics.GeofenceType, limit, offset int) ([]logistics.Geofence, error) {
	filter := bson.M{"tenant_id": tenantID}
	if activeOnly {
		filter["active"] = true
	}
	if kind != "" {
		filter["type"] = kind
	}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.db.Collection(CollectionGeofences).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []logistics.Geofence
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
