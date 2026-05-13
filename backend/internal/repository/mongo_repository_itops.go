package repository

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/logitrack/backend/internal/modules/itops"
)

const maxSearchLen = 200

const (
	CollectionITAssets     = "it_assets"
	CollectionIncidents   = "it_incidents"
	CollectionMonChecks   = "it_monitoring_checks"
	CollectionMonAlerts   = "it_monitoring_alerts"
	CollectionLicenses    = "it_licenses"
	CollectionRunbooks    = "it_runbooks"
)

// EnsureITOpsIndexes creates indexes for the itops collections.
func (r *MongoRepository) EnsureITOpsIndexes(ctx context.Context) error {
	assetIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "kind", Value: 1}},
			Options: options.Index().SetName("tenant_kind")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "serial_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_serial_uniq")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("tenant_status")},
	}
	if _, err := r.db.Collection(CollectionITAssets).Indexes().CreateMany(ctx, assetIdx); err != nil {
		return err
	}

	incidentIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("tenant_inc_status")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "reference", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_inc_ref_uniq")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "opened_at", Value: -1}},
			Options: options.Index().SetName("tenant_inc_opened")},
	}
	if _, err := r.db.Collection(CollectionIncidents).Indexes().CreateMany(ctx, incidentIdx); err != nil {
		return err
	}

	alertIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}, {Key: "severity", Value: 1}},
			Options: options.Index().SetName("tenant_alert_status_sev")},
	}
	if _, err := r.db.Collection(CollectionMonAlerts).Indexes().CreateMany(ctx, alertIdx); err != nil {
		return err
	}

	checkIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "asset_id", Value: 1}, {Key: "check_name", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_asset_check_uniq")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("tenant_check_status")},
	}
	if _, err := r.db.Collection(CollectionMonChecks).Indexes().CreateMany(ctx, checkIdx); err != nil {
		return err
	}

	licenseIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "software", Value: 1}},
			Options: options.Index().SetName("tenant_license_software")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "expiry_date", Value: 1}},
			Options: options.Index().SetName("tenant_license_expiry")},
	}
	if _, err := r.db.Collection(CollectionLicenses).Indexes().CreateMany(ctx, licenseIdx); err != nil {
		return err
	}

	runbookIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "category", Value: 1}},
			Options: options.Index().SetName("tenant_runbook_category")},
	}
	if _, err := r.db.Collection(CollectionRunbooks).Indexes().CreateMany(ctx, runbookIdx); err != nil {
		return err
	}
	return nil
}

// --- IT Assets -----------------------------------------------------------

func (r *MongoRepository) InsertITAsset(ctx context.Context, a *itops.ITAsset) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	_, err := r.db.Collection(CollectionITAssets).InsertOne(ctx, a)
	return err
}

