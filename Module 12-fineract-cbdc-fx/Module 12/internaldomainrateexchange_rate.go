package rate

import (
    "time"

    "github.com/shopspring/decimal"
)

// ExchangeRate represents an exchange rate between two currencies
type ExchangeRate struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    BaseCurrency    string          `json:"baseCurrency" gorm:"type:varchar(3);index"`
    QuoteCurrency   string          `json:"quoteCurrency" gorm:"type:varchar(3);index"`
    BidRate         decimal.Decimal `json:"bidRate" gorm:"type:decimal(38,18)"`
    AskRate         decimal.Decimal `json:"askRate" gorm:"type:decimal(38,18)"`
    MidRate         decimal.Decimal `json:"midRate" gorm:"type:decimal(38,18)"`
    Spread          decimal.Decimal `json:"spread" gorm:"type:decimal(38,18)"`
    Provider        string          `json:"provider" gorm:"type:varchar(50)"`
    Status          RateStatus      `json:"status" gorm:"type:varchar(20)"`
    Timestamp       time.Time       `json:"timestamp"`
    ExpiresAt       time.Time       `json:"expiresAt"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (ExchangeRate) TableName() string {
    return "cbdc_exchange_rates"
}

// IsExpired checks if the rate is expired
func (r *ExchangeRate) IsExpired() bool {
    return time.Now().After(r.ExpiresAt)
}

// IsActive checks if the rate is active
func (r *ExchangeRate) IsActive() bool {
    return r.Status == RateStatusActive && !r.IsExpired()
}

// GetSpreadPercent calculates the spread as a percentage
func (r *ExchangeRate) GetSpreadPercent() decimal.Decimal {
    if r.MidRate.IsZero() {
        return decimal.Zero
    }
    return r.Spread.Div(r.MidRate).Mul(decimal.NewFromInt(100))
}

// Validate validates the exchange rate
func (r *ExchangeRate) Validate() error {
    if r.BaseCurrency == "" || len(r.BaseCurrency) != 3 {
        return ErrInvalidCurrencyCode
    }
    if r.QuoteCurrency == "" || len(r.QuoteCurrency) != 3 {
        return ErrInvalidCurrencyCode
    }
    if r.BidRate.IsNegative() || r.AskRate.IsNegative() || r.MidRate.IsNegative() {
        return ErrInvalidRate
    }
    if r.BidRate.GreaterThan(r.AskRate) {
        return ErrInvalidBidAsk
    }
    return nil
}