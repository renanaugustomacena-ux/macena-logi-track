// Package config centralises environment-driven configuration.
//
// The application follows the Twelve-Factor contract: all
// deployment-varying settings are read from the process environment,
// never compiled into the binary. In local development a .env file is
// loaded by godotenv; in production the variables are injected by the
// orchestrator (Kubernetes Secrets, Docker Compose `environment`, or a
// cloud-native secrets manager such as AWS Secrets Manager or HashiCorp
// Vault).
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the fully resolved runtime configuration.
// Fields are grouped by subsystem to keep call-sites terse.
type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Mongo     MongoConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Session   SessionConfig
	OTEL      OTELConfig
	OSRM      OSRMConfig
	AIDA      AIDAConfig
	RFI       RFIConfig
	Telepass  TelepassConfig
	Albo      AlboConfig
	Identity  IdentityConfig
	Kafka     KafkaConfig
	Geofence  GeofenceConfig
	RateLimit RateLimitConfig
	WebSocket WebSocketConfig
	Demo      DemoConfig
}

// AppConfig covers the top-level lifecycle concerns.
type AppConfig struct {
	Env      string `envconfig:"APP_ENV" default:"development"`
	Name     string `envconfig:"APP_NAME" default:"logitrack"`
	Version  string `envconfig:"APP_VERSION" default:"0.1.0"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// HTTPConfig holds the Gin/HTTP listener parameters.
type HTTPConfig struct {
	Port            int           `envconfig:"APP_PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout     time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"30s"`
	CORSAllowed     []string      `envconfig:"HTTP_CORS_ALLOWED" default:"http://localhost:5174"`
	TrustedProxies  []string      `envconfig:"HTTP_TRUSTED_PROXIES" default:"127.0.0.1"`
}

// MongoConfig configures the MongoDB driver.
type MongoConfig struct {
	URI            string        `envconfig:"MONGO_URI" default:"mongodb://localhost:27017"`
	Database       string        `envconfig:"MONGO_DB" default:"logitrack"`
	MaxPoolSize    uint64        `envconfig:"MONGO_MAX_POOL" default:"100"`
	MinPoolSize    uint64        `envconfig:"MONGO_MIN_POOL" default:"5"`
	ConnectTimeout time.Duration `envconfig:"MONGO_CONNECT_TIMEOUT" default:"10s"`
}

// RedisConfig configures the Redis client (cache + pub/sub).
type RedisConfig struct {
	URL      string        `envconfig:"REDIS_URL" default:"redis://localhost:6379/0"`
	PoolSize int           `envconfig:"REDIS_POOL_SIZE" default:"20"`
	Timeout  time.Duration `envconfig:"REDIS_TIMEOUT" default:"3s"`
}

// JWTConfig describes the signing strategy and token TTLs.
type JWTConfig struct {
	Secret     string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTTL  time.Duration `envconfig:"JWT_ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"JWT_REFRESH_TTL" default:"720h"`
	Issuer     string        `envconfig:"JWT_ISSUER" default:"logitrack.it"`
}

// OTELConfig configures the OpenTelemetry OTLP exporter.
type OTELConfig struct {
	ServiceName  string  `envconfig:"OTEL_SERVICE_NAME" default:"logitrack-backend"`
	OTLPEndpoint string  `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:""`
	SampleRatio  float64 `envconfig:"OTEL_SAMPLE_RATIO" default:"0.1"`
}

// OSRMConfig points to the route-optimisation backend.
//
// AllowedHosts is the SSRF guard: outbound OSRM calls are rejected
// unless the host in BaseURL (or any configured override) is on the
// allow-list. See internal/services/route_optimizer.go for the check.
//
// TruckProfile names the OSRM profile to use when a request asks for
// vehicle="truck". The public OSRM service ships only the "driving"
// profile; production deployments serve a custom truck profile (e.g.
// "truck" or "hgv") with weight, height, hazardous-goods and ZTL
// awareness. When TruckProfile is empty, the optimiser falls back to
// the "driving" profile and emits a one-shot WARN so the operator
// knows the route plan ignores HGV restrictions.
type OSRMConfig struct {
	BaseURL      string        `envconfig:"OSRM_BASE_URL" default:"https://router.project-osrm.org"`
	Timeout      time.Duration `envconfig:"OSRM_TIMEOUT" default:"5s"`
	AllowedHosts []string      `envconfig:"OSRM_ALLOWED_HOSTS" default:"router.project-osrm.org,osrm.logitrack.local"`
	CacheSize    int           `envconfig:"OSRM_CACHE_SIZE" default:"1000"`
	TruckProfile string        `envconfig:"OSRM_TRUCK_PROFILE" default:""`
}

