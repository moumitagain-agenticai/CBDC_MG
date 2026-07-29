package corridor

import (
    "time"

    "github.com/shopspring/decimal"
)

// FeeCorridor represents a fee corridor configuration
type FeeCorridor struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name            string          `json:"name" gorm:"type:varchar(100);index"`
    Code            string          `json:"code" gorm:"type:varchar(50);uniqueIndex"`
    SourceCountry   string          `json:"sourceCountry" gorm:"type:varchar(3)"`
    TargetCountry   string          `json:"targetCountry" gorm:"type:varchar(3)"`
    SourceCurrency  string          `json:"sourceCurrency" gorm:"type:varchar(3)"`
    TargetCurrency  string          `json:"targetCurrency" gorm:"type:varchar(3)"`
    BaseFee         decimal.Decimal `json:"baseFee" gorm:"type:decimal(38,18)"`
    Markup          decimal.Decimal `json:"markup" gorm:"type:decimal(38,18)"`
    Discount        decimal.Decimal `json:"discount" gorm:"type:decimal(38,18)"`
    MinFee          decimal.Decimal `json:"minFee" gorm:"type:decimal(38,18)"`
    MaxFee          decimal.Decimal `json:"maxFee" gorm:"type:decimal(38,18)"`
    IsActive        bool            `json:"isActive" gorm:"default:true"`
    Priority        int             `json:"priority" gorm:"default:0"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (FeeCorridor) TableName() string {
    return "cbdc_fee_corridors"
}

// Matches checks if the corridor matches the given parameters
func (c *FeeCorridor) Matches(sourceCountry, targetCountry, sourceCurrency, targetCurrency string) bool {
    return c.SourceCountry == sourceCountry &&
        c.TargetCountry == targetCountry &&
        c.SourceCurrency == sourceCurrency &&
        c.TargetCurrency == targetCurrency
}

// CalculateFee calculates the fee using corridor configuration
func (c *FeeCorridor) CalculateFee(amount decimal.Decimal) decimal.Decimal {
    fee := amount.Mul(c.BaseFee).Div(decimal.NewFromInt(100))
    fee = fee.Add(c.Markup).Sub(c.Discount)

    // Apply min/max
    if !c.MinFee.IsZero() && fee.LessThan(c.MinFee) {
        fee = c.MinFee
    }
    if !c.MaxFee.IsZero() && fee.GreaterThan(c.MaxFee) {
        fee = c.MaxFee
    }

    return fee
}