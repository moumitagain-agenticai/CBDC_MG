package statement

import (
    "time"

    "github.com/shopspring/decimal"
)

// BankStatement represents a bank statement
type BankStatement struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    ReconciliationID string         `json:"reconciliationId" gorm:"type:varchar(36);index"`
    AccountID       string          `json:"accountId" gorm:"type:varchar(36)"`
    AccountNumber   string          `json:"accountNumber" gorm:"type:varchar(50)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    StatementDate   time.Time       `json:"statementDate"`
    StatementType   string          `json:"statementType" gorm:"type:varchar(20)"` // MT940, XML, CSV
    OpeningBalance  decimal.Decimal `json:"openingBalance" gorm:"type:decimal(38,18)"`
    ClosingBalance  decimal.Decimal `json:"closingBalance" gorm:"type:decimal(38,18)"`
    TotalDebit      decimal.Decimal `json:"totalDebit" gorm:"type:decimal(38,18)"`
    TotalCredit     decimal.Decimal `json:"totalCredit" gorm:"type:decimal(38,18)"`
    Entries         []StatementEntry `json:"entries" gorm:"type:jsonb"`
    Status          StatementStatus `json:"status" gorm:"type:varchar(20)"`
    FileName        string          `json:"fileName" gorm:"type:varchar(200)"`
    FileContent     []byte          `json:"fileContent" gorm:"type:bytea"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// StatementEntry represents a single statement entry
type StatementEntry struct {
    ID          string          `json:"id"`
    Date        time.Time       `json:"date"`
    ValueDate   time.Time       `json:"valueDate"`
    Description string          `json:"description"`
    Amount      decimal.Decimal `json:"amount"`
    Currency    string          `json:"currency"`
    DebitCredit string          `json:"debitCredit"` // DEBIT, CREDIT
    Balance     decimal.Decimal `json:"balance"`
    Reference   string          `json:"reference"`
    TransactionID string        `json:"transactionId"`
    BankCode    string          `json:"bankCode"`
    BranchCode  string          `json:"branchCode"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// TableName returns the table name for GORM
func (BankStatement) TableName() string {
    return "cbdc_bank_statements"
}