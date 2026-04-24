package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/middleware"
)

// Dependencies bundles the singletons a route-setup call needs.
type Dependencies struct {
	Config  *config.Config
	Logger  *zap.Logger
	Health  *HealthHandler
	Metrics *MetricsHandler
	Ship    *ShipmentHandler
	Track   *TrackingHandler
	Route   *RouteHandler
	Stream  *StreamHandler
	ETA     *ETAHandler
}

// Register wires every route onto the provided engine. Routes are
// grouped by access level (public, authenticated, admin) to keep
// middleware composition explicit.
func Register(r *gin.Engine, deps Dependencies) {
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(deps.Config.HTTP))
	r.Use(middleware.StructuredLogger(deps.Logger))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.PromMiddleware())
	r.Use(middleware.RateLimiter(deps.Config.RateLimit))
	r.Use(gin.Recovery())

	public := r.Group("/api")
	{
		public.GET("/health", deps.Health.Get)
	}
	// Prometheus exposition — mounted at /metrics per the platform
	// gate G4. No auth: in production the endpoint is fronted by the
	// service-mesh sidecar which restricts access to the monitoring
	// namespace. A static-allowlist is documented in docs/RUNBOOK.md.
	r.GET("/metrics", deps.Metrics.Handler())

	// WebSocket stream is authenticated inside the handler (JWT in
	// query / Sec-WebSocket-Protocol / Authorization), so it is mounted
	// outside the JWTAuth middleware group. Origin check and per-IP
	// handshake rate-limit are applied there too.
	r.GET("/api/v1/stream/tracking", deps.Stream.Handle)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(deps.Config.JWT))
	{
		shipments := v1.Group("/shipments")
		{
			shipments.POST("", deps.Ship.Create)
			shipments.GET("", deps.Ship.List)
			shipments.GET("/:id", deps.Ship.Get)
			shipments.POST("/:id/waypoints", deps.Ship.AddWaypoint)
			shipments.GET("/:id/trace", deps.Ship.Trace)
			shipments.GET("/:id/position", deps.Track.LatestPosition)
			shipments.GET("/:id/eta", deps.ETA.Get)
		}
		routes := v1.Group("/routes")
		{
			routes.POST("/optimize", deps.Route.Optimize)
		}
	}
}