// SessionConfig enforces the v2.0 §12 session-timeout rules:
// 15-minute idle window (matching JWT access TTL) and 12-hour absolute
// ceiling on refresh. These values are enforced server-side by the JWT
// middleware — they are NOT advisory fields the client can ignore.
type SessionConfig struct {
	IdleTimeout     time.Duration `envconfig:"SESSION_IDLE_TIMEOUT" default:"15m"`
	AbsoluteTimeout time.Duration `envconfig:"SESSION_ABSOLUTE_TIMEOUT" default:"12h"`
}

// AIDAConfig references the Agenzia delle Dogane customs API.
// LOGITRACK_* env names are the canonical ones documented in the
// remediation brief; the legacy AIDA_* names are kept for backwards
// compatibility during Mission II.5 rollout.
type AIDAConfig struct {
	BaseURL string `envconfig:"LOGITRACK_AIDA_API_BASE" default:""`
	APIKey  string `envconfig:"LOGITRACK_AIDA_API_KEY" default:""`
}

// RFIConfig references the FERTRAM/RFI intermodal rail-slot API.
//
// FERTRAM is mTLS-protected in production. MTLSCertFile / MTLSKeyFile
// are the PEM-encoded client certificate and private key issued by
// RFI to the operator. They must be supplied as a pair: providing
// only one is a construction-time error so the operator does not
// chase a confusing 403 from FERTRAM. MTLSCAFile is optional and
// defaults to the system root pool (which trusts the publicly-signed
// fertram.rfi.it certificate); set it to a private CA bundle when
// FERTRAM presents a private/test certificate.
type RFIConfig struct {
	BaseURL      string `envconfig:"LOGITRACK_RFI_API_BASE" default:""`
	ClientID     string `envconfig:"LOGITRACK_RFI_CLIENT_ID" default:""`
	ClientSecret string `envconfig:"LOGITRACK_RFI_CLIENT_SECRET" default:""`
	MTLSCertFile string `envconfig:"LOGITRACK_RFI_MTLS_CERT_FILE" default:""`
	MTLSKeyFile  string `envconfig:"LOGITRACK_RFI_MTLS_KEY_FILE" default:""`
	MTLSCAFile   string `envconfig:"LOGITRACK_RFI_MTLS_CA_FILE" default:""`
}

// TelepassConfig references the ViaCard/Telepass Business API.
type TelepassConfig struct {
	BaseURL    string `envconfig:"LOGITRACK_TELEPASS_API_BASE" default:""`
	APIKey     string `envconfig:"LOGITRACK_TELEPASS_API_KEY" default:""`
	ContractID string `envconfig:"LOGITRACK_TELEPASS_CONTRACT_ID" default:""`
}

// AlboConfig references the Albo Nazionale degli Autotrasportatori API.
type AlboConfig struct {
	BaseURL string `envconfig:"LOGITRACK_ALBO_API_BASE" default:""`
	APIKey  string `envconfig:"LOGITRACK_ALBO_API_KEY" default:""`
}

// IdentityConfig selects the identity backend. Values:
//
//   - "memory" (default): the in-memory demo store with a single seeded
//     user (username LOGITRACK_IDENTITY_DEMO_USER, password
//     LOGITRACK_IDENTITY_DEMO_PASSWORD). If the password is empty the
//     store ships empty and every login returns invalid_credentials.
//   - "disabled": /api/v1/auth/login returns 503 with a clear pointer
//     to the IDP integration path. Useful in production deployments
//     where an external IDP is the single source of truth.
type IdentityConfig struct {
	Backend      string   `envconfig:"LOGITRACK_IDENTITY_BACKEND" default:"memory"`
	DemoUsername string   `envconfig:"LOGITRACK_IDENTITY_DEMO_USER" default:""`
	DemoPassword string   `envconfig:"LOGITRACK_IDENTITY_DEMO_PASSWORD" default:""`
	DemoTenantID string   `envconfig:"LOGITRACK_IDENTITY_DEMO_TENANT" default:"demo-tenant"`
	DemoRoles    []string `envconfig:"LOGITRACK_IDENTITY_DEMO_ROLES" default:"operator"`
}

