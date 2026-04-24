// Package audit implements the append-only who/what/when/from/to log
// mandated by v2.0 §12. Every state-changing API call (POST, PUT,
// PATCH, DELETE) should emit at least one record.
//
// Records live in the `audit_log` Mongo collection (single source of
// truth per repository.CollectionAuditLog). Writes are fire-and-forget
// to avoid blocking the request path; the writer runs in a bounded
// goroutine and drops records only under backpressure, in which case a
// counter is incremented and logged at WARN.
package audit

import (
	"context"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/repository"
)

// Record is the audit-log document persisted per action.
type Record struct {
	ID            string                 `bson:"_id,omitempty" json:"id,omitempty"`
	TenantID      string                 `bson:"tenant_id" json:"tenantId"`
	ActorID       string                 `bson:"actor_id" json:"actorId"`
	ActorRoles    []string               `bson:"actor_roles,omitempty" json:"actorRoles,omitempty"`
	Action        string                 `bson:"action" json:"action"`
	TargetKind    string                 `bson:"target_kind" json:"targetKind"`
	TargetID      string                 `bson:"target_id,omitempty" json:"targetId,omitempty"`
	Method        string                 `bson:"method" json:"method"`
	Path          string                 `bson:"path" json:"path"`
	Status        int                    `bson:"status" json:"status"`
	SourceIP      string                 `bson:"source_ip,omitempty" json:"sourceIp,omitempty"`
	CorrelationID string                 `bson:"correlation_id,omitempty" json:"correlationId,omitempty"`
	From          map[string]interface{} `bson:"from,omitempty" json:"from,omitempty"`
	To            map[string]interface{} `bson:"to,omitempty" json:"to,omitempty"`
	At            time.Time              `bson:"at" json:"at"`
}

// Writer is the worker that persists Records. Safe for concurrent use.
type Writer struct {
	coll    *mongo.Collection
	queue   chan Record
	log     *zap.Logger
	dropped atomic.Uint64
}

// NewWriter builds a buffered, lossy writer. `buf` is the channel
// capacity; under backpressure the writer drops records and logs at
// warn. Choose buf so it holds at least one second of peak write load.
func NewWriter(ctx context.Context, repo *repository.MongoRepository, log *zap.Logger, buf int) *Writer {
	if buf <= 0 {
		buf = 1024
	}
	client := repo.Client()
	// Note: we depend on the repository exposing the DB name through a
	// dedicated collection name so we don't have to teach this package
	// about MongoConfig. Call Database via the client and rely on the
	// collection const.
	coll := client.Database(repo.DatabaseName()).Collection(repository.CollectionAuditLog)
	w := &Writer{
		coll:  coll,
		queue: make(chan Record, buf),
		log:   log,
	}
	go w.run(ctx)
	return w
}

// Enqueue submits a record for async persistence. Returns false if the
// queue is full (record dropped).
func (w *Writer) Enqueue(r Record) bool {
	if r.At.IsZero() {
		r.At = time.Now().UTC()
	}
	select {
	case w.queue <- r:
		return true
	default:
		n := w.dropped.Add(1)
		if n%100 == 1 {
			w.log.Warn("audit writer dropped records", zap.Uint64("total_dropped", n))
		}
		return false
	}
}

// Dropped returns the monotonic counter of dropped records.
func (w *Writer) Dropped() uint64 { return w.dropped.Load() }

func (w *Writer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Drain on shutdown with a bounded attempt.
			drainCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			for {
				select {
				case r := <-w.queue:
					_, _ = w.coll.InsertOne(drainCtx, r)
				default:
					cancel()
					return
				}
			}
		case r := <-w.queue:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if _, err := w.coll.InsertOne(ctx, r); err != nil {
				w.log.Warn("audit insert failed", zap.Error(err))
			}
			cancel()
		}
	}
}
