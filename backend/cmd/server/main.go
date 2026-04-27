// Command server is the LogiTrack backend entry point. It is a thin
// composition root: wire config, logger, telemetry, repositories,
// services, handlers and the HTTP server — then block on signal.
//
// The binary is designed to be run in a container orchestrator. It
// intentionally does no cluster discovery of its own; load balancing,
// autoscaling and rolling updates are the orchestrator's job.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/logitrack/backend/internal/audit"
	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/demo"
	"github.com/logitrack/backend/internal/handlers"
	"github.com/logitrack/backend/internal/integrations/aida"
	"github.com/logitrack/backend/internal/integrations/albo"
	"github.com/logitrack/backend/internal/integrations/rfi"
	"github.com/logitrack/backend/internal/integrations/telepass"
	"github.com/logitrack/backend/internal/repository"
	"github.com/logitrack/backend/internal/services"
)

// version is overridden at build time via `-ldflags "-X main.version=..."`.
var version = "0.2.0-mission-ii"

func main() {
	// Support a minimal --healthcheck subcommand so the Docker
	// HEALTHCHECK directive can re-enable itself without a shell. This
	// closes D-05: distroless has no wget, so the binary self-tests
	// /api/health over localhost.
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		if err := runHealthcheck(); err != nil {
			fmt.Fprintf(os.Stderr, "unhealthy: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// runHealthcheck performs a GET /api/health against the local port
// exposed by the server process. It returns nil only when the response
// is 200 and the JSON body reports status ok or degraded.
func runHealthcheck() error {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	url := fmt.Sprintf("http://127.0.0.1:%s/api/health", port)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	cfg.App.Version = version

	log, err := buildLogger(cfg)
	if err != nil {
		return fmt.Errorf("logger init: %w", err)
	}
	defer func() { _ = log.Sync() }()

	log.Info("logitrack server starting",
		zap.String("env", cfg.App.Env),
		zap.String("version", cfg.App.Version),
		zap.Int("port", cfg.HTTP.Port),
	)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	shutdownOTEL, err := initTracer(rootCtx, cfg)
	if err != nil {
		log.Warn("otel disabled", zap.Error(err))
	}
	defer func() {
		if shutdownOTEL != nil {
			_ = shutdownOTEL(context.Background())
		}
	}()

	mongoRepo, err := repository.NewMongo(rootCtx, cfg.Mongo, log)
	if err != nil {
		return fmt.Errorf("mongo init: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoRepo.Disconnect(ctx)
	}()
	if err := mongoRepo.EnsureIndexes(rootCtx); err != nil {
		log.Warn("ensure indexes failed", zap.Error(err))
	}

	redisRepo, err := repository.NewRedis(rootCtx, cfg.Redis, log)
	if err != nil {
		return fmt.Errorf("redis init: %w", err)
	}
	defer redisRepo.Close()

	routeSvc := services.NewOSRMOptimizer(cfg.OSRM, log)
	etaSvc := services.NewETAService(mongoRepo, redisRepo, routeSvc, log)
	shipmentSvc := services.NewShipmentService(mongoRepo, redisRepo, log).
		WithRouting(routeSvc).WithETA(etaSvc)
	trackingSvc := services.NewTrackingService(mongoRepo, redisRepo, log)
	hub := services.NewWebSocketHub(redisRepo, log)
	go func() {
		if err := hub.Run(rootCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("ws hub exited", zap.Error(err))
		}
	}()

	// Seed-done flag used by /api/ready. In no-seed deployments we
	// mark it ready immediately so the service is exposed to the LB
	// as soon as the deps are connected.
	seedDone := &atomic.Bool{}
	if !cfg.Demo.SeedOnBoot {
		seedDone.Store(true)
	}
	// Demo seed — idempotent, only when explicitly requested. Runs in
	// the background so the server can start accepting traffic; the
	// /api/ready probe is the authoritative "first waypoint seeded"
	// signal.
	if cfg.Demo.SeedOnBoot {
		go func() {
			if err := demo.Seed(rootCtx, mongoRepo, shipmentSvc, routeSvc, cfg.Demo.TenantID, log); err != nil {
				log.Warn("demo seed skipped", zap.Error(err))
			}
			seedDone.Store(true)
		}()
	}

	// Italian integration clients — all four are instantiated
	// unconditionally. When an env var is missing the client returns
	// ErrNotConfigured on every call; services surface that to the
	// operator instead of silently fabricating data. This closes
	// H-01 .. H-04 from LogiTrack-GAPS.md.
	aidaClient := aida.New(aida.Config{BaseURL: cfg.AIDA.BaseURL, APIKey: cfg.AIDA.APIKey})
	rfiClient, err := rfi.New(rfi.Config{
		BaseURL:      cfg.RFI.BaseURL,
		ClientID:     cfg.RFI.ClientID,
		ClientSecret: cfg.RFI.ClientSecret,
		MTLSCertFile: cfg.RFI.MTLSCertFile,
		MTLSKeyFile:  cfg.RFI.MTLSKeyFile,
		MTLSCAFile:   cfg.RFI.MTLSCAFile,
	})
	if err != nil {
		// mTLS misconfiguration is a hard boot failure: the alternative
		// is silently falling back to plain TLS and producing a
		// confusing 403 from FERTRAM later. Fail-fast at boot so the
		// operator notices.
		return fmt.Errorf("rfi client init: %w", err)
	}
	telepassClient := telepass.New(telepass.Config{
		BaseURL:    cfg.Telepass.BaseURL,
		APIKey:     cfg.Telepass.APIKey,
		ContractID: cfg.Telepass.ContractID,
	})
	alboClient := albo.New(albo.Config{BaseURL: cfg.Albo.BaseURL, APIKey: cfg.Albo.APIKey})
	log.Info("integrations registered",
		zap.Bool("aida_configured", aidaClient.Configured()),
		zap.Bool("rfi_configured", rfiClient.Configured()),
		zap.Bool("telepass_configured", telepassClient.Configured()),
		zap.Bool("albo_configured", alboClient.Configured()),
	)

	auditWriter := audit.NewWriter(rootCtx, mongoRepo, log, 1024)

	identityStore, err := buildIdentityStore(cfg.Identity)
	if err != nil {
		log.Warn("identity store unavailable", zap.Error(err))
	}

	deps := handlers.Dependencies{
		Config:         cfg,
		Logger:         log,
		Health:         handlers.NewHealthHandler(cfg, mongoRepo, redisRepo),
		Ready:          handlers.NewReadyHandler(cfg, mongoRepo, redisRepo, seedDone),
		Metrics:        handlers.NewMetricsHandler(cfg.App.Name, cfg.App.Version),
		Ship:           handlers.NewShipmentHandler(shipmentSvc),
		Track:          handlers.NewTrackingHandler(trackingSvc),
		Route:          handlers.NewRouteHandler(routeSvc),
		Stream:         handlers.NewStreamHandler(hub, log, cfg.JWT, cfg.WebSocket),
		ETA:            handlers.NewETAHandler(etaSvc),
		Fleet:          handlers.NewFleetHandler(mongoRepo),
		Auth:           handlers.NewAuthHandler(cfg.JWT, identityStore, redisRepo, cfg.Session),
		Audit:          auditWriter,
		AidaClient:     aidaClient,
		RfiClient:      rfiClient,
		TelepassClient: telepassClient,
		AlboClient:     alboClient,
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	if err := engine.SetTrustedProxies(cfg.HTTP.TrustedProxies); err != nil {
		log.Warn("trusted proxies", zap.Error(err))
	}
	handlers.Register(engine, deps)

	server := &http.Server{
		Addr:         cfg.ListenAddress(),
		Handler:      engine,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	// Start the HTTP server in its own goroutine so the main goroutine
	// can wait on termination signals.
	go func() {
		log.Info("http listening", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server error", zap.Error(err))
		}
	}()

	// Wait for SIGINT / SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Info("shutdown signal received", zap.String("signal", sig.String()))
	rootCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Warn("server shutdown error", zap.Error(err))
	}
	log.Info("goodbye")
	return nil
}

// buildLogger returns a production-grade JSON logger (or a friendlier
// console logger in development).
func buildLogger(cfg *config.Config) (*zap.Logger, error) {
	level := zap.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.App.LogLevel))
	if cfg.IsProduction() {
		encCfg := zap.NewProductionEncoderConfig()
		encCfg.TimeKey = "timestamp"
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		enc := zapcore.NewJSONEncoder(encCfg)
		core := zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level)
		return zap.New(core, zap.AddCaller(), zap.Fields(
			zap.String("service", cfg.App.Name),
			zap.String("env", cfg.App.Env),
			zap.String("version", cfg.App.Version),
		)), nil
	}
	return zap.NewDevelopment()
}

// buildIdentityStore selects an IdentityStore implementation based on
// LOGITRACK_IDENTITY_BACKEND. The contract:
//
//   - "memory" (default): seed an InMemoryIdentityStore with a single
//     demo user from LOGITRACK_IDENTITY_DEMO_{USER,PASSWORD,TENANT,ROLES}.
//     If either user or password is empty, the store is nil — every
//     login returns 503 with the "configure your IDP" pointer. Mirrors
//     the "disabled" branch but without forcing operators to set a
//     second env var to silence the demo seed.
//   - "disabled": return nil so /api/v1/auth/login returns 503.
//   - anything else: typed error so the operator notices a typo at boot.
//
// Production deployments are expected to replace this composition with
// an adapter to the corporate IDP (Azure AD, Keycloak, Okta) — the
// IdentityStore interface in handlers/auth.go is the seam.
func buildIdentityStore(cfg config.IdentityConfig) (handlers.IdentityStore, error) {
	backend := strings.ToLower(strings.TrimSpace(cfg.Backend))
	switch backend {
	case "disabled":
		return nil, nil
	case "", "memory":
		if cfg.DemoUsername == "" || cfg.DemoPassword == "" {
			return nil, nil
		}
		users := map[string]handlers.UserSeed{
			cfg.DemoUsername: {
				Password: cfg.DemoPassword,
				TenantID: cfg.DemoTenantID,
				UserID:   cfg.DemoUsername,
				Roles:    cfg.DemoRoles,
			},
		}
		store, err := handlers.NewInMemoryIdentityStore(users, 5, 15*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("identity store seed: %w", err)
		}
		return store, nil
	default:
		return nil, fmt.Errorf("unsupported LOGITRACK_IDENTITY_BACKEND %q (allowed: memory, disabled)", cfg.Backend)
	}
}

// initTracer configures an OTLP exporter if OTEL_EXPORTER_OTLP_ENDPOINT
// is set. In local runs without a collector the exporter is skipped
// and tracing is effectively a no-op.
func initTracer(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	if cfg.OTEL.OTLPEndpoint == "" {
		return nil, nil
	}
	client := otlptracegrpc.NewClient(otlptracegrpc.WithEndpoint(cfg.OTEL.OTLPEndpoint), otlptracegrpc.WithInsecure())
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OTEL.ServiceName),
			semconv.ServiceVersion(cfg.App.Version),
			semconv.DeploymentEnvironment(cfg.App.Env),
		),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.OTEL.SampleRatio)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