// KafkaConfig is a placeholder for future event-streaming roll-out.
type KafkaConfig struct {
	Brokers []string `envconfig:"KAFKA_BROKERS" default:""`
	Topic   string   `envconfig:"KAFKA_TOPIC" default:"logitrack.events.v1"`
}

// GeofenceConfig tunes the spatial-index resolution.
type GeofenceConfig struct {
	TileResolution int `envconfig:"GEOFENCE_TILE_RESOLUTION" default:"9"`
}

// RateLimitConfig controls the basic token-bucket limiter.
type RateLimitConfig struct {
	RPS   int `envconfig:"RATE_LIMIT_RPS" default:"30"`
	Burst int `envconfig:"RATE_LIMIT_BURST" default:"60"`
}

// WebSocketConfig tunes the live-tracking WS endpoint.
//
// AllowedOrigins is the hard origin-check performed during the
// handshake: any request whose Origin header is not on this list is
// rejected with 403 before the connection is promoted. Empty list =>
// accept only same-origin requests (no Origin header, or Origin ==
// request host).
type WebSocketConfig struct {
	AllowedOrigins []string      `envconfig:"HTTP_WS_ORIGINS" default:"http://localhost:5174,http://127.0.0.1:5174"`
	RateLimitRPS   int           `envconfig:"WS_RATE_LIMIT_RPS" default:"20"`
	RateLimitBurst int           `envconfig:"WS_RATE_LIMIT_BURST" default:"40"`
	IdleTimeout    time.Duration `envconfig:"WS_IDLE_TIMEOUT" default:"5m"`
	HandshakeSub   string        `envconfig:"WS_JWT_SUBPROTOCOL" default:"logitrack.jwt.v1"`
}

// DemoConfig controls the on-boot seed used by the golden-path demo.
type DemoConfig struct {
	SeedOnBoot bool   `envconfig:"SEED_DEMO" default:"false"`
	TenantID   string `envconfig:"SEED_DEMO_TENANT" default:"demo-tenant"`
}

// Load reads configuration from the environment, optionally pre-loading
// values from a .env file in the working directory. Missing values fall
// back to the `default` tags above; required values missing will return
// a typed error for the caller to log-and-exit cleanly.
func Load() (*Config, error) {
	// godotenv is a best-effort loader: absence of a .env file is not
	// an error (production deployments inject env directly).
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("config: envconfig process: %w", err)
	}
	if err := cfg.guardProductionSecrets(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// guardProductionSecrets refuses to boot a production process with
// demo-quality secrets. The rule closes RA-002 (default JWT secret)
// and makes the "sellable product" contract enforceable at startup
// rather than at audit time.
func (c *Config) guardProductionSecrets() error {
	if !c.IsProduction() {
		return nil
	}
	weak := []string{
		"",
		"change-me",
		"change-me-in-production-use-openssl-rand-hex-32",
		"dev",
		"test",
	}
	for _, w := range weak {
		if strings.EqualFold(c.JWT.Secret, w) {
			return fmt.Errorf("config: JWT_SECRET is a known-weak placeholder (%q); "+
				"generate a 32-byte secret with `openssl rand -hex 32` and inject via secrets manager", w)
		}
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("config: JWT_SECRET must be >= 32 characters in production, got %d", len(c.JWT.Secret))
	}
	return nil
}

// IsProduction reports whether the application is running in a
// production-class environment, used to toggle verbose logging and
// Gin's debug middleware.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

// ListenAddress returns the Go-standard listener string.
func (c *Config) ListenAddress() string {
	return fmt.Sprintf(":%d", c.HTTP.Port)
}
