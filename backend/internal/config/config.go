package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Security SecurityConfig
	Stripe   StripeConfig
	Clerk    ClerkConfig
	Twitter  TwitterConfig
	Trust    TrustConfig
	Storage  StorageConfig
	Email    EmailConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string        `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
	PublicURL    string        `envconfig:"PUBLIC_URL" default:"https://swarmmarket.ai"` // Public URL for sitemap
}

// Address returns the server address string.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	URL          string        `envconfig:"DATABASE_URL"`
	Host         string        `envconfig:"DB_HOST" default:"localhost"`
	Port         int           `envconfig:"DB_PORT" default:"5432"`
	User         string        `envconfig:"DB_USER" default:"postgres"`
	Password     string        `envconfig:"DB_PASSWORD" default:""`
	Name         string        `envconfig:"DB_NAME" default:"postgres"`
	SSLMode      string        `envconfig:"DB_SSL_MODE" default:"disable"`
	MaxConns     int           `envconfig:"DB_MAX_CONNS" default:"25"`
	MinConns     int           `envconfig:"DB_MIN_CONNS" default:"5"`
	MaxConnLife  time.Duration `envconfig:"DB_MAX_CONN_LIFE" default:"1h"`
	MaxConnIdle  time.Duration `envconfig:"DB_MAX_CONN_IDLE" default:"30m"`
}

// DSN returns the PostgreSQL connection string.
// If DATABASE_URL is set, it takes precedence over individual fields.
func (d DatabaseConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	URL      string `envconfig:"REDIS_URL"`
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

// Address returns the Redis address string (host:port).
func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	APIKeyHeader   string        `envconfig:"AUTH_API_KEY_HEADER" default:"X-API-Key"`
	APIKeyLength   int           `envconfig:"AUTH_API_KEY_LENGTH" default:"32"`
	RateLimitRPS   int           `envconfig:"AUTH_RATE_LIMIT_RPS" default:"100"`
	RateLimitBurst int           `envconfig:"AUTH_RATE_LIMIT_BURST" default:"200"`
	TokenTTL       time.Duration `envconfig:"AUTH_TOKEN_TTL" default:"24h"`
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:5173,http://localhost:3000"` // Comma-separated list
	MaxRequestBodySize int64  `envconfig:"MAX_REQUEST_BODY_SIZE" default:"10485760"`                                   // 10MB default
}

// StripeConfig holds Stripe payment configuration.
type StripeConfig struct {
	SecretKey          string  `envconfig:"STRIPE_SECRET_KEY" default:""`
	WebhookSecret      string  `envconfig:"STRIPE_WEBHOOK_SECRET" default:""`
	PlatformFeePercent float64 `envconfig:"STRIPE_PLATFORM_FEE_PERCENT" default:"0.025"` // 2.5%
	DefaultReturnURL   string  `envconfig:"STRIPE_DEFAULT_RETURN_URL" default:""`        // URL for redirect after payment confirmation
}

// ClerkConfig holds Clerk authentication configuration.
type ClerkConfig struct {
	PublishableKey string `envconfig:"CLERK_PUBLISHABLE_KEY" default:""`
	SecretKey      string `envconfig:"CLERK_SECRET_KEY" default:""`
}

// TwitterConfig holds Twitter API configuration for verification.
type TwitterConfig struct {
	BearerToken string `envconfig:"TWITTER_BEARER_TOKEN" default:""`
}

// TrustConfig holds trust system configuration.
// Trust score: 0-100% (stored as 0.0-1.0)
// New agents start at 0%, max is 100%
type TrustConfig struct {
	HumanLinkBonus       float64 `envconfig:"TRUST_HUMAN_LINK_BONUS" default:"0.10"`       // +10% for linking to human
	TwitterTrustBonus    float64 `envconfig:"TRUST_TWITTER_BONUS" default:"0.15"`          // +15% for Twitter verification
	MaxTransactionBonus  float64 `envconfig:"TRUST_MAX_TRANSACTION_BONUS" default:"0.75"`  // up to +75% from trades
	TransactionDecayRate float64 `envconfig:"TRUST_TRANSACTION_DECAY_RATE" default:"0.03"` // Exponential decay rate
}

// StorageConfig holds object storage configuration (Cloudflare R2).
type StorageConfig struct {
	R2AccountID       string `envconfig:"R2_ACCOUNT_ID" default:""`
	R2AccessKeyID     string `envconfig:"R2_ACCESS_KEY_ID" default:""`
	R2SecretAccessKey string `envconfig:"R2_SECRET_ACCESS_KEY" default:""`
	R2BucketName      string `envconfig:"R2_BUCKET_NAME" default:"swarmmarket-images"`
	R2PublicURL       string `envconfig:"R2_PUBLIC_URL" default:""` // Custom domain or R2.dev URL
	MaxFileSizeMB     int    `envconfig:"STORAGE_MAX_FILE_SIZE_MB" default:"10"`
}

// EmailConfig holds email service configuration (SendGrid).
type EmailConfig struct {
	SendGridAPIKey  string `envconfig:"SENDGRID_API_KEY" default:""`
	FromEmail       string `envconfig:"EMAIL_FROM" default:"noreply@swarmmarket.ai"`
	FromName        string `envconfig:"EMAIL_FROM_NAME" default:"SwarmMarket"`
	CooldownMinutes int    `envconfig:"EMAIL_COOLDOWN_MINUTES" default:"5"` // Min time between emails to same recipient
}

// Load reads configuration from environment variables.
// It first attempts to load a .env file if present.
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

// MustLoad reads configuration and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
