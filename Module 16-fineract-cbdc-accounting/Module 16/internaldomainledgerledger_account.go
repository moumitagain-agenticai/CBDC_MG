package ledger

import (
    "time"

    "github.com/shopspring/decimal"
)

// LedgerAccount represents a general ledger account
type LedgerAccount struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Code            string          `json:"code" gorm:"type:varchar(50);uniqueIndex"`
    Name            string          `json:"name" gorm:"type:varchar(100);index"`
    Description     string          `json:"description" gorm:"type:text"`
    Type            LedgerType      `json:"type" gorm:"type:varchar(20)"`
    Category        string          `json:"category" gorm:"type:varchar(50)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    Balance         decimal.Decimal `json:"balance" gorm:"type:decimal(38,18)"`
    DebitBalance    decimal.Decimal `json:"debitBalance" gorm:"type:decimal(38,18)"`
    CreditBalance   decimal.Decimal `json:"creditBalance" gorm:"type:decimal(38,18)"`
    IsActive        bool            `json:"isActive" gorm:"default:true"`
    IsRevaluation   bool            `json:"isRevaluation" gorm:"default:false"`
    ParentID        *string         `json:"parentId" gorm:"type:varchar(36)"`
    TenantID        string          `json:"tenantId" gorm:"type:varchar(36);index"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (LedgerAccount) TableName() string {
    return "cbdc_ledger_accounts"
}

// GetBalance returns the current balance
func (a *LedgerAccount) GetBalance() decimal.Decimal {
    return a.Balance
}

// GetDebitBalance returns the debit balance
func (a *LedgerAccount) GetDebitBalance() decimal.Decimal {
    return a.DebitBalance
}

// GetCreditBalance returns the credit balance
func (a *LedgerAccount) GetCreditBalance() decimal.Decimal {
    return a.CreditBalance
}

// IsBalanced checks if the account is balanced (debit = credit)
func (a *LedgerAccount) IsBalanced() bool {
    return a.DebitBalance.Equals(a.CreditBalance)
}