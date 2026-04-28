package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/logitrack/backend/internal/modules/rifiuti"
)

// Collections specific to the rifiuti module. Names are namespaced
// with the "rifiuti_" prefix so the platform's repository singleton
// can serve every module without collisions.
const (
	CollectionRifiutiProduttori    = "rifiuti_produttori"
	CollectionRifiutiTrasportatori = "rifiuti_trasportatori"
	CollectionRifiutiDestinatari   = "rifiuti_destinatari"
	CollectionRifiutiFIR           = "rifiuti_fir"
	CollectionRifiutiRegistro      = "rifiuti_registro"
)

// EnsureRifiutiIndexes creates the indexes required by the rifiuti
// collections. Called from EnsureIndexes after the platform-grade
// indexes are in place. Kept in a separate method so a future
// repository split can move it to a per-module repository without
// touching the cross-cutting one.
func (r *MongoRepository) EnsureRifiutiIndexes(ctx context.Context) error {
	produttoriIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "codice_fiscale", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_cf_uniq")},
	}
	if _, err := r.db.Collection(CollectionRifiutiProduttori).Indexes().CreateMany(ctx, produttoriIdx); err != nil {
		return err
	}
	trasportatoriIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "codice_fiscale", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_cf_uniq")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "albo_scadenza", Value: 1}},
			Options: options.Index().SetName("tenant_albo_scadenza")},
	}
	if _, err := r.db.Collection(CollectionRifiutiTrasportatori).Indexes().CreateMany(ctx, trasportatoriIdx); err != nil {
		return err
	}
	destinatariIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "codice_fiscale", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tenant_cf_uniq")},
	}
	if _, err := r.db.Collection(CollectionRifiutiDestinatari).Indexes().CreateMany(ctx, destinatariIdx); err != nil {
		return err
	}
	firIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "state", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("tenant_state_updated")},
		{Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "numero_rentri", Value: 1}},
			Options: options.Index().SetSparse(true).SetUnique(true).SetName("tenant_numero_rentri_uniq")},
	}
	if _, err := r.db.Collection(CollectionRifiutiFIR).Indexes().CreateMany(ctx, firIdx); err != nil {
		return err
	}
	registroIdx := []mongo.IndexModel{
		{Keys: bson.D{
			{Key: "tenant_id", Value: 1},
			{Key: "operatore_id", Value: 1},
			{Key: "operatore_ruolo", Value: 1},
			{Key: "numero_progressivo", Value: 1},
		}, Options: options.Index().SetUnique(true).SetName("tenant_op_progressivo_uniq")},
	}
	if _, err := r.db.Collection(CollectionRifiutiRegistro).Indexes().CreateMany(ctx, registroIdx); err != nil {
		return err
	}
	return nil
}

// --- Produttore ---------------------------------------------------------

// InsertProduttore persists a new produttore document. Generates a
// UUID id when one is not supplied; sets timestamps.
func (r *MongoRepository) InsertProduttore(ctx context.Context, p *rifiuti.Produttore) error {
	now := time.Now().UTC()
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	_, err := r.db.Collection(CollectionRifiutiProduttori).InsertOne(ctx, p)
	return err
}

