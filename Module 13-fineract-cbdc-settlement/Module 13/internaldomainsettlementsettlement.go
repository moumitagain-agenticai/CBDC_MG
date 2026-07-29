package settlement

import (
    "time"

    "github.com/shopspring/decimal"
)

// Settlement represents an atomic settlement transaction
type Settlement struct {
    ID                 string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TransactionID      string          `json:"transactionId" gorm:"type:varchar(36);index"`
    SourceNetwork      string          `json:"sourceNetwork" gorm:"type:varchar(20)"`
    TargetNetwork      string          `json:"targetNetwork" gorm:"type:varchar(20)"`
    SourceAccountID    string          `json:"sourceAccountId" gorm:"type:varchar(100)"`
    TargetAccountID    string          `json:"targetAccountId" gorm:"type:varchar(100)"`
    SourceCurrency     string          `json:"sourceCurrency" gorm:"type:varchar(3)"`
    TargetCurrency     string          `json:"targetCurrency" gorm:"type:varchar(3)"`
    SourceAmount       decimal.Decimal `json:"sourceAmount" gorm:"type:decimal(38,18)"`
    TargetAmount       decimal.Decimal `json:"targetAmount" gorm:"type:decimal(38,18)"`
    ConversionRate     decimal.Decimal `json:"conversionRate" gorm:"type:decimal(38,18)"`
    SourceLockID       string          `json:"sourceLockId" gorm:"type:varchar(100)"`
    TargetLockID       string          `json:"targetLockId" gorm:"type:varchar(100)"`
    BurnTransactionID  string          `json:"burnTransactionId" gorm:"type:varchar(100)"`
    IssueTransactionID string          `json:"issueTransactionId" gorm:"type:varchar(100)"`
    Status             SettlementStatus `json:"status" gorm:"type:varchar(30)"`
    Type               SettlementType  `json:"type" gorm:"type:varchar(20)"`
    ErrorMessage       string          `json:"errorMessage" gorm:"type:text"`
    RetryCount         int             `json:"retryCount" gorm:"default:0"`
    Metadata           map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt          time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt          time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    StartedAt          *time.Time      `json:"startedAt,omitempty"`
    CompletedAt        *time.Time      `json:"completedAt,omitempty"`
    FailedAt           *time.Time      `json:"failedAt,omitempty"`
    RolledBackAt       *time.Time      `json:"rolledBackAt,omitempty"`
    Version            int64           `json:"version" gorm:"default:0"`
}

// TableName returns the table name for GORM
func (Settlement) TableName() string {
    return "cbdc_settlements"
}

// IsTerminal checks if the settlement is in a terminal state
func (s *Settlement) IsTerminal() bool {
    return s.Status == StatusCompleted ||
        s.Status == StatusFailed ||
        s.Status == StatusRolledBack
}

// IsPending checks if the settlement is pending
func (s *Settlement) IsPending() bool {
    return s.Status == StatusPending ||
        s.Status == StatusLocking ||
        s.Status == StatusBurning ||
        s.Status == StatusIssuing ||
        s.Status == StatusReleasing
}

// CanTransitionTo checks if the settlement can transition to a new status
func (s *Settlement) CanTransitionTo(newStatus SettlementStatus) bool {
    validTransitions := map[SettlementStatus][]SettlementStatus{
        StatusPending:    {StatusLocking, StatusFailed},
        StatusLocking:    {StatusBurning, StatusFailed, StatusRollingBack},
        StatusBurning:    {StatusIssuing, StatusFailed, StatusRollingBack},
        StatusIssuing:    {StatusReleasing, StatusFailed, StatusRollingBack},
        StatusReleasing:  {StatusCompleted, StatusFailed, StatusRollingBack},
        StatusCompleted:  {},
        StatusFailed:     {},
        StatusRolledBack: {},
        StatusRollingBack: {StatusRolledBack},
    }

    transitions, exists := validTransitions[s.Status]
    if !exists {
        return false
    }

    for _, allowed := range transitions {
        if allowed == newStatus {
            return true
        }
    }
    return false
}