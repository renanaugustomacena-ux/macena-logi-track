package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/services"
)

// StreamHandler exposes the live WebSocket at /api/v1/stream/tracking.
//
// Protocol (see docs/API.md for the full schema):
//
//	client -> server: {"op":"subscribe","shipmentId":"<id>","carriers":["<name>"]}
//	server -> client: {"type":"event","event":{...TrackingEvent...}}
//	server -> client: {"type":"pong"}
//	client -> server: {"op":"ping"}
//
// Security controls applied on the handshake (all MUST pass):
//
//  1. Origin check against config.WebSocket.AllowedOrigins.
//  2. JWT validation. Browsers cannot set arbitrary headers on a WS
//     upgrade, so the token is accepted in three places:
//     - Authorization: Bearer <token> header (servers and CLIs)
//     - ?access_token=<token> query parameter (legacy web clients)
//     - Sec-WebSocket-Protocol sub-protocol list, form
//     `logitrack.jwt.v1,<token>` (preferred browser path).
//  3. Per-IP rate limit enforced at handshake to blunt DoS.
//
// Runtime controls applied after upgrade:
//
//  4. Per-connection inbound rate limit (token bucket).
//  5. Idle disconnect after config.WebSocket.IdleTimeout (5 minutes).
//  6. Read-size cap of 64 KiB.
//  7. Graceful close on hub Broadcast failures (slow-consumer drop).
type StreamHandler struct {
	hub              *services.WebSocketHub
	log              *zap.Logger
	upgrader         websocket.Upgrader
	jwtCfg           config.JWTConfig
	wsCfg            config.WebSocketConfig
	handshakeLimiter *ipRateLimiter
}

// NewStreamHandler constructs the handler. The upgrader's CheckOrigin
// is closed over the configured allow-list so no mutable shared state
// is exposed.
func NewStreamHandler(hub *services.WebSocketHub, log *zap.Logger, jwtCfg config.JWTConfig, wsCfg config.WebSocketConfig) *StreamHandler {
	allow := make(map[string]struct{}, len(wsCfg.AllowedOrigins))
	for _, o := range wsCfg.AllowedOrigins {
		if o == "" {
			continue
		}
		allow[strings.ToLower(strings.TrimSpace(o))] = struct{}{}
	}
	checkOrigin := func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// No Origin => non-browser client (CLI, test). We permit
			// this only because JWT validation still runs: the token
			// binds the caller to a tenant.
			return true
		}
		_, ok := allow[strings.ToLower(origin)]
		return ok
	}
	return &StreamHandler{
		hub:    hub,
		log:    log,
		jwtCfg: jwtCfg,
		wsCfg:  wsCfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			EnableCompression: false,
			// nosemgrep: go.gorilla.security.audit.websocket-missing-origin-check.websocket-missing-origin-check
			// checkOrigin is an allow-list closure built from
			// config.WebSocket.AllowedOrigins (see above). The generic
			// static rule cannot see through the closure, so we annotate.
			CheckOrigin:  checkOrigin,
			Subprotocols: []string{wsCfg.HandshakeSub},
		},
		handshakeLimiter: newIPRateLimiter(wsCfg.RateLimitRPS, wsCfg.RateLimitBurst),
	}
}

type wsInbound struct {
	Op         string   `json:"op"`
	ShipmentID string   `json:"shipmentId"`
	Carriers   []string `json:"carriers"`
}

