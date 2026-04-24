package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/logitrack/backend/internal/models"
)

// --- Vehicles ----------------------------------------------------------

// InsertVehicle persists a new vehicle. CreatedAt/UpdatedAt are set here
// so callers don't have to remember.
func (r *MongoRepository) InsertVehicle(ctx context.Context, v *models.Vehicle) error {
	now := time.Now().UTC()
	if v.CreatedAt.IsZero() {
		v.CreatedAt = now
	}
	v.UpdatedAt = now
	_, err := r.db.Collection(CollectionVehicles).InsertOne(ctx, v)
	return err
}

// GetVehicle retrieves a vehicle by id, scoped to a tenant.
func (r *MongoRepository) GetVehicle(ctx context.Context, tenantID, id string) (*models.Vehicle, error) {
	var v models.Vehicle
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
func (r *MongoRepository) ListVehicles(ctx context.Context, tenantID string, activeOnly bool, limit, offset int) ([]models.Vehicle, error) {
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
	var out []models.Vehicle
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Drivers -----------------------------------------------------------

// InsertDriver persists a new driver.
func (r *MongoRepository) InsertDriver(ctx context.Context, d *models.Driver) error {
	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	_, err := r.db.Collection(CollectionDrivers).InsertOne(ctx, d)
	return err
}

// GetDriver retrieves a driver by id, scoped to a tenant.
func (r *MongoRepository) GetDriver(ctx context.Context, tenantID, id string) (*models.Driver, error) {
	var d models.Driver
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
func (r *MongoRepository) ListDrivers(ctx context.Context, tenantID string, activeOnly bool, limit, offset int) ([]models.Driver, error) {
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
	var out []models.Driver
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Geofences ---------------------------------------------------------

// InsertGeofence persists a new geofence polygon.
func (r *MongoRepository) InsertGeofence(ctx context.Context, g *models.Geofence) error {
	now := time.Now().UTC()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	g.UpdatedAt = now
	_, err := r.db.Collection(CollectionGeofences).InsertOne(ctx, g)
	return err
}

// GetGeofence retrieves a geofence by id, scoped to a tenant.
func (r *MongoRepository) GetGeofence(ctx context.Context, tenantID, id string) (*models.Geofence, error) {
	var g models.Geofence
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
func (r *MongoRepository) ListGeofences(ctx context.Context, tenantID string, activeOnly bool, kind models.GeofenceType, limit, offset int) ([]models.Geofence, error) {
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
	var out []models.Geofence
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
