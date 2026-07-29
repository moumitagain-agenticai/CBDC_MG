package accounting

import (
    "context"
    "time"

    "github.com/apache/fineract-cbdc-accounting/internal/domain/ledger"
    "github.com/apache/fineract-cbdc-accounting/internal/domain/position"
    "github.com/apache/fineract-cbdc-accounting/internal/domain/revaluation"

    "github.com/shopspring/decimal"
)

// Service defines the accounting service interface
type Service interface {
    // Ledger Operations
    CreateAccount(ctx context.Context, req *CreateAccountRequest) (*ledger.LedgerAccount, error)
    GetAccount(ctx context.Context, id string) (*ledger.LedgerAccount, error)
    GetAccountByCode(ctx context.Context, code string) (*ledger.LedgerAccount, error)
    ListAccounts(ctx context.Context, filter *AccountFilter) ([]*ledger.LedgerAccount, int64, error)
    UpdateAccount(ctx context.Context, id string, req *UpdateAccountRequest) (*ledger.LedgerAccount, error)
    DeleteAccount(ctx context.Context, id string) error

    // Journal Entry Operations
    PostJournalEntry(ctx context.Context, req *JournalEntryRequest) (*JournalEntryResult, error)
    GetJournalEntry(ctx context.Context, id string) (*ledger.LedgerEntry, error)
    ListJournalEntries(ctx context.Context, filter *JournalEntryFilter) ([]*ledger.LedgerEntry, int64, error)

    // Position Operations
    GetCurrencyPosition(ctx context.Context, currency, tenantID string) (*position.CurrencyPosition, error)
    ListCurrencyPositions(ctx context.Context, filter *PositionFilter) ([]*position.CurrencyPosition, int64, error)
    UpdatePosition(ctx context.Context, req *UpdatePositionRequest) (*position.CurrencyPosition, error)

    // Revaluation Operations
    RevalueCurrency(ctx context.Context, req *RevaluationRequest) (*revaluation.Revaluation, error)
    GetRevaluation(ctx context.Context, id string) (*revaluation.Revaluation, error)
    ListRevaluations(ctx context.Context, filter *RevaluationFilter) ([]*revaluation.Revaluation, int64, error)

    // Balance & Reporting
    GetBalanceSheet(ctx context.Context, req *BalanceSheetRequest) (*BalanceSheet, error)
    GetTrialBalance(ctx context.Context, req *TrialBalanceRequest) (*TrialBalance, error)
    GetCurrencyPositionsReport(ctx context.Context, req *PositionReportRequest) (*PositionReport, error)

    // Health & Monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*ServiceMetrics, error)
}

