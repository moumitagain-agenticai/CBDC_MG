package config

import (
    "fmt"
    "os"
    "time"

    "github.com/joho/godotenv"
    "gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
    Version     string           `yaml:"version"`
    Environment string           `yaml:"environment"`
    API         APIConfig        `yaml:"api"`
    Database    DatabaseConfig   `yaml:"database"`
    CBDC        CBDCConfig       `yaml:"cbdc"`
    Fineract    FineractConfig   `yaml:"fineract"`
    Log         LogConfig        `yaml:"log"`
    Metrics     MetricsConfig    `yaml:"metrics"`
    Retry       RetryConfig      `yaml:"retry"`
    CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
}

// APIConfig represents API server configuration
type APIConfig struct {
    Port            int           `yaml:"port"`
    ReadTimeout     time.Duration `yaml:"read_timeout"`
    WriteTimeout    time.Duration `yaml:"write_timeout"`
    IdleTimeout     time.Duration `yaml:"idle_timeout"`
    ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
    RateLimit       struct {
        Enabled bool `yaml:"enabled"`
        RPM     int  `yaml:"rpm"`
        Burst   int  `yaml:"burst"`
    } `yaml:"rate_limit"`
    Cors struct {
        Enabled     bool     `yaml:"enabled"`
        AllowedOrigins []string `yaml:"allowed_origins"`
        AllowedMethods []string `yaml:"allowed_methods"`
        AllowedHeaders []string `yaml:"allowed_headers"`
    } `yaml:"cors"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
    Driver          string `yaml:"driver"`
    Host            string `yaml:"host"`
    Port            int    `yaml:"port"`
    User            string `yaml:"user"`
    Password        string `yaml:"password"`
    DBName          string `yaml:"dbname"`
    SSLMode         string `yaml:"ssl_mode"`
    MaxOpenConns    int    `yaml:"max_open_conns"`
    MaxIdleConns    int    `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// CBDCConfig represents Indian CBDC (e₹) API configuration
type CBDCConfig struct {
    BaseURL       string        `yaml:"base_url"`
    APIKey        string        `yaml:"api_key"`
    APISecret     string        `yaml:"api_secret"`
    Timeout       time.Duration `yaml:"timeout"`
    SponsorBankID string        `yaml:"sponsor_bank_id"`
    MerchantID    string        `yaml:"merchant_id"`
    TerminalID    string        `yaml:"terminal_id"`
    Environment   string        `yaml:"environment"` // "sandbox", "pilot", "production"
    Endpoints     struct {
        Issue     string `yaml:"issue"`
        Transfer  string `yaml:"transfer"`
        Lock      string `yaml:"lock"`
        Burn      string `yaml:"burn"`
        Redeem    string `yaml:"redeem"`
        Balance   string `yaml:"balance"`
        Status    string `yaml:"status"`
        Health    string `yaml:"health"`
    } `yaml:"endpoints"`
}

// FineractConfig represents Apache Fineract core configuration
type FineractConfig struct {
    BaseURL      string        `yaml:"base_url"`
    TenantID     string        `yaml:"tenant_id"`
    Username     string        `yaml:"username"`
    Password     string        `yaml:"password"`
    Token        string        `yaml:"token"`
    Timeout      time.Duration `yaml:"timeout"`
    Endpoints    struct {
        Transactions string `yaml:"transactions"`
        Accounts     string `yaml:"accounts"`
        Customers    string `yaml:"customers"`
        Health       string `yaml:"health"`
    } `yaml:"endpoints"`
}

// LogConfig represents logging configuration
type LogConfig struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"` // "json" or "console"
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
    Enabled bool   `yaml:"enabled"`
    Path    string `yaml:"path"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
    MaxAttempts      int           `yaml:"max_attempts"`
    InitialDelay     time.Duration `yaml:"initial_delay"`
    MaxDelay         time.Duration `yaml:"max_delay"`
    Multiplier       float64       `yaml:"multiplier"`
    RetryableErrors  []string      `yaml:"retryable_errors"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
    Enabled          bool          `yaml:"enabled"`
    MaxRequests      uint32        `yaml:"max_requests"`
    Interval         time.Duration `yaml:"interval"`
    Timeout          time.Duration `yaml:"timeout"`
    FailureThreshold float64       `yaml:"failure_threshold"`
    SuccessThreshold uint32        `yaml:"success_threshold"`
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
    // Load .env file if exists
    _ = godotenv.Load()

    configFile := os.Getenv("CONFIG_FILE")
    if configFile == "" {
        configFile = "configs/config.yaml"
    }

    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    // Override with environment variables
    if v := os.Getenv("CBDC_API_KEY"); v != "" {
        cfg.CBDC.APIKey = v
    }
    if v := os.Getenv("CBDC_API_SECRET"); v != "" {
        cfg.CBDC.APISecret = v
    }
    if v := os.Getenv("FINERACT_TOKEN"); v != "" {
        cfg.Fineract.Token = v
    }
    if v := os.Getenv("DB_PASSWORD"); v != "" {
        cfg.Database.Password = v
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
    if c.CBDC.BaseURL == "" {
        return fmt.Errorf("CBDC base URL is required")
    }
    if c.CBDC.APIKey == "" {
        return fmt.Errorf("CBDC API key is required")
    }
    if c.CBDC.APISecret == "" {
        return fmt.Errorf("CBDC API secret is required")
    }
    if c.Fineract.BaseURL == "" {
        return fmt.Errorf("Fineract base URL is required")
    }
    if c.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }
    if c.Database.DBName == "" {
        return fmt.Errorf("database name is required")
    }
    return nil
}