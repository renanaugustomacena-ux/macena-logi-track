// Package handlers — auth.go implements the minimal identity surface:
// POST /api/v1/auth/login. The kit deliberately does NOT ship a
// refresh endpoint or full IAM. Customers integrate their corporate
// IDP (Keycloak, Azure AD, Okta) at the IdentityStore seam below;
// session lifecycle becomes the IDP's responsibility.
//
// The in-memory identity store is demo-only. Production operators
// either (a) replace it with an IDP adapter, or (b) set
// LOGITRACK_IDENTITY_BACKEND=disabled to refuse login entirely.
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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
	cfg    config.JWTConfig
	store  IdentityStore
	tokens *repository.RedisRepository
}

// NewAuthHandler constructs the handler. The Redis dependency is
// retained for future per-tenant lockout coordination across
// instances; it is unused in the single-instance kit deployment.
func NewAuthHandler(cfg config.JWTConfig, store IdentityStore, tokens *repository.RedisRepository) *AuthHandler {
	return &AuthHandler{cfg: cfg, store: store, tokens: tokens}
}

type loginBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	TokenType   string `json:"tokenType"`
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
	IssuedAt    int64  `json:"issuedAt"`
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
	resp, err := h.issueAccessToken(c.Request.Context(), tenantID, userID, roles)
	if err != nil {
		problem.Internal(c, "token_issue_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) issueAccessToken(_ context.Context, tenantID, userID string, roles []string) (*tokenResponse, error) {
	now := time.Now().UTC()
	claims := middleware.Claims{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.cfg.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(h.cfg.AccessTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(h.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}
	return &tokenResponse{
		TokenType:   "Bearer",
		AccessToken: signed,
		ExpiresIn:   int(h.cfg.AccessTTL.Seconds()),
		IssuedAt:    now.Unix(),
	}, nil
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

// checkPasswordPolicy enforces NIST SP 800-63B-style minimums:
// >= 12 characters, no whitespace-only.
//
// HIBP breach-check is intentionally NOT performed inside this kit:
// the seed mechanism is meant for a single demo user. Production
// deployments use the customer's IDP (Keycloak, Azure AD, Okta), which
// already runs its own breach-check policy. This contract is stricter
// than the previous silent-stub: a kit deployed in production with
// APP_ENV=production AND LOGITRACK_IDENTITY_BACKEND=memory must
// explicitly opt in via LOGITRACK_IDENTITY_DEMO_BREACH_ACK=true.
func checkPasswordPolicy(p string) error {
	if len(p) < 12 {
		return errors.New("password must be >=12 characters (NIST SP 800-63B)")
	}
	if strings.TrimSpace(p) == "" {
		return errors.New("password must contain at least one non-whitespace character")
	}
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") &&
		strings.EqualFold(os.Getenv("LOGITRACK_IDENTITY_BACKEND"), "memory") &&
		!strings.EqualFold(os.Getenv("LOGITRACK_IDENTITY_DEMO_BREACH_ACK"), "true") {
		return errors.New("production + memory identity backend requires LOGITRACK_IDENTITY_DEMO_BREACH_ACK=true (acknowledge that the seed password has been verified against HIBP externally)")
	}
	return nil
}