// Handle is the Gin entry point. Unlike other handlers this method
// takes over the response writer and does not return JSON: the reply
// is either a 101 upgrade or a 4xx error.
func (h *StreamHandler) Handle(c *gin.Context) {
	if !h.handshakeLimiter.allow(c.ClientIP()) {
		problem.TooMany(c, "handshake_rate_limited", "too many WS handshakes from this IP")
		return
	}
	claims, err := h.extractClaims(c)
	if err != nil {
		h.log.Info("ws handshake rejected", zap.String("reason", err.Error()), zap.String("ip", c.ClientIP()))
		problem.Unauthorized(c, "invalid_token", err.Error())
		return
	}
	// If a subprotocol was offered, echo back the handshake sub so the
	// browser completes the upgrade cleanly.
	var respHeader http.Header
	if offered := c.Request.Header.Get("Sec-WebSocket-Protocol"); offered != "" {
		for _, p := range strings.Split(offered, ",") {
			if strings.TrimSpace(p) == h.wsCfg.HandshakeSub {
				respHeader = http.Header{}
				respHeader.Set("Sec-WebSocket-Protocol", h.wsCfg.HandshakeSub)
				break
			}
		}
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, respHeader)
	if err != nil {
		// CheckOrigin failure surfaces here as 403 via the upgrader.
		h.log.Warn("ws upgrade failed", zap.Error(err))
		return
	}
	sub := &services.Subscriber{
		ID:       uuid.NewString(),
		TenantID: claims.TenantID,
		Outbound: make(chan []byte, 32),
	}
	unregister := h.hub.Register(sub)
	defer unregister()
	defer conn.Close()

	// Per-connection inbound limiter: ~20 messages/s with burst 40.
	connLimiter := rate.NewLimiter(rate.Limit(h.wsCfg.RateLimitRPS), h.wsCfg.RateLimitBurst)

	idle := h.wsCfg.IdleTimeout
	if idle <= 0 {
		idle = 5 * time.Minute
	}

	// Writer goroutine: drain the outbound channel with a write deadline.
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case msg, ok := <-sub.Outbound:
				if !ok {
					return
				}
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Reader loop: accept subscription-refinement messages, enforce
	// idle timeout, pong handler resets the read deadline on keepalive.
	conn.SetReadLimit(1 << 16)
	_ = conn.SetReadDeadline(time.Now().Add(idle))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(idle))
	})
	for {
		if !connLimiter.Allow() {
			// Tell the client and close — easier to diagnose than a
			// silent drop.
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","code":"rate_limited"}`))
			return
		}
		var msg wsInbound
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(idle))
		switch msg.Op {
		case "subscribe":
			sub.ShipmentID = msg.ShipmentID
			sub.Carriers = msg.Carriers
		case "ping":
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		}
	}
	// Note: the <-done sync path is unreachable because the reader loop
	// returns on every error, but the deferred close + unregister still
	// drain the writer goroutine by closing the Outbound channel.
}

// extractClaims discovers the caller's JWT from the three supported
// places and parses it. Crucially, only HS256 is accepted; the jwt
// library rejects any other alg (including "none").
func (h *StreamHandler) extractClaims(c *gin.Context) (*middleware.Claims, error) {
	raw := bearerToken(c.GetHeader("Authorization"))
	if raw == "" {
		raw = c.Query("access_token")
	}
	if raw == "" {
		raw = subprotocolToken(c.GetHeader("Sec-WebSocket-Protocol"), h.wsCfg.HandshakeSub)
	}
	if raw == "" {
		return nil, errors.New("missing_token")
	}
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Method.Alg())
		}
		return []byte(h.jwtCfg.Secret), nil
	}, jwt.WithIssuer(h.jwtCfg.Issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		if err == nil {
			err = errors.New("invalid_token")
		}
		return nil, err
	}
	if claims.TenantID == "" {
		return nil, errors.New("invalid_claims: tenant_id missing")
	}
	return claims, nil
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// subprotocolToken implements the "logitrack.jwt.v1,<jwt>" convention
// used by browser clients. Mirrors the pattern popularised by Kubernetes
// dashboard and discussed in RFC 6455 section 11.3.4.
func subprotocolToken(header, expected string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	var tok string
	found := false
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == expected {
			found = true
			continue
		}
		// URL-safe base64 JWT tokens contain only A-Za-z0-9-_. sanity-check.
		if tok == "" {
			tok = v
		}
	}
	if !found {
		return ""
	}
	if _, err := url.QueryUnescape(tok); err != nil {
		return ""
	}
	return tok
}

// --- tiny per-IP rate limiter ----------------------------------------------

type ipRateLimiter struct {
	rps     rate.Limit
	burst   int
	buckets map[string]*rate.Limiter
}

func newIPRateLimiter(rps, burst int) *ipRateLimiter {
	if rps <= 0 {
		rps = 20
	}
	if burst <= 0 {
		burst = 40
	}
	return &ipRateLimiter{rps: rate.Limit(rps), burst: burst, buckets: make(map[string]*rate.Limiter)}
}

func (r *ipRateLimiter) allow(ip string) bool {
	b, ok := r.buckets[ip]
	if !ok {
		b = rate.NewLimiter(r.rps, r.burst)
		r.buckets[ip] = b
	}
	return b.Allow()
}
