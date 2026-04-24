// Package middleware holds Gin handlers that augment the request
// pipeline with cross-cutting concerns: authentication, logging,
// rate-limiting and CORS.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/problem"
)

// ClaimsContextKey is the Gin-context key under which authenticated
// claims are stored after successful validation.
const ClaimsContextKey = "logitrack.claims"

// Claims is the superset of the JWT body fields LogiTrack honours.
type Claims struct {
	TenantID string   `json:"tenantId"`
	UserID   string   `json:"sub"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTAuth returns a Gin middleware that validates bearer tokens
// signed with the configured HS256 secret. On success the parsed
// claims are attached to the context under ClaimsContextKey.
func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abort(c, http.StatusUnauthorized, "missing_authorization")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abort(c, http.StatusUnauthorized, "invalid_authorization_scheme")
			return
		}
		token, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(tk *jwt.Token) (interface{}, error) {
			if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.Secret), nil
		}, jwt.WithIssuer(cfg.Issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			abort(c, http.StatusUnauthorized, "invalid_token")
			return
		}
		claims, ok := token.Claims.(*Claims)
		if !ok || claims.TenantID == "" {
			abort(c, http.StatusUnauthorized, "invalid_claims")
			return
		}
		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}

// MustClaims extracts the Claims set by JWTAuth or panics — the
// middleware contract guarantees presence on any authenticated route.
func MustClaims(c *gin.Context) *Claims {
	v, exists := c.Get(ClaimsContextKey)
	if !exists {
		panic("claims missing: JWTAuth middleware not wired")
	}
	return v.(*Claims)
}

// RequireRole returns a middleware that admits the request only if
// the authenticated principal holds at least one of `roles`.
func RequireRole(roles ...string) gin.HandlerFunc {
	wanted := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		wanted[r] = struct{}{}
	}
	return func(c *gin.Context) {
		claims := MustClaims(c)
		for _, r := range claims.Roles {
			if _, ok := wanted[r]; ok {
				c.Next()
				return
			}
		}
		abort(c, http.StatusForbidden, "insufficient_role")
	}
}

func abort(c *gin.Context, status int, code string) {
	problem.Emit(c, status, code, "")
}
