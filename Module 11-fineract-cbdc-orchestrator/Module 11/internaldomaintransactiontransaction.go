package transaction

import (
    "time"

    "github.com/shopspring/decimal"
)

// Transaction represents a cross-border payment transaction
type Transaction struct {
    ID              string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Type            TransactionType        `json:"type" gorm:"type:varchar(20)"`
    State           TransactionState       `json:"state" gorm:"type:varchar(30)"`
    Status          TransactionStatus      `json:"status" gorm:"type:varchar(30)"`
    SourceCountry   string                 `json:"sourceCountry" gorm:"type:varchar(3)"`
    TargetCountry   string                 `json:"targetCountry" gorm:"type:varchar(3)"`
    SourceAccountID string                 `json:"sourceAccountId" gorm:"type:varchar(100)"`
    TargetAccountID string                 `json:"targetAccountId" gorm:"type:varchar(100)"`
    SourceCurrency  string                 `json:"sourceCurrency" gorm:"type:varchar(3)"`
    TargetCurrency  string                 `json:"targetCurrency" gorm:"type:varchar(3)"`
    SourceAmount    decimal.Decimal        `json:"sourceAmount" gorm:"type:decimal(38,18)"`
    TargetAmount    decimal.Decimal        `json:"targetAmount" gorm:"type:decimal(38,18)"`
    ConversionRate  decimal.Decimal        `json:"conversionRate" gorm:"type:decimal(38,18)"`
    LockReference   string                 `json:"lockReference" gorm:"type:varchar(100)"`
    SettlementID    string                 `json:"settlementId" gorm:"type:varchar(36)"`
    ErrorMessage    string                 `json:"errorMessage" gorm:"type:text"`
    CancelReason    string                 `json:"cancelReason" gorm:"type:text"`
    IdempotencyKey  string                 `json:"idempotencyKey" gorm:"type:varchar(100);uniqueIndex"`
    Attempts        int                    `json:"attempts" gorm:"default:0"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time              `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
    CompletedAt     *time.Time             `json:"completedAt,omitempty"`
    FailedAt        *time.Time             `json:"failedAt,omitempty"`
    CancelledAt     *time.Time             `json:"cancelledAt,omitempty"`
    Version         int64                  `json:"version" gorm:"default:0"`
}

// TableName returns the table name for GORM
func (Transaction) TableName() string {
    return "cbdc_transactions"
}

// IsTerminal checks if the transaction is in a terminal state
func (t *Transaction) IsTerminal() bool {
    return t.State == StateCompleted || t.State == StateFailed
}

// IsFinal checks if the transaction is in a final status
func (t *Transaction) IsFinal() bool {
    return t.Status == StatusCompleted ||
        t.Status == StatusFailed ||
        t.Status == StatusCancelled
}

// CanTransitionTo checks if the transaction can transition to a new state
func (t *Transaction) CanTransitionTo(newState TransactionState) bool {
    validTransitions := map[TransactionState][]TransactionState{
        StateInitiated:         {StateComplianceCheck},
        StateComplianceCheck:   {StateFXConversion, StateFailed},
        StateFXConversion:      {StateLockFunds, StateFailed},
        StateLockFunds:         {StateSettlement, StateFailed},
        StateSettlement:        {StateReleaseLock, StateFailed},
        StateReleaseLock:       {StateConfirmCompletion, StateFailed},
        StateConfirmCompletion: {StateCompleted, StateFailed},
        StateCompleted:         {},
        StateFailed:            {},
    }

    transitions, exists := validTransitions[t.State]
    if !exists {
        return false
    }

    for _, allowed := range transitions {
        if allowed == newState {
            return true
        }
    }
    return false
}