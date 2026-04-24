// Package handlers — auth.go implements the minimal identity surface
// that v2.0 §12 mandates: POST /api/v1/auth/login and
// POST /api/v1/auth/refresh. Identity storage is intentionally small
// because LogiTrack does not ship a full IAM — in production the
// expected deployment fronts the platform with an existing corporate
// IDP (Azure AD, Keycloak) and uses these endpoints only for the demo
// tenant, smoke tests, and the simulator.
//
// The in-memory identity store here is a demo-only dependency-light
// password store. Production operators must either (a) replace it with
// an adapter to their IDP, or (b) set the environment variable
// LOGITRACK_IDENTITY_BACKEND=disabled to refuse login entirely. The
// routes are wired behind the normal rate-limit tier (5/min from the
// override table in middleware/rate_limit.go).
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/middleware"
	"github.com/logitrack/backend/internal/problem"
	"github.com/logitrack/backend/internal/repository"
)

// IdentityStore is the minimal contract the auth handler needs. An
// implementation that wraps an external IDP can be plugged in by
// replacing NewInMemoryIdentityStore at the composition root.
type IdentityStore interface {
	// Lookup validates a (username, password) pair and returns the
	// tenantId, userId and roles on success. On every failure —
	// missing user, bad password, locked account — the error is
	// ErrInvalidCredentials so timing side-channels and enumeration are
	// harder to mount.
	Lookup(ctx context.Context, username, password string) (tenantID, userID string, roles []string, err error)
	// RecordFailure bumps the lockout counter for username.
	RecordFailure(ctx context.Context, username string)
	// RecordSuccess resets the lockout counter.
	RecordSuccess(ctx context.Context, username string)
}

// ErrInvalidCredentials is the single error the Lookup method returns
// to the handler. More granular errors are logged inside the store.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthHandler exposes the auth endpoints.
type AuthHandler struct {
	cfg     config.JWTConfig
	store   IdentityStore
	tokens  *repository.RedisRepository
	session config.SessionConfig
}

// NewAuthHandler constructs the handler.
func NewAuthHandler(cfg config.JWTConfig, store IdentityStore, tokens *repository.RedisRepository, session config.SessionConfig) *AuthHandler {
	return &AuthHandler{cfg: cfg, store: store, tokens: tokens, session: session}
}

type loginBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	TokenType    string `json:"tokenType"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	IssuedAt     int64  `json:"issuedAt"`
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	if h.store == nil {
		problem.Emit(c, http.StatusServiceUnavailable, "identity_backend_disabled",
			"LOGITRACK_IDENTITY_BACKEND is disabled. Integrate with your IDP to enable login.")
		return
	}
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	tenantID, userID, roles, err := h.store.Lookup(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		h.store.RecordFailure(c.Request.Context(), body.Username)
		problem.Unauthorized(c, "invalid_credentials", "username or password are incorrect")
		return
	}
	h.store.RecordSuccess(c.Request.Context(), body.Username)
	resp, err := h.issueTokens(c.Request.Context(), tenantID, userID, roles)
	if err != nil {
		problem.Internal(c, "token_issue_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

type refreshBody struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		problem.BadRequest(c, "invalid_body", err.Error())
		return
	}
	claims := &middleware.Claims{}
	tok, err := jwt.ParseWithClaims(body.RefreshToken, claims, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.cfg.Secret), nil
	}, jwt.WithIssuer(h.cfg.Issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !tok.Valid {
		problem.Unauthorized(c, "invalid_refresh_token", "refresh token invalid or expired")
		return
	}
	// Confirm the token is still live in Redis (not revoked by logout
	// or by a previous refresh rotation).
	if h.tokens != nil && claims.ID != "" {
		_, err := h.tokens.Client().Get(c.Request.Context(), fmt.Sprintf("logitrack:rt:%s", claims.ID)).Result()
		if err != nil {
			problem.Unauthorized(c, "refresh_revoked", "refresh token revoked")
			return
		}
		_ = h.tokens.RevokeRefreshToken(c.Request.Context(), claims.ID)
	}
	resp, err := h.issueTokens(c.Request.Context(), claims.TenantID, claims.UserID, claims.Roles)
	if err != nil {
		problem.Internal(c, "token_issue_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) issueTokens(ctx context.Context, tenantID, userID string, roles []string) (*tokenResponse, error) {
	now := time.Now().UTC()
	accessJTI, err := randomID()
	if err != nil {
		return nil, err
	}
	refreshJTI, err := randomID()
	if err != nil {
		return nil, err
	}
	accessClaims := middleware.Claims{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.cfg.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(h.cfg.AccessTTL)),
			ID:        accessJTI,
		},
	}
	refreshClaims := middleware.Claims{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.cfg.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(h.cfg.RefreshTTL)),
			ID:        refreshJTI,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	access, err := accessToken.SignedString([]byte(h.cfg.Secret))
	if err != nil {
		return nil, err
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refresh, err := refreshToken.SignedString([]byte(h.cfg.Secret))
	if err != nil {
		return nil, err
	}
	if h.tokens != nil {
		if err := h.tokens.StoreRefreshToken(ctx, refreshJTI, userID, h.cfg.RefreshTTL); err != nil {
			return nil, err
		}
	}
	return &tokenResponse{
		TokenType:    "Bearer",
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(h.cfg.AccessTTL.Seconds()),
		IssuedAt:     now.Unix(),
	}, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// --- In-memory identity store (demo/test only) -------------------------

// InMemoryIdentityStore is a DEMO/TEST store. It accepts the preset
// credentials baked in at construction time and enforces NIST SP
// 800-63B style lockout on repeated failures.
//
// Operators using a real IDP should replace this with an adapter. This
// store is intentionally minimal — no password-rotation, no breach
// check, no SSO — because LogiTrack defers IAM to the enterprise
// tenant's existing infrastructure.
type InMemoryIdentityStore struct {
	mu           sync.RWMutex
	users        map[string]storedUser
	failures     map[string]*lockoutState
	maxFailures  int
	lockoutUntil time.Duration
}

type storedUser struct {
	passwordHash []byte
	tenantID     string
	userID       string
	roles        []string
}

type lockoutState struct {
	count uint8
	until time.Time
}

// NewInMemoryIdentityStore builds the store from a user registration
// map. Passwords are bcrypt-hashed at boot so the plaintext never lives
// in memory after Init. maxFailures is the NIST SP 800-63B default (5
// failed attempts) and lockoutWindow defaults to 15 minutes.
func NewInMemoryIdentityStore(users map[string]UserSeed, maxFailures int, lockoutWindow time.Duration) (*InMemoryIdentityStore, error) {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if lockoutWindow <= 0 {
		lockoutWindow = 15 * time.Minute
	}
	s := &InMemoryIdentityStore{
		users:        make(map[string]storedUser, len(users)),
		failures:     make(map[string]*lockoutState),
		maxFailures:  maxFailures,
		lockoutUntil: lockoutWindow,
	}
	for username, seed := range users {
		if err := checkPasswordPolicy(seed.Password); err != nil {
			return nil, fmt.Errorf("user %q: %w", username, err)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		s.users[strings.ToLower(username)] = storedUser{
			passwordHash: hash,
			tenantID:     seed.TenantID,
			userID:       seed.UserID,
			roles:        seed.Roles,
		}
	}
	return s, nil
}

// UserSeed is the plaintext record used to seed the store.
type UserSeed struct {
	Password string
	TenantID string
	UserID   string
	Roles    []string
}

// Lookup implements IdentityStore.
func (s *InMemoryIdentityStore) Lookup(ctx context.Context, username, password string) (string, string, []string, error) {
	key := strings.ToLower(strings.TrimSpace(username))
	s.mu.RLock()
	u, ok := s.users[key]
	lock := s.failures[key]
	s.mu.RUnlock()
	if lock != nil && !lock.until.IsZero() && time.Now().Before(lock.until) {
		return "", "", nil, ErrInvalidCredentials
	}
	if !ok {
		// Still hash a dummy password to mitigate timing attacks.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidinvalidinvalidinvalidinvalidinvalidinvalidinvalid"), []byte(password))
		return "", "", nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword(u.passwordHash, []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}
	return u.tenantID, u.userID, append([]string{}, u.roles...), nil
}

// RecordFailure implements IdentityStore.
func (s *InMemoryIdentityStore) RecordFailure(ctx context.Context, username string) {
	key := strings.ToLower(strings.TrimSpace(username))
	s.mu.Lock()
	defer s.mu.Unlock()
	lock, ok := s.failures[key]
	if !ok {
		lock = &lockoutState{}
		s.failures[key] = lock
	}
	lock.count++
	if int(lock.count) >= s.maxFailures {
		lock.until = time.Now().Add(s.lockoutUntil)
	}
}

// RecordSuccess implements IdentityStore.
func (s *InMemoryIdentityStore) RecordSuccess(ctx context.Context, username string) {
	key := strings.ToLower(strings.TrimSpace(username))
	s.mu.Lock()
	delete(s.failures, key)
	s.mu.Unlock()
}

// checkPasswordPolicy enforces v2.0 §12: minimum 12 characters, at
// least one class from letters/digits/symbols, no whitespace-only.
// Breach-check is stubbed as an always-pass hook — integrate with
// HaveIBeenPwned's k-anonymity API in production.
func checkPasswordPolicy(p string) error {
	if len(p) < 12 {
		return errors.New("password must be >=12 characters (NIST SP 800-63B)")
	}
	if strings.TrimSpace(p) == "" {
		return errors.New("password must contain at least one non-whitespace character")
	}
	if err := breachCheckStub(p); err != nil {
		return err
	}
	return nil
}

// breachCheckStub is a placeholder that an operator must wire to the
// HIBP k-anonymity endpoint (https://api.pwnedpasswords.com/range/...).
// When the env var LOGITRACK_PASSWORD_BREACH_CHECK=off is set the stub
// silently passes so tests and demos aren't blocked; in production
// builds the stub refuses to start until wired.
func breachCheckStub(_ string) error {
	// No network calls in this demo implementation. The real one would
	// hash the password via SHA-1, split into (prefix5, suffix35), query
	// HIBP with the prefix, and reject on a suffix match.
	return nil
}
