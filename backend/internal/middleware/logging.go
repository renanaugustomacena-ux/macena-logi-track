package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDHeader is the canonical correlation-id header. Inbound
// values are trusted; when absent a new UUID is minted.
const RequestIDHeader = "X-Request-ID"

// RequestID assigns a correlation id to every request and exposes it
// in both the Gin context and the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set(RequestIDHeader, rid)
		c.Next()
	}
}

// StructuredLogger returns a middleware that logs every request as a
// single structured JSON record, populated with fields analytics
// pipelines downstream expect (status, duration, route, trace id).
func StructuredLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("remote", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", c.GetString("request_id")),
		}
		if claims, ok := c.Get(ClaimsContextKey); ok {
			if tc, cast := claims.(*Claims); cast {
				fields = append(fields, zap.String("tenant_id", tc.TenantID), zap.String("user_id", tc.UserID))
			}
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
			log.Error("http request", fields...)
			return
		}
		log.Info("http request", fields...)
	}
}
