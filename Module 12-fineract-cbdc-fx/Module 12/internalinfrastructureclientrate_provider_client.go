package client

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-fx/internal/domain/rate"
    "github.com/apache/fineract-cbdc-fx/internal/infrastructure/config"

    "github.com/shopspring/decimal"
    "github.com/sony/gobreaker"
    "go.uber.org/zap"
)

// RateProvider defines the interface for rate providers
type RateProvider interface {
    GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error)
    GetSupportedPairs(ctx context.Context) ([]CurrencyPair, error)
    HealthCheck(ctx context.Context) error
}

// CurrencyPair represents a currency pair
type CurrencyPair struct {
    Base  string
    Quote string
}

// RateProviderClient implements the RateProvider interface
type RateProviderClient struct {
    config      *config.RateProviderConfig
    httpClient  *http.Client
    cb          *gobreaker.CircuitBreaker
    logger      *zap.Logger
    providerURL string
    apiKey      string
}

// NewRateProviderClient creates a new rate provider client
func NewRateProviderClient(cfg *config.RateProviderConfig, logger *zap.Logger) *RateProviderClient {
    // Initialize circuit breaker
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "rate-provider",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures > 5
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            logger.Info("Circuit breaker state changed",
                zap.String("from", from.String()),
                zap.String("to", to.String()),
            )
        },
    })

    client := &http.Client{
        Timeout: cfg.Timeout,
        Transport: &http.Transport{
            MaxIdleConns:    10,
            IdleConnTimeout: 30 * time.Second,
        },
    }

    return &RateProviderClient{
        config:      cfg,
        httpClient:  client,
        cb:          cb,
        logger:      logger,
        providerURL: cfg.URL,
        apiKey:      cfg.APIKey,
    }
}

// GetRate fetches the exchange rate from the provider
func (c *RateProviderClient) GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error) {
    // Execute with circuit breaker
    result, err := c.cb.Execute(func() (interface{}, error) {
        return c.doGetRate(ctx, baseCurrency, quoteCurrency)
    })

    if err != nil {
        return nil, err
    }

    return result.(*rate.ExchangeRate), nil
}

// doGetRate performs the actual HTTP request
func (c *RateProviderClient) doGetRate(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error) {
    url := fmt.Sprintf("%s/latest?base=%s&symbols=%s&apikey=%s",
        c.providerURL, baseCurrency, quoteCurrency, c.apiKey)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Add headers
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", "fineract-cbdc-fx/1.0")

    // Execute request with retry
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
    }

    // Parse response
    var response struct {
        Success bool    `json:"success"`
        Rates   map[string]string `json:"rates"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    if !response.Success {
        return nil, fmt.Errorf("provider returned error")
    }

    // Extract rate
    rateStr, exists := response.Rates[quoteCurrency]
    if !exists {
        return nil, fmt.Errorf("rate not found for %s/%s", baseCurrency, quoteCurrency)
    }

    rateValue, err := decimal.NewFromString(rateStr)
    if err != nil {
        return nil, fmt.Errorf("invalid rate value: %s", rateStr)
    }

    // Create exchange rate object
    exchangeRate := &rate.ExchangeRate{
        ID:            uuid.New().String(),
        BaseCurrency:  baseCurrency,
        QuoteCurrency: quoteCurrency,
        BidRate:       rateValue,
        AskRate:       rateValue,
        MidRate:       rateValue,
        Spread:        decimal.Zero,
        Provider:      "provider",
        Status:        rate.RateStatusActive,
        Timestamp:     time.Now(),
        ExpiresAt:     time.Now().Add(c.config.CacheDuration),
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    return exchangeRate, nil
}

// GetSupportedPairs returns the supported currency pairs
func (c *RateProviderClient) GetSupportedPairs(ctx context.Context) ([]CurrencyPair, error) {
    // For now, return a static list
    // In production, this would fetch from the provider
    return []CurrencyPair{
        {Base: "USD", Quote: "INR"},
        {Base: "USD", Quote: "AED"},
        {Base: "INR", Quote: "AED"},
        {Base: "AED", Quote: "INR"},
        {Base: "EUR", Quote: "USD"},
        {Base: "GBP", Quote: "USD"},
        {Base: "USD", Quote: "EUR"},
        {Base: "USD", Quote: "GBP"},
    }, nil
}

// HealthCheck checks the health of the provider
func (c *RateProviderClient) HealthCheck(ctx context.Context) error {
    url := fmt.Sprintf("%s/health", c.providerURL)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check failed: %d", resp.StatusCode)
    }

    return nil
}