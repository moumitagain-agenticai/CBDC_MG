package reconciliation

import (
    "context"
    "io"
    "time"

    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/exception"
    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/statement"

    "github.com/shopspring/decimal"
)

// Service defines the reconciliation service interface
type Service interface {
    // Reconciliation Operations
    CreateReconciliation(ctx context.Context, req *CreateReconciliationRequest) (*Reconciliation, error)
    GetReconciliation(ctx context.Context, id string) (*Reconciliation, error)
    ListReconciliations(ctx context.Context, filter *ReconciliationFilter) ([]*Reconciliation, int64, error)
    UpdateReconciliation(ctx context.Context, id string, req *UpdateReconciliationRequest) (*Reconciliation, error)
    DeleteReconciliation(ctx context.Context, id string) error

    // Statement Operations
    UploadStatement(ctx context.Context, req *UploadStatementRequest) (*statement.BankStatement, error)
    GetStatement(ctx context.Context, id string) (*statement.BankStatement, error)
    ListStatements(ctx context.Context, filter *StatementFilter) ([]*statement.BankStatement, int64, error)

    // Matching Operations
    ExecuteMatching(ctx context.Context, reconciliationID string) (*MatchingResult, error)
    GetMatchingResult(ctx context.Context, reconciliationID string) (*MatchingResult, error)

    // Exception Operations
    ListExceptions(ctx context.Context, filter *ExceptionFilter) ([]*exception.ExceptionItem, int64, error)
    ResolveException(ctx context.Context, id string, req *ResolveExceptionRequest) (*exception.ExceptionItem, error)

    // Reporting
    GenerateReconciliationReport(ctx context.Context, req *ReportRequest) (*ReconciliationReport, error)
    GenerateExceptionReport(ctx context.Context, req *ReportRequest) (*ExceptionReport, error)

    // Health & Monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*ServiceMetrics, error)
}

// CreateReconciliationRequest represents a request to create a reconciliation
type CreateReconciliationRequest struct {
    Name          string                 `json:"name"`
    Type          ReconciliationType     `json:"type"`
    AccountID     string                 `json:"accountId"`
    AccountNumber string                 `json:"accountNumber"`
    Currency      string                 `json:"currency"`
    StartDate     time.Time              `json:"startDate"`
    EndDate       time.Time              `json:"endDate"`
    OpeningBalance decimal.Decimal       `json:"openingBalance"`
    TenantID      string                 `json:"tenantId"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateReconciliationRequest represents a request to update a reconciliation
type UpdateReconciliationRequest struct {
    Name      *string                `json:"name"`
    Status    *ReconciliationStatus  `json:"status"`
    Metadata  map[string]interface{} `json:"metadata"`
}

// ReconciliationFilter represents filters for listing reconciliations
type ReconciliationFilter struct {
    Status    ReconciliationStatus `json:"status"`
    Type      ReconciliationType   `json:"type"`
    AccountID string               `json:"accountId"`
    TenantID  string               `json:"tenantId"`
    FromDate  *time.Time           `json:"fromDate"`
    ToDate    *time.Time           `json:"toDate"`
    Limit     int                  `json:"limit"`
    Offset    int                  `json:"offset"`
}

// UploadStatementRequest represents a request to upload a bank statement
type UploadStatementRequest struct {
    ReconciliationID string                 `json:"reconciliationId"`
    StatementType    string                 `json:"statementType"` // MT940, CSV, XML
    FileName         string                 `json:"fileName"`
    FileContent      []byte                 `json:"fileContent"`
    Metadata         map[string]interface{} `json:"metadata"`
}

// StatementFilter represents filters for listing statements
type StatementFilter struct {
    ReconciliationID string              `json:"reconciliationId"`
    AccountID        string              `json:"accountId"`
    StatementType    string              `json:"statementType"`
    Status           statement.StatementStatus `json:"status"`
    FromDate         *time.Time          `json:"fromDate"`
    ToDate           *time.Time          `json:"toDate"`
    Limit            int                 `json:"limit"`
    Offset           int                 `json:"offset"`
}

// MatchingResult represents the result of a matching operation
type MatchingResult struct {
    ReconciliationID string          `json:"reconciliationId"`
    TotalEntries     int             `json:"totalEntries"`
    MatchedEntries   int             `json:"matchedEntries"`
    UnmatchedEntries int             `json:"unmatchedEntries"`
    SystemBalance    decimal.Decimal `json:"systemBalance"`
    BankBalance      decimal.Decimal `json:"bankBalance"`
    Difference       decimal.Decimal `json:"difference"`
    IsBalanced       bool            `json:"isBalanced"`
    Details          []MatchDetail   `json:"details"`
    CompletedAt      time.Time       `json:"completedAt"`
}

// MatchDetail represents a single match detail
type MatchDetail struct {
    SystemEntryID     string          `json:"systemEntryId"`
    SystemAmount      decimal.Decimal `json:"systemAmount"`
    SystemDate        time.Time       `json:"systemDate"`
    BankEntryID       string          `json:"bankEntryId"`
    BankAmount        decimal.Decimal `json:"bankAmount"`
    BankDate          time.Time       `json:"bankDate"`
    MatchStatus       string          `json:"matchStatus"` // MATCHED, UNMATCHED, PARTIAL
    ConfidenceScore   float64         `json:"confidenceScore"`
}

// ExceptionFilter represents filters for listing exceptions
type ExceptionFilter struct {
    ReconciliationID string              `json:"reconciliationId"`
    Type             exception.ExceptionType `json:"type"`
    Status           exception.ExceptionStatus `json:"status"`
    Priority         string              `json:"priority"`
    TenantID         string              `json:"tenantId"`
    Limit            int                 `json:"limit"`
    Offset           int                 `json:"offset"`
}

// ResolveExceptionRequest represents a request to resolve an exception
type ResolveExceptionRequest struct {
    Resolution string                 `json:"resolution"`
    ResolvedBy string                 `json:"resolvedBy"`
    Metadata   map[string]interface{} `json:"metadata"`
}

// ReportRequest represents a request for a report
type ReportRequest struct {
    ReconciliationID string     `json:"reconciliationId"`
    Format           string     `json:"format"` // PDF, CSV, EXCEL
    IncludeDetails   bool       `json:"includeDetails"`
}

// ReconciliationReport represents a reconciliation report
type ReconciliationReport struct {
    ReportID        string                 `json:"reportId"`
    Reconciliation  *Reconciliation        `json:"reconciliation"`
    Summary         ReportSummary          `json:"summary"`
    MatchedItems    []ReportItem           `json:"matchedItems"`
    UnmatchedItems  []ReportItem           `json:"unmatchedItems"`
    Exceptions      []exception.ExceptionItem `json:"exceptions"`
    GeneratedAt     time.Time              `json:"generatedAt"`
}

// ReportSummary represents a report summary
type ReportSummary struct {
    TotalEntries     int             `json:"totalEntries"`
    MatchedEntries   int             `json:"matchedEntries"`
    UnmatchedEntries int             `json:"unmatchedEntries"`
    TotalAmount      decimal.Decimal `json:"totalAmount"`
    MatchedAmount    decimal.Decimal `json:"matchedAmount"`
    UnmatchedAmount  decimal.Decimal `json:"unmatchedAmount"`
}

// ReportItem represents a report item
type ReportItem struct {
    EntryID    string          `json:"entryId"`
    Date       time.Time       `json:"date"`
    Description string         `json:"description"`
    Amount     decimal.Decimal `json:"amount"`
    Currency   string          `json:"currency"`
    Status     string          `json:"status"`
    Reference  string          `json:"reference"`
}