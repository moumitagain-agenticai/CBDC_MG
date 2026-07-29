package fee

import (
    "time"

    "github.com/shopspring/decimal"
)

// Fee represents a fee configuration
type Fee struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name            string          `json:"name" gorm:"type:varchar(100);index"`
    Code            string          `json:"code" gorm:"type:varchar(50);uniqueIndex"`
    Type            FeeType         `json:"type" gorm:"type:varchar(20)"`
    Structure       FeeStructure    `json:"structure" gorm:"type:varchar(20)"`
    Value           decimal.Decimal `json:"value" gorm:"type:decimal(38,18)"`
    MinAmount       decimal.Decimal `json:"minAmount" gorm:"type:decimal(38,18)"`
    MaxAmount       decimal.Decimal `json:"maxAmount" gorm:"type:decimal(38,18)"`
    TieredStructure []Tier          `json:"tieredStructure" gorm:"type:jsonb"`
    CorridorID      string          `json:"corridorId" gorm:"type:varchar(36);index"`
    SourceCountry   string          `json:"sourceCountry" gorm:"type:varchar(3)"`
    TargetCountry   string          `json:"targetCountry" gorm:"type:varchar(3)"`
    SourceCurrency  string          `json:"sourceCurrency" gorm:"type:varchar(3)"`
    TargetCurrency  string          `json:"targetCurrency" gorm:"type:varchar(3)"`
    IsActive        bool            `json:"isActive" gorm:"default:true"`
    Priority        int             `json:"priority" gorm:"default:0"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Tier represents a tier in a tiered fee structure
type Tier struct {
    FromAmount  decimal.Decimal `json:"fromAmount"`
    ToAmount    decimal.Decimal `json:"toAmount"`
    Rate        decimal.Decimal `json:"rate"`
    FixedAmount decimal.Decimal `json:"fixedAmount"`
}

// TableName returns the table name for GORM
func (Fee) TableName() string {
    return "cbdc_fees"
}

// IsApplicable checks if the fee is applicable for given parameters
func (f *Fee) IsApplicable(corridorID, sourceCountry, targetCountry, sourceCurrency, targetCurrency string) bool {
    if !f.IsActive {
        return false
    }

    if f.CorridorID != "" && f.CorridorID != corridorID {
        return false
    }

    if f.SourceCountry != "" && f.SourceCountry != sourceCountry {
        return false
    }

    if f.TargetCountry != "" && f.TargetCountry != targetCountry {
        return false
    }

    if f.SourceCurrency != "" && f.SourceCurrency != sourceCurrency {
        return false
    }

    if f.TargetCurrency != "" && f.TargetCurrency != targetCurrency {
        return false
    }

    return true
}

// Calculate calculates the fee amount
func (f *Fee) Calculate(amount decimal.Decimal) (decimal.Decimal, error) {
    if amount.IsNegative() {
        return decimal.Zero, ErrInvalidAmount
    }

    // Check min/max
    if !f.MinAmount.IsZero() && amount.LessThan(f.MinAmount) {
        return decimal.Zero, ErrAmountBelowMinimum
    }

    if !f.MaxAmount.IsZero() && amount.GreaterThan(f.MaxAmount) {
        return decimal.Zero, ErrAmountExceedsMaximum
    }

    switch f.Structure {
    case StructureFlat:
        return f.calculateFlat(), nil
    case StructurePercentage:
        return f.calculatePercentage(amount), nil
    case StructureTiered:
        return f.calculateTiered(amount), nil
    default:
        return decimal.Zero, ErrInvalidStructure
    }
}

// calculateFlat calculates flat fee
func (f *Fee) calculateFlat() decimal.Decimal {
    return f.Value
}

// calculatePercentage calculates percentage fee
func (f *Fee) calculatePercentage(amount decimal.Decimal) decimal.Decimal {
    return amount.Mul(f.Value).Div(decimal.NewFromInt(100))
}

// calculateTiered calculates tiered fee
func (f *Fee) calculateTiered(amount decimal.Decimal) decimal.Decimal {
    totalFee := decimal.Zero

    for _, tier := range f.TieredStructure {
        if amount.GreaterThanOrEqual(tier.FromAmount) {
            if amount.LessThanOrEqual(tier.ToAmount) || tier.ToAmount.IsZero() {
                // Within this tier
                tierAmount := amount.Sub(tier.FromAmount)
                feeAmount := tierAmount.Mul(tier.Rate).Div(decimal.NewFromInt(100))
                totalFee = totalFee.Add(feeAmount).Add(tier.FixedAmount)
                break
            } else {
                // Partial tier amount
                tierRange := tier.ToAmount.Sub(tier.FromAmount)
                feeAmount := tierRange.Mul(tier.Rate).Div(decimal.NewFromInt(100))
                totalFee = totalFee.Add(feeAmount).Add(tier.FixedAmount)
            }
        }
    }

    return totalFee
}