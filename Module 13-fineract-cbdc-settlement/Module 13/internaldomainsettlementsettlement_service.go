package settlement

import (
    "context"

    "github.com/shopspring/decimal"
)

// Service defines the settlement service interface
type Service interface {
    // Settlement Operations
    InitiateSettlement(ctx context.Context, req *SettlementRequest) (*Settlement, error)
    GetSettlement(ctx context.Context, id string) (*Settlement, error)
    GetSettlementByTransaction(ctx context.Context, transactionID string) (*Settlement, error)
    ListSettlements(ctx context.Context, filter *SettlementFilter) ([]*Settlement, int64, error)
    RetrySettlement(ctx context.Context, id string) (*Settlement, error)
    CancelSettlement(ctx context.Context, id string, reason string) error

    // Lock Operations
    LockFunds(ctx context.Context, req *LockRequest) (*lock.FundLock, error)
    ReleaseLock(ctx context.Context, lockID string) error
    GetLockStatus(ctx context.Context, lockID string) (*lock.FundLock, error)

    // Burn Operations
    BurnFunds(ctx context.Context, req *BurnRequest) (*burn.BurnRecord, error)
    GetBurnStatus(ctx context.Context, burnID string) (*burn.BurnRecord, error)

    // Compensation Operations
    CompensateBurn(ctx context.Context, req *CompensationRequest) (*compensation.Compensation, error)
    GetCompensationStatus(ctx context.Context, compID string) (*compensation.Compensation, error)

    // Health & Monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*ServiceMetrics, error)
}

// SettlementRequest represents a request to initiate a settlement
type SettlementRequest struct {
    TransactionID    string                 `json:"transactionId"`
    SourceNetwork    string                 `json:"sourceNetwork"`
    TargetNetwork    string                 `json:"targetNetwork"`
    SourceAccountID  string                 `json:"sourceAccountId"`
    TargetAccountID  string                 `json:"targetAccountId"`
    SourceCurrency   string                 `json:"sourceCurrency"`
    TargetCurrency   string                 `json:"targetCurrency"`
    SourceAmount     decimal.Decimal        `json:"sourceAmount"`
    TargetAmount     decimal.Decimal        `json:"targetAmount"`
    ConversionRate   decimal.Decimal        `json:"conversionRate"`
    Type             SettlementType         `json:"type"`
    Metadata         map[string]interface{} `json:"metadata"`
}

// SettlementFilter represents filters for listing settlements
type SettlementFilter struct {
    Status     SettlementStatus `json:"status"`
    Type       SettlementType   `json:"type"`
    Source     string           `json:"source"`
    Target     string           `json:"target"`
    FromDate   *time.Time       `json:"fromDate"`
    ToDate     *time.Time       `json:"toDate"`
    Limit      int              `json:"limit"`
    Offset     int              `json:"offset"`
}

// LockRequest represents a request to lock funds
type LockRequest struct {
    SettlementID string          `json:"settlementId"`
    Network      string          `json:"network"`
    AccountID    string          `json:"accountId"`
    Amount       decimal.Decimal `json:"amount"`
    Currency     string          `json:"currency"`
    Duration     time.Duration   `json:"duration"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// BurnRequest represents a request to burn funds
type BurnRequest struct {
    SettlementID string          `json:"settlementId"`
    Network      string          `json:"network"`
    LockID       string          `json:"lockId"`
    Amount       decimal.Decimal `json:"amount"`
    Currency     string          `json:"currency"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// CompensationRequest represents a request to compensate a failed burn
type CompensationRequest struct {
    SettlementID   string          `json:"settlementId"`
    Network        string          `json:"network"`
    AccountID      string          `json:"accountId"`
    Amount         decimal.Decimal `json:"amount"`
    Currency       string          `json:"currency"`
    OriginalBurnTx string          `json:"originalBurnTx"`
    Reason         string          `json:"reason"`
    Metadata       map[string]interface{} `json:"metadata"`
}