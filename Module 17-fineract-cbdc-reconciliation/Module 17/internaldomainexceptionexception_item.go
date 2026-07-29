package exception

import (
    "time"

    "github.com/shopspring/decimal"
)

// ExceptionItem represents a reconciliation exception
type ExceptionItem struct {
    ID                string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    ReconciliationID  string          `json:"reconciliationId" gorm:"type:varchar(36);index"`
    Type              ExceptionType   `json:"type" gorm:"type:varchar(30)"`
    Status            ExceptionStatus `json:"status" gorm:"type:varchar(20)"`
    Priority          string          `json:"priority" gorm:"type:varchar(10)"` // HIGH, MEDIUM, LOW
    Description       string          `json:"description" gorm:"type:text"`
    Amount            decimal.Decimal `json:"amount" gorm:"type:decimal(38,18)"`
    Currency          string          `json:"currency" gorm:"type:varchar(3)"`
    SystemTransactionID string        `json:"systemTransactionId" gorm:"type:varchar(36)"`
    BankTransactionID string          `json:"bankTransactionId" gorm:"type:varchar(100)"`
    Date              time.Time       `json:"date"`
    Resolution        string          `json:"resolution" gorm:"type:text"`
    ResolvedBy        string          `json:"resolvedBy" gorm:"type:varchar(100)"`
    ResolvedAt        *time.Time      `json:"resolvedAt,omitempty"`
    TenantID          string          `json:"tenantId" gorm:"type:varchar(36);index"`
    Metadata          map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt         time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt         time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (ExceptionItem) TableName() string {
    return "cbdc_exception_items"
}

// IsResolved checks if the exception is resolved
func (e *ExceptionItem) IsResolved() bool {
    return e.Status == StatusResolved || e.Status == StatusClosed
}