// GetProduttore returns the produttore with the supplied id scoped
// to tenant. ErrNotFound on miss.
func (r *MongoRepository) GetProduttore(ctx context.Context, tenantID, id string) (*rifiuti.Produttore, error) {
	var p rifiuti.Produttore
	err := r.db.Collection(CollectionRifiutiProduttori).
		FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListProduttori returns the produttori for the tenant, paginated.
func (r *MongoRepository) ListProduttori(ctx context.Context, tenantID string, limit, offset int) ([]rifiuti.Produttore, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	opts := options.Find().SetSort(bson.D{{Key: "ragione_sociale", Value: 1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cur, err := r.db.Collection(CollectionRifiutiProduttori).Find(ctx, bson.M{"tenant_id": tenantID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []rifiuti.Produttore{}
	for cur.Next(ctx) {
		var p rifiuti.Produttore
		if err := cur.Decode(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, cur.Err()
}

// --- Trasportatore -------------------------------------------------------

// InsertTrasportatore persists a new trasportatore.
func (r *MongoRepository) InsertTrasportatore(ctx context.Context, t *rifiuti.Trasportatore) error {
	now := time.Now().UTC()
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	_, err := r.db.Collection(CollectionRifiutiTrasportatori).InsertOne(ctx, t)
	return err
}

// GetTrasportatore loads a trasportatore by id scoped to tenant.
func (r *MongoRepository) GetTrasportatore(ctx context.Context, tenantID, id string) (*rifiuti.Trasportatore, error) {
	var t rifiuti.Trasportatore
	err := r.db.Collection(CollectionRifiutiTrasportatori).
		FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListTrasportatori returns the trasportatori for the tenant.
func (r *MongoRepository) ListTrasportatori(ctx context.Context, tenantID string, limit, offset int) ([]rifiuti.Trasportatore, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	opts := options.Find().SetSort(bson.D{{Key: "ragione_sociale", Value: 1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cur, err := r.db.Collection(CollectionRifiutiTrasportatori).Find(ctx, bson.M{"tenant_id": tenantID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []rifiuti.Trasportatore{}
	for cur.Next(ctx) {
		var t rifiuti.Trasportatore
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, cur.Err()
}

// --- Destinatario -------------------------------------------------------

// InsertDestinatario persists a new destinatario.
func (r *MongoRepository) InsertDestinatario(ctx context.Context, d *rifiuti.Destinatario) error {
	now := time.Now().UTC()
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	_, err := r.db.Collection(CollectionRifiutiDestinatari).InsertOne(ctx, d)
	return err
}

// GetDestinatario loads a destinatario by id.
func (r *MongoRepository) GetDestinatario(ctx context.Context, tenantID, id string) (*rifiuti.Destinatario, error) {
	var d rifiuti.Destinatario
	err := r.db.Collection(CollectionRifiutiDestinatari).
		FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDestinatari returns the destinatari for the tenant.
func (r *MongoRepository) ListDestinatari(ctx context.Context, tenantID string, limit, offset int) ([]rifiuti.Destinatario, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	opts := options.Find().SetSort(bson.D{{Key: "ragione_sociale", Value: 1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cur, err := r.db.Collection(CollectionRifiutiDestinatari).Find(ctx, bson.M{"tenant_id": tenantID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []rifiuti.Destinatario{}
	for cur.Next(ctx) {
		var d rifiuti.Destinatario
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, cur.Err()
}

// --- FIR ----------------------------------------------------------------

// InsertFIR persists a new FIR document.
func (r *MongoRepository) InsertFIR(ctx context.Context, f *rifiuti.FIR) error {
	now := time.Now().UTC()
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	f.UpdatedAt = now
	_, err := r.db.Collection(CollectionRifiutiFIR).InsertOne(ctx, f)
	return err
}

// GetFIR loads a FIR by id.
func (r *MongoRepository) GetFIR(ctx context.Context, tenantID, id string) (*rifiuti.FIR, error) {
	var f rifiuti.FIR
	err := r.db.Collection(CollectionRifiutiFIR).
		FindOne(ctx, bson.M{"_id": id, "tenant_id": tenantID}).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// ListFIR returns the FIR documents for the tenant, optionally filtered
// by state, sorted by updated_at desc.
func (r *MongoRepository) ListFIR(ctx context.Context, tenantID string, state rifiuti.FIRState, limit, offset int) ([]rifiuti.FIR, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	filter := bson.M{"tenant_id": tenantID}
	if state != "" {
		filter["state"] = state
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cur, err := r.db.Collection(CollectionRifiutiFIR).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []rifiuti.FIR{}
	for cur.Next(ctx) {
		var f rifiuti.FIR
		if err := cur.Decode(&f); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, cur.Err()
}

// UpdateFIR persists an updated FIR document — used for state
// transitions, vidimazione results, and signature timestamps. Refreshes
// updated_at.
func (r *MongoRepository) UpdateFIR(ctx context.Context, f *rifiuti.FIR) error {
	f.UpdatedAt = time.Now().UTC()
	res, err := r.db.Collection(CollectionRifiutiFIR).
		ReplaceOne(ctx, bson.M{"_id": f.ID, "tenant_id": f.TenantID}, f)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
