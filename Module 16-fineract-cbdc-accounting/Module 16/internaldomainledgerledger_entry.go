package ledger

import (
    "time"

    "github.com/shopspring/decimal"
)

// LedgerEntry represents a journal entry in the ledger
type LedgerEntry struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TransactionID   string          `json:"transactionId" gorm:"type:varchar(36);index"`
    AccountID       string          `json:"accountId" gorm:"type:varchar(36);index"`
    AccountCode     string          `json:"accountCode" gorm:"type:varchar(50)"`
    EntryType       EntryType       `json:"entryType" gorm:"type:varchar(10)"` // DEBIT, CREDIT
    Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(38,18)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    BalanceAfter    decimal.Decimal `json:"balanceAfter" gorm:"type:decimal(38,18)"`
    Description     string          `json:"description" gorm:"type:text"`
    ReferenceID     string          `json:"referenceId" gorm:"type:varchar(36)"`
    ReferenceType   string          `json:"referenceType" gorm:"type:varchar(50)"`
    TenantID        string          `json:"tenantId" gorm:"type:varchar(36);index"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (LedgerEntry) TableName() string {
    return "cbdc_ledger_entries"
}

// IsDebit checks if the entry is a debit
func (e *LedgerEntry) IsDebit() bool {
    return e.EntryType == EntryTypeDebit
}

// IsCredit checks if the entry is a credit
func (e *LedgerEntry) IsCredit() bool {
    return e.EntryType == EntryTypeCredit
}