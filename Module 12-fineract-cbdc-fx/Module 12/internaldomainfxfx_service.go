package fx

import (
    "context"
    "time"

    "github.com/apache/fineract-cbdc-fx/internal/domain/conversion"
    "github.com/apache/fineract-cbdc-fx/internal/domain/quote"
    "github.com/apache/fineract-cbdc-fx/internal/domain/rate"

    "github.com/shopspring/decimal"
)

// Service defines the FX service interface
type Service interface {
    // Rate Operations
    GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error)
    RefreshRates(ctx context.Context) error
    GetHistoricalRates(ctx context.Context, baseCurrency, quoteCurrency string, from, to time.Time) ([]*rate.ExchangeRate, error)
    GetRateProviders(ctx context.Context) ([]string, error)

    // Quote Operations
    GetQuote(ctx context.Context, req *QuoteRequest) (*quote.FXQuote, error)
    LockQuote(ctx context.Context, quoteID string) (*quote.FXQuote, error)
    ReleaseQuote(ctx context.Context, quoteID string) error
    GetQuoteStatus(ctx context.Context, quoteID string) (*quote.FXQuote, error)

    // Conversion Operations
    ConvertCurrency(ctx context.Context, req *ConversionRequest) (*conversion.Conversion, error)
    ConvertWithQuote(ctx context.Context, quoteID string, amount decimal.Decimal) (*conversion.Conversion, error)
    GetConversionStatus(ctx context.Context, conversionID string) (*conversion.Conversion, error)
    RollbackConversion(ctx context.Context, conversionID string) error

    // Health & Monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*ServiceMetrics, error)
}

// QuoteRequest represents a request for an FX quote
type QuoteRequest struct {
    TransactionID   string          `json:"transactionId"`
    BaseCurrency    string          `json:"baseCurrency"`
    QuoteCurrency   string          `json:"quoteCurrency"`
    Amount          decimal.Decimal `json:"amount"`
    LockDuration    time.Duration   `json:"lockDuration"`
    SlippageTolerance decimal.Decimal `json:"slippageTolerance"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ConversionRequest represents a request for currency conversion
type ConversionRequest struct {
    TransactionID   string          `json:"transactionId"`
    FromCurrency    string          `json:"fromCurrency"`
    ToCurrency      string          `json:"toCurrency"`
    Amount          decimal.Decimal `json:"amount"`
    Rate            *decimal.Decimal `json:"rate,omitempty"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ServiceMetrics represents FX service metrics
type ServiceMetrics struct {
    RatesCached      int64   `json:"ratesCached"`
    QuotesActive     int64   `json:"quotesActive"`
    ConversionsToday int64   `json:"conversionsToday"`
    AverageLatency   float64 `json:"averageLatency"`
    SuccessRate      float64 `json:"successRate"`
    ProviderStatus   map[string]string `json:"providerStatus"`
}