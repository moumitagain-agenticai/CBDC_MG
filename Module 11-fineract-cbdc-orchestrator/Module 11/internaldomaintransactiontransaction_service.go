package transaction

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

// Service defines the transaction domain service interface
type Service interface {
    // InitiatePayment initiates a new payment
    InitiatePayment(ctx context.Context, req *PaymentRequest) (*Transaction, error)

    // GetTransaction gets a transaction by ID
    GetTransaction(ctx context.Context, id string) (*Transaction, error)

    // GetTransactionByIdempotencyKey gets a transaction by idempotency key
    GetTransactionByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)

    // ListTransactions lists transactions with filters
    ListTransactions(ctx context.Context, filter *TransactionFilter) ([]*Transaction, int64, error)

    // CancelTransaction cancels a pending transaction
    CancelTransaction(ctx context.Context, id string, reason string) error

    // RetryTransaction retries a failed transaction
    RetryTransaction(ctx context.Context, id string) error

    // ProcessState processes the current state of a transaction
    ProcessState(ctx context.Context, tx *Transaction) (*Transaction, error)

    // CompleteTransaction marks a transaction as completed
    CompleteTransaction(ctx context.Context, tx *Transaction) error

    // FailTransaction marks a transaction as failed
    FailTransaction(ctx context.Context, tx *Transaction, errorMsg string) error
}

