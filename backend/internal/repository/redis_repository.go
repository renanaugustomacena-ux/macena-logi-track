package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/models"
)

// Channel names used for pub/sub fan-out. The namespace prefix avoids
// collisions with other Redis tenants when the instance is shared.
const (
	ChannelTrackingEvents = "logitrack:events:tracking"
	KeyLatestPositionTmpl = "logitrack:position:%s"       // %s = shipmentID
	KeyRateLimitTmpl      = "logitrack:rl:%s:%s"          // %s tenant, %s bucket
	KeyRefreshTokenTmpl   = "logitrack:rt:%s"             // %s = jti
)

// PositionTTL is how long we keep the "last known position" cache
// entry. Shipments that have not emitted a telematics event in this
// window are assumed stale and re-fetched from Mongo.
const PositionTTL = 15 * time.Minute

// RedisRepository wraps the Redis client with domain-specific methods.
type RedisRepository struct {
	client *redis.Client
	log    *zap.Logger
}

// NewRedis builds the repository and validates connectivity via PING.
func NewRedis(ctx context.Context, cfg config.RedisConfig, log *zap.Logger) (*RedisRepository, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("redis parse url: %w", err)
	}
	opts.PoolSize = cfg.PoolSize
	opts.DialTimeout = cfg.Timeout
	opts.ReadTimeout = cfg.Timeout
	opts.WriteTimeout = cfg.Timeout
	client := redis.NewClient(opts)
	pingCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisRepository{client: client, log: log}, nil
}

// Client returns the wrapped client for health checks.
func (r *RedisRepository) Client() *redis.Client { return r.client }

// Close releases underlying network resources.
func (r *RedisRepository) Close() error { return r.client.Close() }

// CacheLatestPosition upserts the last known position of a shipment
// in Redis. The dashboard and WebSocket hub read from here rather
// than hitting Mongo on every client reconnect.
func (r *RedisRepository) CacheLatestPosition(ctx context.Context, shipmentID string, wp models.Waypoint) error {
	payload, err := json.Marshal(wp)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(KeyLatestPositionTmpl, shipmentID)
	return r.client.Set(ctx, key, payload, PositionTTL).Err()
}

// GetLatestPosition returns the cached position or ErrNotFound.
func (r *RedisRepository) GetLatestPosition(ctx context.Context, shipmentID string) (*models.Waypoint, error) {
	key := fmt.Sprintf(KeyLatestPositionTmpl, shipmentID)
	raw, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var wp models.Waypoint
	if err := json.Unmarshal(raw, &wp); err != nil {
		return nil, err
	}
	return &wp, nil
}

// PublishTrackingEvent publishes an event on the tracking channel.
// Fan-out to connected WebSocket clients happens in the subscriber
// loop of the websocket hub.
func (r *RedisRepository) PublishTrackingEvent(ctx context.Context, evt models.TrackingEvent) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, ChannelTrackingEvents, payload).Err()
}

// SubscribeTrackingEvents returns a Redis pub/sub subscription the
// caller can range over. Callers must invoke Close on the returned
// subscription when finished.
func (r *RedisRepository) SubscribeTrackingEvents(ctx context.Context) *redis.PubSub {
	return r.client.Subscribe(ctx, ChannelTrackingEvents)
}

// StoreRefreshToken persists a refresh-token JTI with its expiry so
// revocation and one-time-use semantics can be enforced.
func (r *RedisRepository) StoreRefreshToken(ctx context.Context, jti, userID string, ttl time.Duration) error {
	key := fmt.Sprintf(KeyRefreshTokenTmpl, jti)
	return r.client.Set(ctx, key, userID, ttl).Err()
}

// RevokeRefreshToken deletes the stored JTI (used on logout or
// rotation).
func (r *RedisRepository) RevokeRefreshToken(ctx context.Context, jti string) error {
	key := fmt.Sprintf(KeyRefreshTokenTmpl, jti)
	return r.client.Del(ctx, key).Err()
}
