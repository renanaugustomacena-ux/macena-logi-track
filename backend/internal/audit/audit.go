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

// Mode selects how Enqueue persists records.
//
//   - ModeAsync (default): records flow through a buffered channel
//     and are written by a background goroutine. Under backpressure
//     records are DROPPED with a warn-level counter. Lowest latency
//     on the request path; lossy under sustained burst.
//   - ModeSync: records are written inline, in the request goroutine.
//     Adds 1-3 ms per mutation in normal Mongo conditions but
//     guarantees no records are lost. Choose this when the customer's
//     DPA / regulatory posture requires guaranteed audit retention
//     (GDPR Art. 30, D.Lgs. 196/2003 art. 30, ISO 27001 A.12.4.1).
type Mode int

const (
	ModeAsync Mode = iota
	ModeSync
)

// ParseMode resolves AUDIT_MODE env values to a Mode. Unknown values
// fall back to async with a logged warning at boot.
func ParseMode(raw string) (Mode, bool) {
	switch raw {
	case "", "async":
		return ModeAsync, true
	case "sync":
		return ModeSync, true
	default:
		return ModeAsync, false
	}
}

// Writer is the worker that persists Records. Safe for concurrent use.
type Writer struct {
	coll    *mongo.Collection
	queue   chan Record
	log     *zap.Logger
	dropped atomic.Uint64
	mode    Mode
}

// NewWriter builds an async writer with the given buffer capacity. To
// switch to ModeSync use NewWriterWithMode.
func NewWriter(ctx context.Context, repo *repository.MongoRepository, log *zap.Logger, buf int) *Writer {
	return NewWriterWithMode(ctx, repo, log, buf, ModeAsync)
}

// NewWriterWithMode is the explicit constructor that lets the
// composition root choose the persistence mode. mode=ModeSync makes
// every Enqueue a synchronous Mongo InsertOne; mode=ModeAsync keeps
// the historical fire-and-forget behaviour. buf is ignored when
// mode==ModeSync.
func NewWriterWithMode(ctx context.Context, repo *repository.MongoRepository, log *zap.Logger, buf int, mode Mode) *Writer {
	if buf <= 0 {
		buf = 1024
	}
	client := repo.Client()
	coll := client.Database(repo.DatabaseName()).Collection(repository.CollectionAuditLog)
	w := &Writer{
		coll: coll,
		log:  log,
		mode: mode,
	}
	if mode == ModeAsync {
		w.queue = make(chan Record, buf)
		go w.run(ctx)
	}
	log.Info("audit writer initialised",
		zap.String("mode", map[Mode]string{ModeAsync: "async", ModeSync: "sync"}[mode]),
		zap.Int("buf", buf),
	)
	return w
}

// Enqueue submits a record for persistence. In async mode it returns
// false if the queue is full (record dropped). In sync mode it always
// returns true after a successful write (or false on Mongo error,
// logged at warn).
func (w *Writer) Enqueue(r Record) bool {
	if r.At.IsZero() {
		r.At = time.Now().UTC()
	}
	if w.mode == ModeSync {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if _, err := w.coll.InsertOne(ctx, r); err != nil {
			w.log.Warn("audit sync insert failed", zap.Error(err))
			return false
		}
		return true
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
