package config

import (
	"fmt"
	"time"
)

// ConnectorConfig contains the configuration for a connector.
type ConnectorConfig struct {
	// ID is the unique identifier of the connector.
	ID string `json:"id" yaml:"id"`

	// Name is the display name of the connector.
	Name string `json:"name" yaml:"name"`

	// Type is the connector type.
	Type string `json:"type" yaml:"type"`

	// Enabled indicates if the connector is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// NetworkConfig contains network-specific configuration.
	NetworkConfig NetworkConfig `json:"network" yaml:"network"`

	// AuthConfig contains authentication configuration.
	AuthConfig AuthConfig `json:"auth" yaml:"auth"`

	// TimeoutConfig contains timeout configuration.
	TimeoutConfig TimeoutConfig `json:"timeout" yaml:"timeout"`

	// RetryConfig contains retry configuration.
	RetryConfig RetryConfig `json:"retry" yaml:"retry"`

	// RateLimitConfig contains rate limit configuration.
	RateLimitConfig RateLimitConfig `json:"rateLimit" yaml:"rateLimit"`

	// TLSConfig contains TLS configuration.
	TLSConfig TLSConfig `json:"tls" yaml:"tls"`

	// LogConfig contains logging configuration.
	LogConfig LogConfig `json:"log" yaml:"log"`

	// Metadata contains additional configuration.
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// NetworkConfig contains network configuration.
type NetworkConfig struct {
	// NetworkID is the network identifier.
	NetworkID string `json:"networkId" yaml:"networkId"`

	// NetworkName is the network name.
	NetworkName string `json:"networkName" yaml:"networkName"`

	// Endpoint is the network endpoint URL.
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Timeout is the network timeout.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// MaxConnections is the maximum number of connections.
	MaxConnections int `json:"maxConnections" yaml:"maxConnections"`

	// IdleTimeout is the idle connection timeout.
	IdleTimeout time.Duration `json:"idleTimeout" yaml:"idleTimeout"`

	// KeepAlive enables keep-alive.
	KeepAlive bool `json:"keepAlive" yaml:"keepAlive"`

	// Extra contains additional network configuration.
	Extra map[string]interface{} `json:"extra" yaml:"extra"`
}

// AuthConfig contains authentication configuration.
type AuthConfig struct {
	// Type is the authentication type.
	Type string `json:"type" yaml:"type"` // "api_key", "oauth2", "basic", etc.

	// APIKey is the API key for API key authentication.
	APIKey string `json:"apiKey" yaml:"apiKey"`

	// Secret is the API secret.
	Secret string `json:"secret" yaml:"secret"`

	// TokenEndpoint is the token endpoint for OAuth2.
	TokenEndpoint string `json:"tokenEndpoint" yaml:"tokenEndpoint"`

	// ClientID is the client ID for OAuth2.
	ClientID string `json:"clientId" yaml:"clientId"`

	// ClientSecret is the client secret for OAuth2.
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`

	// Scopes is the OAuth2 scopes.
	Scopes []string `json:"scopes" yaml:"scopes"`

	// RefreshToken is the refresh token.
	RefreshToken string `json:"refreshToken" yaml:"refreshToken"`

	// Extra contains additional authentication configuration.
	Extra map[string]interface{} `json:"extra" yaml:"extra"`
}

// TimeoutConfig contains timeout configuration.
type TimeoutConfig struct {
	// Connect is the connection timeout.
	Connect time.Duration `json:"connect" yaml:"connect"`

	// Read is the read timeout.
	Read time.Duration `json:"read" yaml:"read"`

	// Write is the write timeout.
	Write time.Duration `json:"write" yaml:"write"`

	// Request is the request timeout.
	Request time.Duration `json:"request" yaml:"request"`

	// Transaction is the transaction timeout.
	Transaction time.Duration `json:"transaction" yaml:"transaction"`

	// Idle is the idle timeout.
	Idle time.Duration `json:"idle" yaml:"idle"`
}

// RetryConfig contains retry configuration.
type RetryConfig struct {
	// Enabled indicates if retries are enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int `json:"maxAttempts" yaml:"maxAttempts"`

	// InitialDelay is the initial retry delay.
	InitialDelay time.Duration `json:"initialDelay" yaml:"initialDelay"`

	// MaxDelay is the maximum retry delay.
	MaxDelay time.Duration `json:"maxDelay" yaml:"maxDelay"`

	// Multiplier is the exponential backoff multiplier.
	Multiplier float64 `json:"multiplier" yaml:"multiplier"`

	// RetryableErrors are the error codes that trigger retries.
	RetryableErrors []string `json:"retryableErrors" yaml:"retryableErrors"`
}

// RateLimitConfig contains rate limit configuration.
type RateLimitConfig struct {
	// Enabled indicates if rate limiting is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RequestsPerSecond is the maximum requests per second.
	RequestsPerSecond int `json:"requestsPerSecond" yaml:"requestsPerSecond"`

	// BurstSize is the burst size.
	BurstSize int `json:"burstSize" yaml:"burstSize"`
}

// TLSConfig contains TLS configuration.
type TLSConfig struct {
	// Enabled indicates if TLS is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// InsecureSkipVerify skips certificate verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`

	// CertificateFile is the certificate file path.
	CertificateFile string `json:"certificateFile" yaml:"certificateFile"`

	// KeyFile is the key file path.
	KeyFile string `json:"keyFile" yaml:"keyFile"`

	// CAFile is the CA certificate file path.
	CAFile string `json:"caFile" yaml:"caFile"`

	// ServerName is the server name for SNI.
	ServerName string `json:"serverName" yaml:"serverName"`
}

// LogConfig contains logging configuration.
type LogConfig struct {
	// Level is the log level.
	Level string `json:"level" yaml:"level"`

	// Format is the log format.
	Format string `json:"format" yaml:"format"` // "json", "console"

	// SensitiveDataMasking masks sensitive data in logs.
	SensitiveDataMasking bool `json:"sensitiveDataMasking" yaml:"sensitiveDataMasking"`
}

// Validate validates the connector configuration.
func (c *ConnectorConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("connector ID is required")
	}

	if c.Name == "" {
		return fmt.Errorf("connector name is required")
	}

	if c.Type == "" {
		return fmt.Errorf("connector type is required")
	}

	if c.NetworkConfig.Endpoint == "" {
		return fmt.Errorf("network endpoint is required")
	}

	return nil
}
