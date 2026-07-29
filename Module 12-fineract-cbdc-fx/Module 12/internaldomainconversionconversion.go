package conversion

import (
    "time"

    "github.com/shopspring/decimal"
)

// Conversion represents a currency conversion operation
type Conversion struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TransactionID   string          `json:"transactionId" gorm:"type:varchar(36);index"`
    QuoteID         string          `json:"quoteId" gorm:"type:varchar(36)"`
    FromCurrency    string          `json:"fromCurrency" gorm:"type:varchar(3)"`
    ToCurrency      string          `json:"toCurrency" gorm:"type:varchar(3)"`
    FromAmount      decimal.Decimal `json:"fromAmount" gorm:"type:decimal(38,18)"`
    ToAmount        decimal.Decimal `json:"toAmount" gorm:"type:decimal(38,18)"`
    RateUsed        decimal.Decimal `json:"rateUsed" gorm:"type:decimal(38,18)"`
    FeeAmount       decimal.Decimal `json:"feeAmount" gorm:"type:decimal(38,18)"`
    FeeCurrency     string          `json:"feeCurrency" gorm:"type:varchar(3)"`
    MarkupApplied   decimal.Decimal `json:"markupApplied" gorm:"type:decimal(38,18)"`
    SlippageApplied decimal.Decimal `json:"slippageApplied" gorm:"type:decimal(38,18)"`
    Status          ConversionStatus `json:"status" gorm:"type:varchar(20)"`
    CompletedAt     *time.Time      `json:"completedAt,omitempty"`
    FailedAt        *time.Time      `json:"failedAt,omitempty"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (Conversion) TableName() string {
    return "cbdc_conversions"
}

// IsCompleted checks if the conversion is completed
func (c *Conversion) IsCompleted() bool {
    return c.Status == ConversionStatusCompleted
}

// IsFailed checks if the conversion failed
func (c *Conversion) IsFailed() bool {
    return c.Status == ConversionStatusFailed
}