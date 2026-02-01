package config

import (
	"testing"
	"time"
)

func TestServerConfig_Address(t *testing.T) {
	tests := []struct {
		name     string
		config   ServerConfig
		expected string
	}{
		{
			name:     "default",
			config:   ServerConfig{Host: "0.0.0.0", Port: 8080},
			expected: "0.0.0.0:8080",
		},
		{
			name:     "localhost",
			config:   ServerConfig{Host: "localhost", Port: 3000},
			expected: "localhost:3000",
		},
		{
			name:     "custom host",
			config:   ServerConfig{Host: "192.168.1.1", Port: 9000},
			expected: "192.168.1.1:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Address()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "from URL",
			config: DatabaseConfig{
				URL: "postgres://user:pass@host:5432/db",
			},
			expected: "postgres://user:pass@host:5432/db",
		},
		{
			name: "from fields",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "secret",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=disable",
		},
		{
			name: "url takes precedence",
			config: DatabaseConfig{
				URL:      "postgres://priority@host:5432/db",
				Host:     "ignored",
				Port:     1234,
				User:     "ignored",
				Password: "ignored",
				Name:     "ignored",
			},
			expected: "postgres://priority@host:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.DSN()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRedisConfig_Address(t *testing.T) {
	tests := []struct {
		name     string
		config   RedisConfig
		expected string
	}{
		{
			name:     "default",
			config:   RedisConfig{Host: "localhost", Port: 6379},
			expected: "localhost:6379",
		},
		{
			name:     "custom",
			config:   RedisConfig{Host: "redis.example.com", Port: 6380},
			expected: "redis.example.com:6380",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Address()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "secret",
			Name:     "swarmmarket",
			SSLMode:  "disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Auth: AuthConfig{
			APIKeyHeader:   "X-API-Key",
			APIKeyLength:   32,
			RateLimitRPS:   100,
			RateLimitBurst: 200,
			TokenTTL:       24 * time.Hour,
		},
		Stripe: StripeConfig{
			SecretKey:          "sk_test_xxx",
			WebhookSecret:      "whsec_xxx",
			PlatformFeePercent: 0.025,
		},
		Clerk: ClerkConfig{
			PublishableKey: "pk_test_xxx",
			SecretKey:      "sk_test_xxx",
		},
	}

	// Verify server config
	if cfg.Server.Host != "0.0.0.0" {
		t.Error("server host not set correctly")
	}
	if cfg.Server.Port != 8080 {
		t.Error("server port not set correctly")
	}
	if cfg.Server.Address() != "0.0.0.0:8080" {
		t.Error("server address not computed correctly")
	}

	// Verify database config
	if cfg.Database.Host != "localhost" {
		t.Error("database host not set correctly")
	}
	if cfg.Database.MaxConns != 25 {
		t.Error("database max conns not set correctly")
	}

	// Verify redis config
	if cfg.Redis.Host != "localhost" {
		t.Error("redis host not set correctly")
	}
	if cfg.Redis.Address() != "localhost:6379" {
		t.Error("redis address not computed correctly")
	}

	// Verify auth config
	if cfg.Auth.APIKeyHeader != "X-API-Key" {
		t.Error("auth api key header not set correctly")
	}
	if cfg.Auth.RateLimitRPS != 100 {
		t.Error("auth rate limit not set correctly")
	}

	// Verify stripe config
	if cfg.Stripe.PlatformFeePercent != 0.025 {
		t.Error("stripe platform fee not set correctly")
	}

	// Verify clerk config
	if cfg.Clerk.PublishableKey != "pk_test_xxx" {
		t.Error("clerk publishable key not set correctly")
	}
}

func TestServerConfig_Defaults(t *testing.T) {
	cfg := ServerConfig{}

	// Check that zero values work
	if cfg.Address() != ":0" {
		t.Errorf("expected :0 for empty config, got %s", cfg.Address())
	}
}

func TestDatabaseConfig_EmptyURL(t *testing.T) {
	cfg := DatabaseConfig{
		URL:      "",
		Host:     "myhost",
		Port:     5433,
		User:     "myuser",
		Password: "mypass",
		Name:     "mydb",
		SSLMode:  "require",
	}

	dsn := cfg.DSN()
	expected := "host=myhost port=5433 user=myuser password=mypass dbname=mydb sslmode=require"

	if dsn != expected {
		t.Errorf("expected %s, got %s", expected, dsn)
	}
}

func TestRedisConfig_Defaults(t *testing.T) {
	cfg := RedisConfig{}

	// Zero values
	if cfg.Address() != ":0" {
		t.Errorf("expected :0 for empty config, got %s", cfg.Address())
	}
}

func TestAuthConfig_Struct(t *testing.T) {
	cfg := AuthConfig{
		APIKeyHeader:   "Authorization",
		APIKeyLength:   64,
		RateLimitRPS:   50,
		RateLimitBurst: 100,
		TokenTTL:       12 * time.Hour,
	}

	if cfg.APIKeyHeader != "Authorization" {
		t.Error("api key header not set correctly")
	}
	if cfg.APIKeyLength != 64 {
		t.Error("api key length not set correctly")
	}
	if cfg.TokenTTL != 12*time.Hour {
		t.Error("token ttl not set correctly")
	}
}

func TestStripeConfig_Struct(t *testing.T) {
	cfg := StripeConfig{
		SecretKey:          "sk_live_xxx",
		WebhookSecret:      "whsec_live_xxx",
		PlatformFeePercent: 0.03, // 3%
	}

	if cfg.SecretKey != "sk_live_xxx" {
		t.Error("secret key not set correctly")
	}
	if cfg.PlatformFeePercent != 0.03 {
		t.Error("platform fee not set correctly")
	}
}

func TestClerkConfig_Struct(t *testing.T) {
	cfg := ClerkConfig{
		PublishableKey: "pk_live_xxx",
		SecretKey:      "sk_live_xxx",
	}

	if cfg.PublishableKey != "pk_live_xxx" {
		t.Error("publishable key not set correctly")
	}
	if cfg.SecretKey != "sk_live_xxx" {
		t.Error("secret key not set correctly")
	}
}

func TestDatabaseConfig_ConnectionParams(t *testing.T) {
	cfg := DatabaseConfig{
		MaxConns:    50,
		MinConns:    10,
		MaxConnLife: 2 * time.Hour,
		MaxConnIdle: 1 * time.Hour,
	}

	if cfg.MaxConns != 50 {
		t.Errorf("expected max conns 50, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 10 {
		t.Errorf("expected min conns 10, got %d", cfg.MinConns)
	}
	if cfg.MaxConnLife != 2*time.Hour {
		t.Error("max conn life not set correctly")
	}
	if cfg.MaxConnIdle != 1*time.Hour {
		t.Error("max conn idle not set correctly")
	}
}

func TestLoad(t *testing.T) {
	// This test just verifies Load doesn't panic
	// Actual values depend on environment
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}
