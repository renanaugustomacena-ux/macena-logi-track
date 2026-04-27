package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/audit"
	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/integrations/aida"
	"github.com/logitrack/backend/internal/integrations/albo"
	"github.com/logitrack/backend/internal/integrations/rfi"
	"github.com/logitrack/backend/internal/integrations/telepass"
	"github.com/logitrack/backend/internal/middleware"
)

// Dependencies bundles the singletons a route-setup call needs.
type Dependencies struct {
	Config         *config.Config
	Logger         *zap.Logger
	Health         *HealthHandler
	Ready          *ReadyHandler
	Metrics        *MetricsHandler
	Ship           *ShipmentHandler
	Track          *TrackingHandler
	Route          *RouteHandler
	Stream         *StreamHandler
	ETA            *ETAHandler
	Fleet          *FleetHandler
	Auth           *AuthHandler
	Audit          *audit.Writer
	AidaClient     *aida.Client
	RfiClient      *rfi.Client
	TelepassClient *telepass.Client
	AlboClient     *albo.Client
}

// Register wires every route onto the provided engine. Routes are
// grouped by access level (public, authenticated, admin) to keep
// middleware composition explicit.
func Register(r *gin.Engine, deps Dependencies) {
	// Recovery must wrap every other middleware, so a panic anywhere
	// in the chain (including the security/CORS/logger middlewares
	// themselves) is converted to a 500 instead of crashing the worker.
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(deps.Config.HTTP))
	r.Use(middleware.StructuredLogger(deps.Logger))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.PromMiddleware())
	r.Use(middleware.RateLimiter(deps.Config.RateLimit))

	public := r.Group("/api")
	{
		public.GET("/health", deps.Health.Get)
		if deps.Ready != nil {
			public.GET("/ready", deps.Ready.Get)
		}
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

	// Auth routes are public (login + refresh produce tokens). Rate
	// limiting is applied globally; auth-tier override lives in the
	// rate_limit middleware.
	if deps.Auth != nil {
		authGroup := r.Group("/api/v1/auth")
		{
			authGroup.POST("/login", deps.Auth.Login)
			authGroup.POST("/refresh", deps.Auth.Refresh)
		}
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(deps.Config.JWT))
	if deps.Audit != nil {
		v1.Use(middleware.AuditMiddleware(deps.Audit))
	}
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
		if deps.Fleet != nil {
			vehicles := v1.Group("/vehicles")
			{
				vehicles.POST("", deps.Fleet.CreateVehicle)
				vehicles.GET("", deps.Fleet.ListVehicles)
				vehicles.GET("/:id", deps.Fleet.GetVehicle)
			}
			drivers := v1.Group("/drivers")
			{
				drivers.POST("", deps.Fleet.CreateDriver)
				drivers.GET("", deps.Fleet.ListDrivers)
				drivers.GET("/:id", deps.Fleet.GetDriver)
			}
			geofences := v1.Group("/geofences")
			{
				geofences.POST("", deps.Fleet.CreateGeofence)
				geofences.GET("", deps.Fleet.ListGeofences)
				geofences.GET("/:id", deps.Fleet.GetGeofence)
			}
		}
	}

	// Suppress unused-field warnings for integration clients — they are
	// held in Dependencies so future handlers (customs declarations,
	// rail slot bookings, toll feeds, carrier verification) can wire
	// them without changing this signature.
	_ = deps.AidaClient
	_ = deps.RfiClient
	_ = deps.TelepassClient
	_ = deps.AlboClient
}
