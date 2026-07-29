package burn

import (
    "time"

    "github.com/shopspring/decimal"
)

// BurnRecord represents a burn operation on a CBDC network
type BurnRecord struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    SettlementID    string          `json:"settlementId" gorm:"type:varchar(36);index"`
    Network         string          `json:"network" gorm:"type:varchar(20)"`
    LockID          string          `json:"lockId" gorm:"type:varchar(100)"`
    TransactionID   string          `json:"transactionId" gorm:"type:varchar(100);uniqueIndex"`
    Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(38,18)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    Status          BurnStatus      `json:"status" gorm:"type:varchar(20)"`
    ConfirmationBlock string        `json:"confirmationBlock" gorm:"type:varchar(100)"`
    ConfirmedAt     *time.Time      `json:"confirmedAt,omitempty"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (BurnRecord) TableName() string {
    return "cbdc_burn_records"
}