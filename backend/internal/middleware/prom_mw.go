package middleware

import (
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/obs"
)

// PromMiddleware tallies every HTTP request into the metrics registry
// exposed at /metrics. To keep cardinality bounded it coarsens the
// path by replacing literal UUIDs / ObjectIDs with the Gin route
// template (`/shipments/:id`).
func PromMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		obs.IncHTTP(c.Request.Method, coarsenPath(c.FullPath(), c.Request.URL.Path), c.Writer.Status())
	}
}

// Matches a UUID or 24-hex Mongo ObjectID in a URL segment.
var idSegmentRE = regexp.MustCompile(`/[0-9a-fA-F]{24}|/[0-9a-fA-F-]{32,36}`)

func coarsenPath(route, actual string) string {
	if route != "" {
		return route
	}
	return idSegmentRE.ReplaceAllString(actual, "/:id")
}
