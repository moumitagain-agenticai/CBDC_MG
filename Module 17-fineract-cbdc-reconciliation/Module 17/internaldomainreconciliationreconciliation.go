package reconciliation

import (
    "time"

    "github.com/shopspring/decimal"
)

// Reconciliation represents a reconciliation session
type Reconciliation struct {
    ID                string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name              string          `json:"name" gorm:"type:varchar(100)"`
    Type              ReconciliationType `json:"type" gorm:"type:varchar(20)"`
    Status            ReconciliationStatus `json:"status" gorm:"type:varchar(30)"`
    AccountID         string          `json:"accountId" gorm:"type:varchar(36);index"`
    AccountNumber     string          `json:"accountNumber" gorm:"type:varchar(50)"`
    Currency          string          `json:"currency" gorm:"type:varchar(3)"`
    StartDate         time.Time       `json:"startDate"`
    EndDate           time.Time       `json:"endDate"`
    OpeningBalance    decimal.Decimal `json:"openingBalance" gorm:"type:decimal(38,18)"`
    ClosingBalance    decimal.Decimal `json:"closingBalance" gorm:"type:decimal(38,18)"`
    SystemBalance     decimal.Decimal `json:"systemBalance" gorm:"type:decimal(38,18)"`
    BankBalance       decimal.Decimal `json:"bankBalance" gorm:"type:decimal(38,18)"`
    Difference        decimal.Decimal `json:"difference" gorm:"type:decimal(38,18)"`
    TotalEntries      int             `json:"totalEntries"`
    MatchedEntries    int             `json:"matchedEntries"`
    UnmatchedEntries  int             `json:"unmatchedEntries"`
    TenantID          string          `json:"tenantId" gorm:"type:varchar(36);index"`
    Metadata          map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt         time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt         time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    CompletedAt       *time.Time      `json:"completedAt,omitempty"`
}

// TableName returns the table name for GORM
func (Reconciliation) TableName() string {
    return "cbdc_reconciliations"
}

// IsTerminal checks if the reconciliation is in a terminal state
func (r *Reconciliation) IsTerminal() bool {
    return r.Status == StatusCompleted || r.Status == StatusFailed || r.Status == StatusCancelled
}

// IsBalanced checks if the reconciliation is balanced
func (r *Reconciliation) IsBalanced() bool {
    return r.Difference.IsZero()
}