package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders writes the baseline security headers described in
// docs/SECURITY.md. They complement the CORS middleware and the CSP
// delivered by the nginx frontend.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// A deliberately minimal CSP. The SPA is served from a separate
		// origin with its own CSP; this one covers the JSON API surface.
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Next()
	}
}
