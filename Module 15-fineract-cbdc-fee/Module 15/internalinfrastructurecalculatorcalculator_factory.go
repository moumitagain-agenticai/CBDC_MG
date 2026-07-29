package calculator

import (
    "fmt"

    "github.com/apache/fineract-cbdc-fee/internal/domain/fee"

    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// FeeCalculator defines the interface for fee calculators
type FeeCalculator interface {
    Calculate(amount decimal.Decimal, fee *fee.Fee) (decimal.Decimal, error)
}

// CalculatorFactory creates fee calculators
type CalculatorFactory struct {
    logger *zap.Logger
}

// NewCalculatorFactory creates a new calculator factory
func NewCalculatorFactory(logger *zap.Logger) *CalculatorFactory {
    return &CalculatorFactory{
        logger: logger,
    }
}

// GetCalculator returns a calculator for the given fee structure
func (f *CalculatorFactory) GetCalculator(structure fee.FeeStructure) (FeeCalculator, error) {
    switch structure {
    case fee.StructureFlat:
        return &FlatCalculator{}, nil
    case fee.StructurePercentage:
        return &PercentageCalculator{}, nil
    case fee.StructureTiered:
        return &TieredCalculator{}, nil
    default:
        return nil, fmt.Errorf("unsupported fee structure: %s", structure)
    }
}

// FlatCalculator calculates flat fees
type FlatCalculator struct{}

func (c *FlatCalculator) Calculate(amount decimal.Decimal, fee *fee.Fee) (decimal.Decimal, error) {
    return fee.Value, nil
}

// PercentageCalculator calculates percentage fees
type PercentageCalculator struct{}

func (c *PercentageCalculator) Calculate(amount decimal.Decimal, fee *fee.Fee) (decimal.Decimal, error) {
    return amount.Mul(fee.Value).Div(decimal.NewFromInt(100)), nil
}

// TieredCalculator calculates tiered fees
type TieredCalculator struct{}

func (c *TieredCalculator) Calculate(amount decimal.Decimal, fee *fee.Fee) (decimal.Decimal, error) {
    totalFee := decimal.Zero

    for _, tier := range fee.TieredStructure {
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

    return totalFee, nil
}