// PaymentRequest represents a payment initiation request
type PaymentRequest struct {
    SourceCountry   string                 `json:"sourceCountry"`
    TargetCountry   string                 `json:"targetCountry"`
    SourceAccountID string                 `json:"sourceAccountId"`
    TargetAccountID string                 `json:"targetAccountId"`
    SourceCurrency  string                 `json:"sourceCurrency"`
    TargetCurrency  string                 `json:"targetCurrency"`
    Amount          decimal.Decimal        `json:"amount"`
    Description     string                 `json:"description"`
    IdempotencyKey  string                 `json:"idempotencyKey"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// domainService implements the Service interface
type domainService struct {
    repo         Repository
    stateMachine *StateMachine
    eventBus     EventBus
    config       *ServiceConfig
}

// ServiceConfig holds the service configuration
type ServiceConfig struct {
    MaxConcurrent      int           `mapstructure:"max_concurrent"`
    ProcessingTimeout  time.Duration `mapstructure:"processing_timeout"`
    MaxRetries         int           `mapstructure:"max_retries"`
    RetryDelay         time.Duration `mapstructure:"retry_delay"`
    LockDuration       time.Duration `mapstructure:"lock_duration"`
}

// NewDomainService creates a new transaction domain service
func NewDomainService(
    repo Repository,
    stateMachine *StateMachine,
    eventBus EventBus,
    config *ServiceConfig,
) Service {
    return &domainService{
        repo:         repo,
        stateMachine: stateMachine,
        eventBus:     eventBus,
        config:       config,
    }
}

// InitiatePayment initiates a new payment
func (s *domainService) InitiatePayment(ctx context.Context, req *PaymentRequest) (*Transaction, error) {
    // Validate request
    if err := s.validatePaymentRequest(req); err != nil {
        return nil, err
    }

    // Check idempotency
    if req.IdempotencyKey != "" {
        existing, err := s.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
        if err == nil && existing != nil {
            return existing, nil
        }
    }

    // Create transaction
    tx := &Transaction{
        ID:              uuid.New().String(),
        Type:            TypeCrossBorder,
        State:           StateInitiated,
        Status:          StatusPending,
        SourceCountry:   req.SourceCountry,
        TargetCountry:   req.TargetCountry,
        SourceAccountID: req.SourceAccountID,
        TargetAccountID: req.TargetAccountID,
        SourceCurrency:  req.SourceCurrency,
        TargetCurrency:  req.TargetCurrency,
        SourceAmount:    req.Amount,
        TargetAmount:    decimal.Zero,
        ConversionRate:  decimal.Zero,
        Attempts:        0,
        IdempotencyKey:  req.IdempotencyKey,
        Metadata:        req.Metadata,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Save transaction
    if err := s.repo.Create(ctx, tx); err != nil {
        return nil, err
    }

    return tx, nil
}

// GetTransaction gets a transaction by ID
func (s *domainService) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
    return s.repo.GetByID(ctx, id)
}

// GetTransactionByIdempotencyKey gets a transaction by idempotency key
func (s *domainService) GetTransactionByIdempotencyKey(ctx context.Context, key string) (*Transaction, error) {
    return s.repo.GetByIdempotencyKey(ctx, key)
}

// ListTransactions lists transactions with filters
func (s *domainService) ListTransactions(ctx context.Context, filter *TransactionFilter) ([]*Transaction, int64, error) {
    return s.repo.List(ctx, filter)
}

// CancelTransaction cancels a pending transaction
func (s *domainService) CancelTransaction(ctx context.Context, id string, reason string) error {
    tx, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    if tx.IsFinal() {
        return NewInvalidStateError("transaction already in final state", tx.State)
    }

    tx.State = StateFailed
    tx.Status = StatusCancelled
    tx.CancelReason = reason
    tx.CancelledAt = timePtr(time.Now())
    tx.UpdatedAt = time.Now()

    return s.repo.Update(ctx, tx)
}

// RetryTransaction retries a failed transaction
func (s *domainService) RetryTransaction(ctx context.Context, id string) error {
    tx, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    if tx.Status != StatusFailed {
        return NewInvalidStateError("transaction must be failed to retry", tx.State)
    }

    if tx.Attempts >= s.config.MaxRetries {
        return NewRetryExceededError(tx.ID, tx.Attempts)
    }

    // Reset state and increment attempts
    tx.State = StateInitiated
    tx.Status = StatusPending
    tx.Attempts++
    tx.ErrorMessage = ""
    tx.UpdatedAt = time.Now()

    return s.repo.Update(ctx, tx)
}

// ProcessState processes the current state of a transaction
func (s *domainService) ProcessState(ctx context.Context, tx *Transaction) (*Transaction, error) {
    if s.stateMachine.IsTerminalState(tx.State) {
        return tx, nil
    }

    nextState, err := s.stateMachine.GetNextState(tx.State)
    if err != nil {
        return nil, err
    }

    return s.stateMachine.Transition(ctx, tx, nextState)
}

// CompleteTransaction marks a transaction as completed
func (s *domainService) CompleteTransaction(ctx context.Context, tx *Transaction) error {
    tx.State = StateCompleted
    tx.Status = StatusCompleted
    tx.CompletedAt = timePtr(time.Now())
    tx.UpdatedAt = time.Now()

    return s.repo.Update(ctx, tx)
}

// FailTransaction marks a transaction as failed
func (s *domainService) FailTransaction(ctx context.Context, tx *Transaction, errorMsg string) error {
    tx.State = StateFailed
    tx.Status = StatusFailed
    tx.ErrorMessage = errorMsg
    tx.FailedAt = timePtr(time.Now())
    tx.UpdatedAt = time.Now()

    return s.repo.Update(ctx, tx)
}

// validatePaymentRequest validates the payment request
func (s *domainService) validatePaymentRequest(req *PaymentRequest) error {
    if req.Amount.IsZero() || req.Amount.IsNegative() {
        return ErrInvalidAmount
    }
    if req.SourceCurrency == "" || req.TargetCurrency == "" {
        return ErrInvalidCurrency
    }
    if req.SourceCountry == "" || req.TargetCountry == "" {
        return ErrInvalidCountry
    }
    if req.SourceAccountID == "" {
        return ErrInvalidAccount
    }
    return nil
}

// timePtr returns a pointer to a time
func timePtr(t time.Time) *time.Time {
    return &t
}