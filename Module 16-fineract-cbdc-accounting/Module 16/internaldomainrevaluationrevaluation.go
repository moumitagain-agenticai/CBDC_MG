package revaluation

import (
    "time"

    "github.com/shopspring/decimal"
)

// Revaluation represents a currency revaluation
type Revaluation struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3);index"`
    TenantID        string          `json:"tenantId" gorm:"type:varchar(36);index"`
    OldRate         decimal.Decimal `json:"oldRate" gorm:"type:decimal(38,18)"`
    NewRate         decimal.Decimal `json:"newRate" gorm:"type:decimal(38,18)"`
    OldPosition     decimal.Decimal `json:"oldPosition" gorm:"type:decimal(38,18)"`
    NewPosition     decimal.Decimal `json:"newPosition" gorm:"type:decimal(38,18)"`
    GainLoss        decimal.Decimal `json:"gainLoss" gorm:"type:decimal(38,18)"`
    GainLossType    GainLossType    `json:"gainLossType" gorm:"type:varchar(10)"`
    Status          RevaluationStatus `json:"status" gorm:"type:varchar(20)"`
    RevaluationDate time.Time       `json:"revaluationDate"`
    ReferenceID     string          `json:"referenceId" gorm:"type:varchar(36)"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (Revaluation) TableName() string {
    return "cbdc_revaluations"
}