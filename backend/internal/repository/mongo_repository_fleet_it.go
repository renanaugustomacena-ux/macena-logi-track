package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/logitrack/backend/internal/modules/fleet_it"
)

const (
	CollectionDriverDevices   = "fleet_it_devices"
	CollectionTelematicsUnits = "fleet_it_telematics"
)

// EnsureFleetITIndexes creates indexes for the fleet-IT collections.
func (r *MongoRepository) EnsureFleetITIndexes(ctx context.Context) error {
	deviceIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "driver_id", Value: 1}},
			Options: options.Index().SetName("tenant_driver")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "serial_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_device_serial_uniq")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("tenant_device_status")},
	}
	if _, err := r.db.Collection(CollectionDriverDevices).Indexes().CreateMany(ctx, deviceIdx); err != nil {
		return err
	}

	telIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "vehicle_id", Value: 1}},
			Options: options.Index().SetName("tenant_vehicle")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "device_serial", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_tel_serial_uniq")},
	}
	if _, err := r.db.Collection(CollectionTelematicsUnits).Indexes().CreateMany(ctx, telIdx); err != nil {
		return err
	}
	return nil
}

// --- Driver Devices ------------------------------------------------------

func (r *MongoRepository) InsertDriverDevice(ctx context.Context, d *fleet_it.DriverDevice) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	_, err := r.db.Collection(CollectionDriverDevices).InsertOne(ctx, d)
	return err
}

func (r *MongoRepository) GetDriverDevice(ctx context.Context, tenantID, id string) (*fleet_it.DriverDevice, error) {
	var d fleet_it.DriverDevice
	err := r.db.Collection(CollectionDriverDevices).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *MongoRepository) ListDriverDevices(ctx context.Context, tenantID, driverID, status string, limit, offset int) ([]fleet_it.DriverDevice, error) {
	filter := bson.M{"tenant_id": tenantID}
	if driverID != "" {
		filter["driver_id"] = driverID
	}
	if status != "" {
		filter["status"] = status
	}
	opts := options.Find().SetSort(bson.D{{Key: "serial_number", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.db.Collection(CollectionDriverDevices).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []fleet_it.DriverDevice
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MongoRepository) UpdateDriverDevice(ctx context.Context, d *fleet_it.DriverDevice) error {
	d.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": d.ID, "tenant_id": d.TenantID}
	res, err := r.db.Collection(CollectionDriverDevices).ReplaceOne(ctx, filter, d)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Telematics Units ----------------------------------------------------

func (r *MongoRepository) InsertTelematicsUnit(ctx context.Context, t *fleet_it.TelematicsUnit) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	_, err := r.db.Collection(CollectionTelematicsUnits).InsertOne(ctx, t)
	return err
}

func (r *MongoRepository) GetTelematicsUnit(ctx context.Context, tenantID, id string) (*fleet_it.TelematicsUnit, error) {
	var t fleet_it.TelematicsUnit
	err := r.db.Collection(CollectionTelematicsUnits).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MongoRepository) ListTelematicsUnits(ctx context.Context, tenantID, vehicleID string, limit, offset int) ([]fleet_it.TelematicsUnit, error) {
	filter := bson.M{"tenant_id": tenantID}
	if vehicleID != "" {
		filter["vehicle_id"] = vehicleID
	}
	opts := options.Find().SetSort(bson.D{{Key: "device_serial", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.db.Collection(CollectionTelematicsUnits).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []fleet_it.TelematicsUnit
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MongoRepository) UpdateTelematicsUnit(ctx context.Context, t *fleet_it.TelematicsUnit) error {
	t.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": t.ID, "tenant_id": t.TenantID}
	res, err := r.db.Collection(CollectionTelematicsUnits).ReplaceOne(ctx, filter, t)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