func (r *MongoRepository) GetITAsset(ctx context.Context, tenantID, id string) (*itops.ITAsset, error) {
	var a itops.ITAsset
	err := r.db.Collection(CollectionITAssets).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// AssetFilter describes query parameters for ListITAssets.
type AssetFilter struct {
	TenantID string
	Kind     string
	Status   string
	Search   string
	Limit    int
	Offset   int
}

func (r *MongoRepository) ListITAssets(ctx context.Context, f AssetFilter) ([]itops.ITAsset, int64, error) {
	filter := bson.M{"tenant_id": f.TenantID}
	if f.Kind != "" {
		filter["kind"] = f.Kind
	}
	if f.Status != "" {
		filter["status"] = f.Status
	}
	if f.Search != "" {
		s := f.Search
		if len(s) > maxSearchLen {
			s = s[:maxSearchLen]
		}
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(s), "$options": "i"}
	}

	total, err := r.db.Collection(CollectionITAssets).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	if f.Limit > 0 {
		opts.SetLimit(int64(f.Limit))
	}
	if f.Offset > 0 {
		opts.SetSkip(int64(f.Offset))
	}
	cur, err := r.db.Collection(CollectionITAssets).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.ITAsset
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *MongoRepository) UpdateITAsset(ctx context.Context, a *itops.ITAsset) error {
	a.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": a.ID, "tenant_id": a.TenantID}
	res, err := r.db.Collection(CollectionITAssets).ReplaceOne(ctx, filter, a)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CountITAssetsByKind returns per-kind counts for a tenant.
func (r *MongoRepository) CountITAssetsByKind(ctx context.Context, tenantID string) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"tenant_id": tenantID}}},
		{{Key: "$group", Value: bson.M{"_id": "$kind", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.db.Collection(CollectionITAssets).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	result := make(map[string]int64)
	for cur.Next(ctx) {
		var row struct {
			Kind  string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Kind] = row.Count
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Incidents -----------------------------------------------------------

func (r *MongoRepository) InsertIncident(ctx context.Context, inc *itops.Incident) error {
	if inc.ID == "" {
		inc.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if inc.CreatedAt.IsZero() {
		inc.CreatedAt = now
	}
	inc.UpdatedAt = now
	_, err := r.db.Collection(CollectionIncidents).InsertOne(ctx, inc)
	return err
}

func (r *MongoRepository) GetIncident(ctx context.Context, tenantID, id string) (*itops.Incident, error) {
	var inc itops.Incident
	err := r.db.Collection(CollectionIncidents).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&inc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

// IncidentFilter describes query parameters for ListIncidents.
type IncidentFilter struct {
	TenantID string
	Status   string
	Priority string
	Category string
	Limit    int
	Offset   int
}

func (r *MongoRepository) ListIncidents(ctx context.Context, f IncidentFilter) ([]itops.Incident, int64, error) {
	filter := bson.M{"tenant_id": f.TenantID}
	if f.Status != "" {
		filter["status"] = f.Status
	}
	if f.Priority != "" {
		filter["priority"] = f.Priority
	}
	if f.Category != "" {
		filter["category"] = f.Category
	}

	total, err := r.db.Collection(CollectionIncidents).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "opened_at", Value: -1}})
	if f.Limit > 0 {
		opts.SetLimit(int64(f.Limit))
	}
	if f.Offset > 0 {
		opts.SetSkip(int64(f.Offset))
	}
	cur, err := r.db.Collection(CollectionIncidents).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.Incident
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *MongoRepository) UpdateIncident(ctx context.Context, inc *itops.Incident) error {
	inc.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": inc.ID, "tenant_id": inc.TenantID}
	res, err := r.db.Collection(CollectionIncidents).ReplaceOne(ctx, filter, inc)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CountIncidentsByStatus returns per-status counts for a tenant.
func (r *MongoRepository) CountIncidentsByStatus(ctx context.Context, tenantID string) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"tenant_id": tenantID}}},
		{{Key: "$group", Value: bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.db.Collection(CollectionIncidents).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	result := make(map[string]int64)
	for cur.Next(ctx) {
		var row struct {
			Status string `bson:"_id"`
			Count  int64  `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Status] = row.Count
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// CountIncidentsByPriority returns per-priority counts for open incidents.
func (r *MongoRepository) CountIncidentsByPriority(ctx context.Context, tenantID string) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"tenant_id": tenantID,
			"status":    bson.M{"$nin": bson.A{"resolved", "closed"}},
		}}},
		{{Key: "$group", Value: bson.M{"_id": "$priority", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.db.Collection(CollectionIncidents).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	result := make(map[string]int64)
	for cur.Next(ctx) {
		var row struct {
			Priority string `bson:"_id"`
			Count    int64  `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Priority] = row.Count
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Monitoring Alerts ---------------------------------------------------

func (r *MongoRepository) InsertMonitoringAlert(ctx context.Context, a *itops.MonitoringAlert) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	_, err := r.db.Collection(CollectionMonAlerts).InsertOne(ctx, a)
	return err
}

func (r *MongoRepository) ListMonitoringAlerts(ctx context.Context, tenantID, status string, limit, offset int) ([]itops.MonitoringAlert, error) {
	filter := bson.M{"tenant_id": tenantID}
	if status != "" {
		filter["status"] = status
	}
	opts := options.Find().SetSort(bson.D{{Key: "occurred_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.db.Collection(CollectionMonAlerts).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.MonitoringAlert
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MongoRepository) GetMonitoringAlert(ctx context.Context, tenantID, id string) (*itops.MonitoringAlert, error) {
	var a itops.MonitoringAlert
	err := r.db.Collection(CollectionMonAlerts).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *MongoRepository) UpdateMonitoringAlert(ctx context.Context, a *itops.MonitoringAlert) error {
	filter := bson.M{"_id": a.ID, "tenant_id": a.TenantID}
	res, err := r.db.Collection(CollectionMonAlerts).ReplaceOne(ctx, filter, a)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CountAlertsByStatus returns per-status counts for a tenant.
func (r *MongoRepository) CountAlertsByStatus(ctx context.Context, tenantID string) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"tenant_id": tenantID}}},
		{{Key: "$group", Value: bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.db.Collection(CollectionMonAlerts).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	result := make(map[string]int64)
	for cur.Next(ctx) {
		var row struct {
			Status string `bson:"_id"`
			Count  int64  `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Status] = row.Count
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// ListITAssetsWithWarrantyExpiry returns active assets where warranty_expiry
// is set, sorted by expiry date ascending (soonest first).
func (r *MongoRepository) ListITAssetsWithWarrantyExpiry(ctx context.Context, tenantID string) ([]itops.ITAsset, error) {
	filter := bson.M{
		"tenant_id":      tenantID,
		"status":         bson.M{"$ne": "decommissioned"},
		"warranty_expiry": bson.M{"$ne": nil},
	}
	opts := options.Find().SetSort(bson.D{{Key: "warranty_expiry", Value: 1}})
	cur, err := r.db.Collection(CollectionITAssets).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.ITAsset
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Monitoring Checks ---------------------------------------------------

func (r *MongoRepository) UpsertMonitoringCheck(ctx context.Context, c *itops.MonitoringCheck) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	filter := bson.M{"tenant_id": c.TenantID, "asset_id": c.AssetID, "check_name": c.CheckName}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":        c.ID,
			"tenant_id":  c.TenantID,
			"asset_id":   c.AssetID,
			"check_name": c.CheckName,
			"created_at": now,
		},
		"$set": bson.M{
			"check_type":   c.CheckType,
			"status":       c.Status,
			"output":       c.Output,
			"metric":       c.Metric,
			"metric_unit":  c.MetricUnit,
			"threshold":    c.Threshold,
			"last_checked": c.LastChecked,
			"next_check":   c.NextCheck,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(CollectionMonChecks).UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoRepository) ListMonitoringChecks(ctx context.Context, tenantID string, assetID string, status string) ([]itops.MonitoringCheck, error) {
	filter := bson.M{"tenant_id": tenantID}
	if assetID != "" {
		filter["asset_id"] = assetID
	}
	if status != "" {
		filter["status"] = status
	}
	opts := options.Find().SetSort(bson.D{{Key: "check_name", Value: 1}})
	cur, err := r.db.Collection(CollectionMonChecks).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.MonitoringCheck
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Licenses ------------------------------------------------------------

func (r *MongoRepository) InsertLicense(ctx context.Context, l *itops.License) error {
	if l.ID == "" {
		l.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if l.CreatedAt.IsZero() {
		l.CreatedAt = now
	}
	l.UpdatedAt = now
	_, err := r.db.Collection(CollectionLicenses).InsertOne(ctx, l)
	return err
}

func (r *MongoRepository) GetLicense(ctx context.Context, tenantID, id string) (*itops.License, error) {
	var l itops.License
	err := r.db.Collection(CollectionLicenses).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&l)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// LicenseFilter describes query parameters for ListLicenses.
type LicenseFilter struct {
	TenantID string
	Limit    int
	Offset   int
}

func (r *MongoRepository) ListLicenses(ctx context.Context, f LicenseFilter) ([]itops.License, int64, error) {
	filter := bson.M{"tenant_id": f.TenantID}
	total, err := r.db.Collection(CollectionLicenses).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(bson.D{{Key: "software", Value: 1}})
	if f.Limit > 0 {
		opts.SetLimit(int64(f.Limit))
	}
	if f.Offset > 0 {
		opts.SetSkip(int64(f.Offset))
	}
	cur, err := r.db.Collection(CollectionLicenses).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.License
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *MongoRepository) UpdateLicense(ctx context.Context, l *itops.License) error {
	l.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": l.ID, "tenant_id": l.TenantID}
	res, err := r.db.Collection(CollectionLicenses).ReplaceOne(ctx, filter, l)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// ListLicensesWithExpiry returns licenses where expiry_date is set,
// sorted by expiry ascending. Used by the scadenze aggregator.
func (r *MongoRepository) ListLicensesWithExpiry(ctx context.Context, tenantID string) ([]itops.License, error) {
	filter := bson.M{
		"tenant_id":   tenantID,
		"expiry_date": bson.M{"$ne": nil},
	}
	opts := options.Find().SetSort(bson.D{{Key: "expiry_date", Value: 1}})
	cur, err := r.db.Collection(CollectionLicenses).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.License
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LicenseComplianceSummary returns compliance counts.
func (r *MongoRepository) LicenseComplianceSummary(ctx context.Context, tenantID string) (total, overDeployed, expiringSoon int64, err error) {
	filter := bson.M{"tenant_id": tenantID}
	total, err = r.db.Collection(CollectionLicenses).CountDocuments(ctx, filter)
	if err != nil {
		return
	}
	overFilter := bson.M{
		"tenant_id": tenantID,
		"seats":     bson.M{"$gt": 0},
		"$expr":     bson.M{"$gt": bson.A{"$seats_used", "$seats"}},
	}
	overDeployed, err = r.db.Collection(CollectionLicenses).CountDocuments(ctx, overFilter)
	if err != nil {
		return
	}
	thirtyDays := time.Now().UTC().AddDate(0, 0, 30)
	expiringFilter := bson.M{
		"tenant_id":   tenantID,
		"expiry_date": bson.M{"$ne": nil, "$lte": thirtyDays},
	}
	expiringSoon, err = r.db.Collection(CollectionLicenses).CountDocuments(ctx, expiringFilter)
	return
}

// --- Runbooks ------------------------------------------------------------

func (r *MongoRepository) InsertRunbook(ctx context.Context, rb *itops.Runbook) error {
	if rb.ID == "" {
		rb.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if rb.CreatedAt.IsZero() {
		rb.CreatedAt = now
	}
	rb.UpdatedAt = now
	if rb.Version == 0 {
		rb.Version = 1
	}
	_, err := r.db.Collection(CollectionRunbooks).InsertOne(ctx, rb)
	return err
}

func (r *MongoRepository) GetRunbook(ctx context.Context, tenantID, id string) (*itops.Runbook, error) {
	var rb itops.Runbook
	err := r.db.Collection(CollectionRunbooks).FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&rb)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

// RunbookFilter describes query parameters for ListRunbooks.
type RunbookFilter struct {
	TenantID string
	Category string
	Search   string
	Limit    int
	Offset   int
}

func (r *MongoRepository) ListRunbooks(ctx context.Context, f RunbookFilter) ([]itops.Runbook, int64, error) {
	filter := bson.M{"tenant_id": f.TenantID}
	if f.Category != "" {
		filter["category"] = f.Category
	}
	if f.Search != "" {
		s := f.Search
		if len(s) > maxSearchLen {
			s = s[:maxSearchLen]
		}
		filter["title"] = bson.M{"$regex": regexp.QuoteMeta(s), "$options": "i"}
	}
	total, err := r.db.Collection(CollectionRunbooks).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(bson.D{{Key: "title", Value: 1}})
	if f.Limit > 0 {
		opts.SetLimit(int64(f.Limit))
	}
	if f.Offset > 0 {
		opts.SetSkip(int64(f.Offset))
	}
	cur, err := r.db.Collection(CollectionRunbooks).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []itops.Runbook
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *MongoRepository) UpdateRunbook(ctx context.Context, rb *itops.Runbook) error {
	rb.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": rb.ID, "tenant_id": rb.TenantID}
	res, err := r.db.Collection(CollectionRunbooks).ReplaceOne(ctx, filter, rb)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
