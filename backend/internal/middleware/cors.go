package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/config"
)

// CORS returns a configured CORS middleware. The allow-list is read
// from configuration so production deployments can tighten it beyond
// the open localhost defaults used in development.
func CORS(cfg config.HTTPConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowed,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-Tenant-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
