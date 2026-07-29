package compensation

import (
    "time"

    "github.com/shopspring/decimal"
)

// Compensation represents a compensation operation for failed settlements
type Compensation struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    SettlementID    string          `json:"settlementId" gorm:"type:varchar(36);index"`
    OriginalBurnTx  string          `json:"originalBurnTx" gorm:"type:varchar(100)"`
    Network         string          `json:"network" gorm:"type:varchar(20)"`
    AccountID       string          `json:"accountId" gorm:"type:varchar(100)"`
    Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(38,18)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    Status          CompensationStatus `json:"status" gorm:"type:varchar(20)"`
    CompensationTx  string          `json:"compensationTx" gorm:"type:varchar(100)"`
    Reason          string          `json:"reason" gorm:"type:text"`
    AlertSent       bool            `json:"alertSent" gorm:"default:false"`
    ResolvedAt      *time.Time      `json:"resolvedAt,omitempty"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (Compensation) TableName() string {
    return "cbdc_compensations"
}