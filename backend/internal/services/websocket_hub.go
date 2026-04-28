package services

import (
	"context"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/modules/logistics"
	"github.com/logitrack/backend/internal/repository"
)

// WebSocketHub is an in-process fan-out registry for connected
// WebSocket clients. Messages received on the Redis pub/sub channel
// are dispatched to every client whose subscription filter matches.
//
// The hub is deliberately simple: it is not sharded across processes.
// When horizontal scaling crosses the single-instance ceiling the
// design migrates to Kafka with a consumer per replica, and the hub
// becomes a thin per-process router.
type WebSocketHub struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber
	redis       *repository.RedisRepository
	log         *zap.Logger
}

// Subscriber represents a single WebSocket client.
type Subscriber struct {
	ID         string
	TenantID   string
	ShipmentID string   // "" = subscribe to everything in tenant
	Carriers   []string // optional filter on carrier names
	Outbound   chan []byte
}

// NewWebSocketHub constructs a hub and starts no goroutines; the
// caller invokes Run to begin consuming events.
func NewWebSocketHub(redis *repository.RedisRepository, log *zap.Logger) *WebSocketHub {
	return &WebSocketHub{
		subscribers: make(map[string]*Subscriber),
		redis:       redis,
		log:         log,
	}
}

// Register adds a subscriber and returns a function that removes it.
func (h *WebSocketHub) Register(s *Subscriber) func() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscribers[s.ID] = s
	h.log.Info("ws subscriber registered", zap.String("id", s.ID), zap.String("tenant", s.TenantID))
	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.subscribers, s.ID)
		close(s.Outbound)
	}
}

// Broadcast pushes an event to every matching subscriber.
func (h *WebSocketHub) Broadcast(evt logistics.TrackingEvent) {
	payload, err := json.Marshal(evt)
	if err != nil {
		h.log.Warn("ws marshal event failed", zap.Error(err))
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, sub := range h.subscribers {
		if sub.TenantID != "" && sub.TenantID != evt.TenantID {
			continue
		}
		if sub.ShipmentID != "" && sub.ShipmentID != evt.ShipmentID {
			continue
		}
		select {
		case sub.Outbound <- payload:
		default:
			// Slow consumer — drop the message rather than block the
			// fan-out loop. The client will resync on next reconnect.
			h.log.Warn("ws slow consumer drop", zap.String("subscriber", sub.ID))
		}
	}
}

// Run pumps events from Redis pub/sub into the hub until the context
// is cancelled. Call in its own goroutine at server boot.
func (h *WebSocketHub) Run(ctx context.Context) error {
	sub := h.redis.SubscribeTrackingEvents(ctx)
	defer sub.Close()
	ch := sub.Channel()
	h.log.Info("ws hub subscribed to redis channel", zap.String("channel", repository.ChannelTrackingEvents))
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			var evt logistics.TrackingEvent
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				h.log.Warn("ws hub unmarshal", zap.Error(err))
				continue
			}
			h.Broadcast(evt)
		}
	}
}
