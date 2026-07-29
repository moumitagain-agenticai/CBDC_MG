package quote

import (
    "time"

    "github.com/shopspring/decimal"
)

// FXQuote represents a locked exchange rate quote for a transaction
type FXQuote struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TransactionID   string          `json:"transactionId" gorm:"type:varchar(36);index"`
    BaseCurrency    string          `json:"baseCurrency" gorm:"type:varchar(3)"`
    QuoteCurrency   string          `json:"quoteCurrency" gorm:"type:varchar(3)"`
    BaseAmount      decimal.Decimal `json:"baseAmount" gorm:"type:decimal(38,18)"`
    QuoteAmount     decimal.Decimal `json:"quoteAmount" gorm:"type:decimal(38,18)"`
    Rate            decimal.Decimal `json:"rate" gorm:"type:decimal(38,18)"`
    BidRate         decimal.Decimal `json:"bidRate" gorm:"type:decimal(38,18)"`
    AskRate         decimal.Decimal `json:"askRate" gorm:"type:decimal(38,18)"`
    Spread          decimal.Decimal `json:"spread" gorm:"type:decimal(38,18)"`
    MarkupPercent   decimal.Decimal `json:"markupPercent" gorm:"type:decimal(38,18)"`
    MarkupAmount    decimal.Decimal `json:"markupAmount" gorm:"type:decimal(38,18)"`
    SlippagePercent decimal.Decimal `json:"slippagePercent" gorm:"type:decimal(38,18)"`
    SlippageAmount  decimal.Decimal `json:"slippageAmount" gorm:"type:decimal(38,18)"`
    FinalRate       decimal.Decimal `json:"finalRate" gorm:"type:decimal(38,18)"`
    Status          QuoteStatus     `json:"status" gorm:"type:varchar(20)"`
    LockDuration    time.Duration   `json:"lockDuration"`
    ExpiresAt       time.Time       `json:"expiresAt"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    LockedAt        *time.Time      `json:"lockedAt,omitempty"`
    UsedAt          *time.Time      `json:"usedAt,omitempty"`
    ExpiredAt       *time.Time      `json:"expiredAt,omitempty"`
}

// TableName returns the table name for GORM
func (FXQuote) TableName() string {
    return "cbdc_fx_quotes"
}

// IsExpired checks if the quote is expired
func (q *FXQuote) IsExpired() bool {
    return time.Now().After(q.ExpiresAt)
}

// IsValid checks if the quote is still valid
func (q *FXQuote) IsValid() bool {
    return q.Status == QuoteStatusActive && !q.IsExpired()
}

// CalculateSlippage calculates slippage based on current rate
func (q *FXQuote) CalculateSlippage(currentRate decimal.Decimal) decimal.Decimal {
    if q.Rate.IsZero() {
        return decimal.Zero
    }
    diff := currentRate.Sub(q.Rate).Abs()
    return diff.Div(q.Rate).Mul(decimal.NewFromInt(100))
}