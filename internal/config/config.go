package config

import (
	"cmp"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Meilisearch MeilisearchConfig `mapstructure:"meilisearch"`
	OpenAI      OpenAIConfig      `mapstructure:"openai"`
	Auth        AuthConfig        `mapstructure:"auth"`
	DefaultUser DefaultUserConfig `mapstructure:"default_user"`
	Log         LogConfig         `mapstructure:"log"`
}

type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
}

type HTTPConfig struct {
	Port int       `mapstructure:"port"`
	TLS  TLSConfig `mapstructure:"tls"`
}

// TLSConfig enables TLS + HTTP/3 when both cert paths are set.
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

type MeilisearchConfig struct {
	Host   string `mapstructure:"host"` // e.g. "localhost:7700"
	APIKey string `mapstructure:"api_key"`
}

type OpenAIConfig struct {
	APIKey         string `mapstructure:"api_key"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	ChatModel      string `mapstructure:"chat_model"`
}

type AuthConfig struct {
	JWTSecret                string        `mapstructure:"jwt_secret"`
	AccessTokenExpiry        time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry       time.Duration `mapstructure:"refresh_token_expiry"`
	AccessTokenDefaultExpiry time.Duration `mapstructure:"access_token_default_expiry"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DefaultUserConfig holds the bootstrap user created on startup.
// The password is only applied while it is still the default —
// once changed (or absent), the existing account is left untouched.
type DefaultUserConfig struct {
	Username    string `mapstructure:"username"`
	Email       string `mapstructure:"email"`
	Password    string `mapstructure:"password"`
	DisplayName string `mapstructure:"display_name"`
}

// UsernameOrDefault returns the configured username or the built-in default.
func (d DefaultUserConfig) UsernameOrDefault() string {
	return cmp.Or(d.Username, "admin")
}

// EmailOrDefault returns the configured email or the built-in default.
func (d DefaultUserConfig) EmailOrDefault() string {
	return cmp.Or(d.Email, "admin@mahzen.local")
}

// PasswordOrDefault returns the configured password or the built-in default.
func (d DefaultUserConfig) PasswordOrDefault() string {
	return cmp.Or(d.Password, "mahzen")
}

// DisplayNameOrDefault returns the configured display name or the built-in default.
func (d DefaultUserConfig) DisplayNameOrDefault() string {
	return cmp.Or(d.DisplayName, "Admin")
}

// Load reads config from the given file path; MAHZEN_* env vars override
// values (e.g. MAHZEN_DATABASE_HOST overrides database.host).
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
	if c.Meilisearch.Host == "" {
		return fmt.Errorf("meilisearch.host is required")
	}
	if c.Meilisearch.APIKey == "" {
		return fmt.Errorf("meilisearch.api_key is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	return nil
}
