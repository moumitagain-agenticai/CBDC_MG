package service

import (
    "context"
    "fmt"
    "time"

    "github.com/apache/fineract-cbdc-settlement/internal/domain/burn"
    "github.com/apache/fineract-cbdc-settlement/internal/domain/compensation"
    "github.com/apache/fineract-cbdc-settlement/internal/domain/lock"
    "github.com/apache/fineract-cbdc-settlement/internal/domain/settlement"
    "github.com/apache/fineract-cbdc-settlement/internal/infrastructure/client"
    "github.com/apache/fineract-cbdc-settlement/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-settlement/internal/infrastructure/coordinator"
    "github.com/apache/fineract-cbdc-settlement/pkg/metrics"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type SettlementServiceImpl struct {
    settlementRepo  settlement.Repository
    lockRepo        lock.Repository
    burnRepo        burn.Repository
    compensationRepo compensation.Repository
    coordinator     *coordinator.TwoPhaseCoordinator
    logger          *zap.Logger
    config          *config.SettlementConfig
}

func NewSettlementService(
    settlementRepo settlement.Repository,
    lockRepo lock.Repository,
    burnRepo burn.Repository,
    compensationRepo compensation.Repository,
    coordinator *coordinator.TwoPhaseCoordinator,
    logger *zap.Logger,
    config *config.SettlementConfig,
) settlement.Service {
    return &SettlementServiceImpl{
        settlementRepo:  settlementRepo,
        lockRepo:        lockRepo,
        burnRepo:        burnRepo,
        compensationRepo: compensationRepo,
        coordinator:     coordinator,
        logger:          logger,
        config:          config,
    }
}

// InitiateSettlement initiates a new atomic settlement
func (s *SettlementServiceImpl) InitiateSettlement(ctx context.Context, req *settlement.SettlementRequest) (*settlement.Settlement, error) {
    startTime := time.Now()
    defer func() {
        metrics.SettlementLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate request
    if err := s.validateSettlementRequest(req); err != nil {
        metrics.SettlementErrors.Inc()
        return nil, err
    }

    // Create settlement record
    settlement := &settlement.Settlement{
        ID:              uuid.New().String(),
        TransactionID:   req.TransactionID,
        SourceNetwork:   req.SourceNetwork,
        TargetNetwork:   req.TargetNetwork,
        SourceAccountID: req.SourceAccountID,
        TargetAccountID: req.TargetAccountID,
        SourceCurrency:  req.SourceCurrency,
        TargetCurrency:  req.TargetCurrency,
        SourceAmount:    req.SourceAmount,
        TargetAmount:    req.TargetAmount,
        ConversionRate:  req.ConversionRate,
        Status:          settlement.StatusPending,
        Type:            req.Type,
        RetryCount:      0,
        Metadata:        req.Metadata,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Save settlement
    if err := s.settlementRepo.Create(ctx, settlement); err != nil {
        metrics.SettlementErrors.Inc()
        return nil, fmt.Errorf("failed to create settlement: %w", err)
    }

    // Execute settlement asynchronously
    go func() {
        s.executeSettlement(context.Background(), settlement)
    }()

    metrics.SettlementsInitiated.Inc()
    return settlement, nil
}

// executeSettlement executes the atomic settlement process
func (s *SettlementServiceImpl) executeSettlement(ctx context.Context, settlement *settlement.Settlement) {
    s.logger.Info("Executing settlement",
        zap.String("settlement_id", settlement.ID),
        zap.String("transaction_id", settlement.TransactionID),
    )

    // Execute two-phase commit
    result, err := s.coordinator.ExecuteTwoPhaseCommit(ctx, settlement)
    if err != nil {
        s.logger.Error("Settlement failed",
            zap.String("settlement_id", settlement.ID),
            zap.Error(err),
        )

        // Update settlement status
        settlement.Status = settlement.StatusFailed
        settlement.ErrorMessage = err.Error()
        now := time.Now()
        settlement.FailedAt = &now
        settlement.UpdatedAt = now

        if updateErr := s.settlementRepo.Update(ctx, settlement); updateErr != nil {
            s.logger.Error("Failed to update settlement status",
                zap.String("settlement_id", settlement.ID),
                zap.Error(updateErr),
            )
        }

        // Check if compensation is needed
        if result != nil && result.BurnCompleted && !result.IssueCompleted {
            s.logger.Warn("Burn completed but issue failed - initiating compensation",
                zap.String("settlement_id", settlement.ID),
                zap.String("burn_tx", result.BurnTransactionID),
            )

            // Initiate compensation
            compReq := &settlement.CompensationRequest{
                SettlementID:   settlement.ID,
                Network:        settlement.SourceNetwork,
                AccountID:      settlement.SourceAccountID,
                Amount:         settlement.SourceAmount,
                Currency:       settlement.SourceCurrency,
                OriginalBurnTx: result.BurnTransactionID,
                Reason:         "Target issuance failed - automatic compensation",
            }

            if _, compErr := s.CompensateBurn(ctx, compReq); compErr != nil {
                s.logger.Error("Compensation failed - manual intervention required",
                    zap.String("settlement_id", settlement.ID),
                    zap.Error(compErr),
                )
                // Send critical alert
                s.sendCriticalAlert(settlement.ID, result.BurnTransactionID, compErr)
            }
        }

        metrics.SettlementErrors.Inc()
        return
    }

    // Update settlement status
    settlement.Status = settlement.StatusCompleted
    settlement.SourceLockID = result.SourceLockID
    settlement.TargetLockID = result.TargetLockID
    settlement.BurnTransactionID = result.BurnTransactionID
    settlement.IssueTransactionID = result.IssueTransactionID
    now := time.Now()
    settlement.CompletedAt = &now
    settlement.UpdatedAt = now

    if err := s.settlementRepo.Update(ctx, settlement); err != nil {
        s.logger.Error("Failed to update settlement status",
            zap.String("settlement_id", settlement.ID),
            zap.Error(err),
        )
    }

    metrics.SettlementsCompleted.Inc()
    s.logger.Info("Settlement completed successfully",
        zap.String("settlement_id", settlement.ID),
        zap.String("burn_tx", result.BurnTransactionID),
        zap.String("issue_tx", result.IssueTransactionID),
    )
}

// LockFunds locks funds on a CBDC network
func (s *SettlementServiceImpl) LockFunds(ctx context.Context, req *settlement.LockRequest) (*lock.FundLock, error) {
    // Validate request
    if err := s.validateLockRequest(req); err != nil {
        return nil, err
    }

    // Create lock record
    fundLock := &lock.FundLock{
        ID:           uuid.New().String(),
        SettlementID: req.SettlementID,
        Network:      req.Network,
        AccountID:    req.AccountID,
        LockID:       fmt.Sprintf("lock_%s_%d", req.SettlementID, time.Now().UnixNano()),
        Amount:       req.Amount,
        Currency:     req.Currency,
        Status:       lock.LockStatusActive,
        LockDuration: req.Duration,
        ExpiresAt:    time.Now().Add(req.Duration),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
        Metadata:     req.Metadata,
    }

    // Save lock
    if err := s.lockRepo.Create(ctx, fundLock); err != nil {
        return nil, fmt.Errorf("failed to create lock: %w", err)
    }

    metrics.LocksCreated.Inc()
    return fundLock, nil
}

// ReleaseLock releases a locked fund
func (s *SettlementServiceImpl) ReleaseLock(ctx context.Context, lockID string) error {
    fundLock, err := s.lockRepo.GetByLockID(ctx, lockID)
    if err != nil {
        return err
    }

    if fundLock == nil {
        return lock.ErrLockNotFound
    }

    if fundLock.Status != lock.LockStatusActive {
        return lock.ErrLockNotActive
    }

    fundLock.Status = lock.LockStatusReleased
    now := time.Now()
    fundLock.ReleasedAt = &now
    fundLock.UpdatedAt = now

    if err := s.lockRepo.Update(ctx, fundLock); err != nil {
        return fmt.Errorf("failed to release lock: %w", err)
    }

    metrics.LocksReleased.Inc()
    return nil}

// BurnFunds burns funds on a CBDC network
func (s *SettlementServiceImpl) BurnFunds(ctx context.Context, req *settlement.BurnRequest) (*burn.BurnRecord, error) {
    // Validate request
    if err := s.validateBurnRequest(req); err != nil {
        return nil, err
    }

    // Create burn record
    burnRecord := &burn.BurnRecord{
        ID:            uuid.New().String(),
        SettlementID:  req.SettlementID,
        Network:       req.Network,
        LockID:        req.LockID,
        TransactionID: fmt.Sprintf("burn_%s_%d", req.SettlementID, time.Now().UnixNano()),
        Amount:        req.Amount,
        Currency:      req.Currency,
        Status:        burn.BurnStatusPending,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
        Metadata:      req.Metadata,
    }

    // Save burn record
    if err := s.burnRepo.Create(ctx, burnRecord); err != nil {
        return nil, fmt.Errorf("failed to create burn record: %w", err)
    }

    metrics.BurnsInitiated.Inc()
    return burnRecord, nil
}

// CompensateBurn compensates a failed burn by re-issuing funds
func (s *SettlementServiceImpl) CompensateBurn(ctx context.Context, req *settlement.CompensationRequest) (*compensation.Compensation, error) {
    s.logger.Warn("Initiating burn compensation",
        zap.String("settlement_id", req.SettlementID),
        zap.String("original_burn_tx", req.OriginalBurnTx),
        zap.String("amount", req.Amount.String()),
    )

    // Create compensation record
    comp := &compensation.Compensation{
        ID:              uuid.New().String(),
        SettlementID:    req.SettlementID,
        OriginalBurnTx:  req.OriginalBurnTx,
        Network:         req.Network,
        AccountID:       req.AccountID,
        Amount:          req.Amount,
        Currency:        req.Currency,
        Status:          compensation.CompStatusPending,
        Reason:          req.Reason,
        AlertSent:       false,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
        Metadata:        req.Metadata,
    }

    if err := s.compensationRepo.Create(ctx, comp); err != nil {
        return nil, fmt.Errorf("failed to create compensation record: %w", err)
    }

    // In production, this would actually re-issue the funds
    // For now, we mark it as completed (would call CBDC connector)
    comp.Status = compensation.CompStatusCompleted
    comp.CompensationTx = fmt.Sprintf("comp_%s_%d", req.SettlementID, time.Now().UnixNano())
    now := time.Now()
    comp.ResolvedAt = &now
    comp.UpdatedAt = now

    if err := s.compensationRepo.Update(ctx, comp); err != nil {
        return nil, fmt.Errorf("failed to update compensation: %w", err)
    }

    metrics.CompensationsCompleted.Inc()
    return comp, nil
}

// Helper functions
func (s *SettlementServiceImpl) validateSettlementRequest(req *settlement.SettlementRequest) error {
    if req.TransactionID == "" {
        return settlement.ErrTransactionIDRequired
    }
    if req.SourceNetwork == "" || req.TargetNetwork == "" {
        return settlement.ErrNetworkRequired
    }
    if req.SourceAccountID == "" || req.TargetAccountID == "" {
        return settlement.ErrAccountRequired
    }
    if req.SourceAmount.IsZero() || req.SourceAmount.IsNegative() {
        return settlement.ErrInvalidAmount
    }
    if req.TargetAmount.IsZero() || req.TargetAmount.IsNegative() {
        return settlement.ErrInvalidAmount
    }
    return nil
}

func (s *SettlementServiceImpl) validateLockRequest(req *settlement.LockRequest) error {
    if req.SettlementID == "" {
        return lock.ErrSettlementIDRequired
    }
    if req.Network == "" {
        return lock.ErrNetworkRequired
    }
    if req.AccountID == "" {
        return lock.ErrAccountRequired
    }
    if req.Amount.IsZero() || req.Amount.IsNegative() {
        return lock.ErrInvalidAmount
    }
    if req.Duration <= 0 {
        return lock.ErrInvalidDuration
    }
    return nil
}

func (s *SettlementServiceImpl) validateBurnRequest(req *settlement.BurnRequest) error {
    if req.SettlementID == "" {
        return burn.ErrSettlementIDRequired
    }
    if req.Network == "" {
        return burn.ErrNetworkRequired
    }
    if req.LockID == "" {
        return burn.ErrLockIDRequired
    }
    if req.Amount.IsZero() || req.Amount.IsNegative() {
        return burn.ErrInvalidAmount
    }
    return nil
}

func (s *SettlementServiceImpl) sendCriticalAlert(settlementID, burnTx string, err error) {
    // In production, this would send to PagerDuty/Opsgenie
    s.logger.Error("CRITICAL: Compensation failed - manual intervention required",
        zap.String("settlement_id", settlementID),
        zap.String("burn_transaction", burnTx),
        zap.Error(err),
    )
}