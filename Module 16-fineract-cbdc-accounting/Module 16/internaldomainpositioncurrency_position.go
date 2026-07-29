package position

import (
    "time"

    "github.com/shopspring/decimal"
)

// CurrencyPosition represents a currency position
type CurrencyPosition struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3);index"`
    TenantID        string          `json:"tenantId" gorm:"type:varchar(36);index"`
    LongPosition    decimal.Decimal `json:"longPosition" gorm:"type:decimal(38,18)"`
    ShortPosition   decimal.Decimal `json:"shortPosition" gorm:"type:decimal(38,18)"`
    NetPosition     decimal.Decimal `json:"netPosition" gorm:"type:decimal(38,18)"`
    TotalInflow     decimal.Decimal `json:"totalInflow" gorm:"type:decimal(38,18)"`
    TotalOutflow    decimal.Decimal `json:"totalOutflow" gorm:"type:decimal(38,18)"`
    RevaluationRate decimal.Decimal `json:"revaluationRate" gorm:"type:decimal(38,18)"`
    RevaluationGain decimal.Decimal `json:"revaluationGain" gorm:"type:decimal(38,18)"`
    Status          PositionStatus  `json:"status" gorm:"type:varchar(20)"`
    LastUpdated     time.Time       `json:"lastUpdated"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (CurrencyPosition) TableName() string {
    return "cbdc_currency_positions"
}

// IsBalanced checks if the position is balanced
func (p *CurrencyPosition) IsBalanced() bool {
    return p.LongPosition.Equals(p.ShortPosition)
}

// GetNetPosition returns the net position
func (p *CurrencyPosition) GetNetPosition() decimal.Decimal {
    return p.LongPosition.Sub(p.ShortPosition)
}