// CreateAccountRequest represents a request to create a ledger account
type CreateAccountRequest struct {
    Code          string                 `json:"code"`
    Name          string                 `json:"name"`
    Description   string                 `json:"description"`
    Type          ledger.LedgerType      `json:"type"`
    Category      string                 `json:"category"`
    Currency      string                 `json:"currency"`
    IsRevaluation bool                   `json:"isRevaluation"`
    ParentID      *string                `json:"parentId"`
    TenantID      string                 `json:"tenantId"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateAccountRequest represents a request to update a ledger account
type UpdateAccountRequest struct {
    Name        *string                `json:"name"`
    Description *string                `json:"description"`
    Category    *string                `json:"category"`
    IsActive    *bool                  `json:"isActive"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// AccountFilter represents filters for listing accounts
type AccountFilter struct {
    Type      ledger.LedgerType `json:"type"`
    Category  string            `json:"category"`
    Currency  string            `json:"currency"`
    IsActive  *bool             `json:"isActive"`
    TenantID  string            `json:"tenantId"`
    Limit     int               `json:"limit"`
    Offset    int               `json:"offset"`
}

// JournalEntryRequest represents a request to post a journal entry
type JournalEntryRequest struct {
    TransactionID string                 `json:"transactionId"`
    Entries       []JournalEntryItem     `json:"entries"`
    Description   string                 `json:"description"`
    ReferenceID   string                 `json:"referenceId"`
    ReferenceType string                 `json:"referenceType"`
    TenantID      string                 `json:"tenantId"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// JournalEntryItem represents a single journal entry item
type JournalEntryItem struct {
    AccountID   string          `json:"accountId"`
    AccountCode string          `json:"accountCode"`
    EntryType   ledger.EntryType `json:"entryType"`
    Amount      decimal.Decimal `json:"amount"`
    Currency    string          `json:"currency"`
    Description string          `json:"description"`
}

// JournalEntryResult represents the result of posting a journal entry
type JournalEntryResult struct {
    EntryIDs    []string               `json:"entryIds"`
    TransactionID string               `json:"transactionId"`
    TotalDebit  decimal.Decimal        `json:"totalDebit"`
    TotalCredit decimal.Decimal        `json:"totalCredit"`
    IsBalanced  bool                   `json:"isBalanced"`
    CreatedAt   time.Time              `json:"createdAt"`
}

// JournalEntryFilter represents filters for listing journal entries
type JournalEntryFilter struct {
    AccountID   string     `json:"accountId"`
    TransactionID string   `json:"transactionId"`
    EntryType   ledger.EntryType `json:"entryType"`
    TenantID    string     `json:"tenantId"`
    FromDate    *time.Time `json:"fromDate"`
    ToDate      *time.Time `json:"toDate"`
    Limit       int        `json:"limit"`
    Offset      int        `json:"offset"`
}

// UpdatePositionRequest represents a request to update a currency position
type UpdatePositionRequest struct {
    Currency      string           `json:"currency"`
    TenantID      string           `json:"tenantId"`
    LongPosition  decimal.Decimal  `json:"longPosition"`
    ShortPosition decimal.Decimal  `json:"shortPosition"`
    TotalInflow   decimal.Decimal  `json:"totalInflow"`
    TotalOutflow  decimal.Decimal  `json:"totalOutflow"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// PositionFilter represents filters for listing positions
type PositionFilter struct {
    Currency  string `json:"currency"`
    TenantID  string `json:"tenantId"`
    Status    position.PositionStatus `json:"status"`
    Limit     int    `json:"limit"`
    Offset    int    `json:"offset"`
}

// RevaluationRequest represents a request to revalue a currency
type RevaluationRequest struct {
    Currency        string          `json:"currency"`
    TenantID        string          `json:"tenantId"`
    NewRate         decimal.Decimal `json:"newRate"`
    RevaluationDate time.Time       `json:"revaluationDate"`
    ReferenceID     string          `json:"referenceId"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// RevaluationFilter represents filters for listing revaluations
type RevaluationFilter struct {
    Currency    string     `json:"currency"`
    TenantID    string     `json:"tenantId"`
    Status      revaluation.RevaluationStatus `json:"status"`
    FromDate    *time.Time `json:"fromDate"`
    ToDate      *time.Time `json:"toDate"`
    Limit       int        `json:"limit"`
    Offset      int        `json:"offset"`
}

// BalanceSheetRequest represents a request for balance sheet
type BalanceSheetRequest struct {
    TenantID    string     `json:"tenantId"`
    AsOfDate    time.Time  `json:"asOfDate"`
    IncludeZero bool       `json:"includeZero"`
}

// BalanceSheet represents a balance sheet
type BalanceSheet struct {
    AsOfDate    time.Time              `json:"asOfDate"`
    TotalAssets decimal.Decimal        `json:"totalAssets"`
    TotalLiabilities decimal.Decimal   `json:"totalLiabilities"`
    TotalEquity decimal.Decimal        `json:"totalEquity"`
    Assets      []BalanceSheetItem     `json:"assets"`
    Liabilities []BalanceSheetItem     `json:"liabilities"`
    Equity      []BalanceSheetItem     `json:"equity"`
}

// BalanceSheetItem represents a balance sheet item
type BalanceSheetItem struct {
    AccountID   string          `json:"accountId"`
    AccountCode string          `json:"accountCode"`
    AccountName string          `json:"accountName"`
    Balance     decimal.Decimal `json:"balance"`
    Currency    string          `json:"currency"`
}

// TrialBalanceRequest represents a request for trial balance
type TrialBalanceRequest struct {
    TenantID    string     `json:"tenantId"`
    AsOfDate    time.Time  `json:"asOfDate"`
    IncludeZero bool       `json:"includeZero"`
}

// TrialBalance represents a trial balance
type TrialBalance struct {
    AsOfDate   time.Time             `json:"asOfDate"`
    TotalDebit decimal.Decimal       `json:"totalDebit"`
    TotalCredit decimal.Decimal      `json:"totalCredit"`
    Items      []TrialBalanceItem    `json:"items"`
}

// TrialBalanceItem represents a trial balance item
type TrialBalanceItem struct {
    AccountID   string          `json:"accountId"`
    AccountCode string          `json:"accountCode"`
    AccountName string          `json:"accountName"`
    Debit       decimal.Decimal `json:"debit"`
    Credit      decimal.Decimal `json:"credit"`
    Currency    string          `json:"currency"`
}

// PositionReportRequest represents a request for position report
type PositionReportRequest struct {
    TenantID string `json:"tenantId"`
    Currency string `json:"currency"`
    AsOfDate time.Time `json:"asOfDate"`
}

// PositionReport represents a currency position report
type PositionReport struct {
    AsOfDate    time.Time               `json:"asOfDate"`
    Positions   []PositionReportItem    `json:"positions"`
    TotalLong   decimal.Decimal         `json:"totalLong"`
    TotalShort  decimal.Decimal         `json:"totalShort"`
    TotalNet    decimal.Decimal         `json:"totalNet"`
}