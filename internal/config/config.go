package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Typesense TypesenseConfig `mapstructure:"typesense"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Log       LogConfig       `mapstructure:"log"`
}

type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
}

type HTTPConfig struct {
	Port int       `mapstructure:"port"`
	TLS  TLSConfig `mapstructure:"tls"`
}

// TLSConfig holds TLS/HTTP3 certificate paths.
// When both CertFile and KeyFile are set, the server starts with HTTP/3 (QUIC).
type TLSConfig struct {
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// Enabled reports whether TLS is configured.
func (t TLSConfig) Enabled() bool {
	return t.CertFile != "" && t.KeyFile != ""
}

type DatabaseConfig struct {
	Host     string     `mapstructure:"host"`
	Port     int        `mapstructure:"port"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	Name     string     `mapstructure:"name"`
	SSLMode  string     `mapstructure:"ssl_mode"`
	Pool     PoolConfig `mapstructure:"pool"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type PoolConfig struct {
	MaxConns          int           `mapstructure:"max_conns"`
	MinConns          int           `mapstructure:"min_conns"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

type TypesenseConfig struct {
	Host              string               `mapstructure:"host"`
	Port              int                  `mapstructure:"port"`
	APIKey            string               `mapstructure:"api_key"`
	ConnectionTimeout time.Duration        `mapstructure:"connection_timeout"`
	CircuitBreaker    CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

func (t TypesenseConfig) URL() string {
	return fmt.Sprintf("http://%s:%d", t.Host, t.Port)
}

type CircuitBreakerConfig struct {
	MaxRequests uint32        `mapstructure:"max_requests"`
	Interval    time.Duration `mapstructure:"interval"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type OpenAIConfig struct {
	APIKey         string `mapstructure:"api_key"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	ChatModel      string `mapstructure:"chat_model"`
}

type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads the configuration from the given path and environment variables.
// Environment variables are prefixed with MAHZEN_ and use underscore as separator.
// For example, MAHZEN_DATABASE_HOST overrides database.host.
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvPrefix("MAHZEN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.HTTP.Port == 0 {
		return fmt.Errorf("server.http.port is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if c.Typesense.Host == "" {
		return fmt.Errorf("typesense.host is required")
	}
	if c.Typesense.APIKey == "" {
		return fmt.Errorf("typesense.api_key is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	return nil
}
