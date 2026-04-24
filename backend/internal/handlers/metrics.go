package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/obs"
)

// MetricsHandler renders the Prometheus text-exposition format using
// the shared obs package. Keeping the handler thin makes it easy to
// swap implementations if we adopt the official prometheus/client_golang
// library in the future.
type MetricsHandler struct {
	service string
	version string
	started time.Time
}

// NewMetricsHandler builds a handler pre-populated with the build info.
func NewMetricsHandler(service, version string) *MetricsHandler {
	return &MetricsHandler{service: service, version: version, started: time.Now()}
}

// Handler returns the Gin HandlerFunc that exposes /metrics.
func (m *MetricsHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		obs.Render(c.Writer, m.service, m.version, m.started)
	}
}

// IncHTTP / IncWS* re-exported for callers that already depend on the
// handlers package so they do not need a second import.
func IncHTTP(method, path string, status int) { obs.IncHTTP(method, path, status) }
func IncWSConnect()                           { obs.IncWSConnect() }
func IncWSDisconnect()                        { obs.IncWSDisconnect() }
func IncWSBroadcast()                         { obs.IncWSBroadcast() }
