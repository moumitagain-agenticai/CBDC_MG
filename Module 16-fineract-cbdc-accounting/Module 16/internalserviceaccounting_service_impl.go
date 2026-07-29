package service

import (
    "context"
    "fmt"
    "time"

    "github.com/apache/fineract-cbdc-accounting/internal/domain/ledger"
    "github.com/apache/fineract-cbdc-accounting/internal/domain/position"
    "github.com/apache/fineract-cbdc-accounting/internal/domain/revaluation"
    "github.com/apache/fineract-cbdc-accounting/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-accounting/internal/infrastructure/revaluator"
    "github.com/apache/fineract-cbdc-accounting/pkg/metrics"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type AccountingServiceImpl struct {
    ledgerRepo        ledger.Repository
    positionRepo      position.Repository
    revaluationRepo   revaluation.Repository
    revaluationEngine *revaluator.RevaluationEngine
    logger            *zap.Logger
    config            *config.AccountingConfig
}

func NewAccountingService(
    ledgerRepo ledger.Repository,
    positionRepo position.Repository,
    revaluationRepo revaluation.Repository,
    revaluationEngine *revaluator.RevaluationEngine,
    logger *zap.Logger,
    config *config.AccountingConfig,
) accounting.Service {
    return &AccountingServiceImpl{
        ledgerRepo:        ledgerRepo,
        positionRepo:      positionRepo,
        revaluationRepo:   revaluationRepo,
        revaluationEngine: revaluationEngine,
        logger:            logger,
        config:            config,
    }
}

// PostJournalEntry posts a journal entry to the ledger
func (s *AccountingServiceImpl) PostJournalEntry(ctx context.Context, req *accounting.JournalEntryRequest) (*accounting.JournalEntryResult, error) {
    startTime := time.Now()
    defer func() {
        metrics.JournalEntryLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate request
    if err := s.validateJournalEntryRequest(req); err != nil {
        metrics.JournalEntryErrors.Inc()
        return nil, err
    }

    // Calculate totals
    totalDebit := decimal.Zero
    totalCredit := decimal.Zero
    var entryIDs []string

    // Process each entry in a transaction
    for _, item := range req.Entries {
        // Get account
        account, err := s.ledgerRepo.GetByID(ctx, item.AccountID)
        if err != nil {
            return nil, fmt.Errorf("failed to get account: %w", err)
        }

        if account == nil {
            return nil, ledger.ErrAccountNotFound
        }

        // Create ledger entry
        entry := &ledger.LedgerEntry{
            ID:            uuid.New().String(),
            TransactionID: req.TransactionID,
            AccountID:     item.AccountID,
            AccountCode:   account.Code,
            EntryType:     item.EntryType,
            Amount:        item.Amount,
            Currency:      item.Currency,
            Description:   item.Description,
            ReferenceID:   req.ReferenceID,
            ReferenceType: req.ReferenceType,
            TenantID:      req.TenantID,
            Metadata:      req.Metadata,
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
        }

        // Update account balance
        if item.EntryType == ledger.EntryTypeDebit {
            account.DebitBalance = account.DebitBalance.Add(item.Amount)
            account.Balance = account.Balance.Add(item.Amount)
            totalDebit = totalDebit.Add(item.Amount)
        } else {
            account.CreditBalance = account.CreditBalance.Add(item.Amount)
            account.Balance = account.Balance.Sub(item.Amount)
            totalCredit = totalCredit.Add(item.Amount)
        }

        account.UpdatedAt = time.Now()

        // Save entry and update account
        if err := s.ledgerRepo.CreateEntry(ctx, entry); err != nil {
            return nil, fmt.Errorf("failed to create entry: %w", err)
        }

        if err := s.ledgerRepo.Update(ctx, account); err != nil {
            return nil, fmt.Errorf("failed to update account: %w", err)
        }

        entryIDs = append(entryIDs, entry.ID)

        // Update currency position
        if err := s.updateCurrencyPosition(ctx, req.TenantID, item.Currency, item.EntryType, item.Amount); err != nil {
            s.logger.Warn("Failed to update currency position", zap.Error(err))
        }
    }

    // Verify balance
    isBalanced := totalDebit.Equals(totalCredit)

    if !isBalanced {
        s.logger.Warn("Journal entry not balanced",
            zap.String("transaction_id", req.TransactionID),
            zap.String("total_debit", totalDebit.String()),
            zap.String("total_credit", totalCredit.String()),
        )
    }

    metrics.JournalEntriesPosted.Inc()

    return &accounting.JournalEntryResult{
        EntryIDs:    entryIDs,
        TransactionID: req.TransactionID,
        TotalDebit:  totalDebit,
        TotalCredit: totalCredit,
        IsBalanced:  isBalanced,
        CreatedAt:   time.Now(),
    }, nil
}

// updateCurrencyPosition updates the currency position
func (s *AccountingServiceImpl) updateCurrencyPosition(ctx context.Context, tenantID, currency string, entryType ledger.EntryType, amount decimal.Decimal) error {
    position, err := s.positionRepo.GetByCurrency(ctx, currency, tenantID)
    if err != nil {
        return err
    }

    if position == nil {
        // Create new position
        position = &position.CurrencyPosition{
            ID:            uuid.New().String(),
            Currency:      currency,
            TenantID:      tenantID,
            LongPosition:  decimal.Zero,
            ShortPosition: decimal.Zero,
            NetPosition:   decimal.Zero,
            TotalInflow:   decimal.Zero,
            TotalOutflow:  decimal.Zero,
            Status:        position.StatusOpen,
            LastUpdated:   time.Now(),
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
        }
    }

    if entryType == ledger.EntryTypeDebit {
        position.LongPosition = position.LongPosition.Add(amount)
        position.TotalInflow = position.TotalInflow.Add(amount)
    } else {
        position.ShortPosition = position.ShortPosition.Add(amount)
        position.TotalOutflow = position.TotalOutflow.Add(amount)
    }

    position.NetPosition = position.LongPosition.Sub(position.ShortPosition)
    position.LastUpdated = time.Now()
    position.UpdatedAt = time.Now()

    return s.positionRepo.Update(ctx, position)
}

// RevalueCurrency revalues a currency
func (s *AccountingServiceImpl) RevalueCurrency(ctx context.Context, req *accounting.RevaluationRequest) (*revaluation.Revaluation, error) {
    startTime := time.Now()
    defer func() {
        metrics.RevaluationLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Get current position
    position, err := s.positionRepo.GetByCurrency(ctx, req.Currency, req.TenantID)
    if err != nil {
        return nil, err
    }

    if position == nil {
        return nil, fmt.Errorf("no position found for currency %s", req.Currency)
    }

    oldRate := position.RevaluationRate
    newRate := req.NewRate

    // Calculate gain/loss
    netPosition := position.GetNetPosition()
    gainLoss := netPosition.Mul(newRate.Sub(oldRate))
    gainLossType := revaluation.GainType
    if gainLoss.IsNegative() {
        gainLossType = revaluation.LossType
        gainLoss = gainLoss.Abs()
    }

    // Create revaluation record
    reval := &revaluation.Revaluation{
        ID:              uuid.New().String(),
        Currency:        req.Currency,
        TenantID:        req.TenantID,
        OldRate:         oldRate,
        NewRate:         newRate,
        OldPosition:     netPosition,
        NewPosition:     netPosition,
        GainLoss:        gainLoss,
        GainLossType:    gainLossType,
        Status:          revaluation.StatusCompleted,
        RevaluationDate: req.RevaluationDate,
        ReferenceID:     req.ReferenceID,
        Metadata:        req.Metadata,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    if err := s.revaluationRepo.Create(ctx, reval); err != nil {
        metrics.RevaluationErrors.Inc()
        return nil, fmt.Errorf("failed to create revaluation: %w", err)
    }

    // Update position
    position.RevaluationRate = newRate
    position.RevaluationGain = position.RevaluationGain.Add(gainLoss)
    position.LastUpdated = time.Now()
    position.UpdatedAt = time.Now()

    if err := s.positionRepo.Update(ctx, position); err != nil {
        s.logger.Warn("Failed to update position after revaluation", zap.Error(err))
    }

    metrics.RevaluationsCompleted.Inc()

    return reval, nil
}

// validateJournalEntryRequest validates a journal entry request
func (s *AccountingServiceImpl) validateJournalEntryRequest(req *accounting.JournalEntryRequest) error {
    if req.TransactionID == "" {
        return accounting.ErrTransactionIDRequired
    }
    if len(req.Entries) == 0 {
        return accounting.ErrNoEntries
    }
    if req.TenantID == "" {
        return accounting.ErrTenantIDRequired
    }
    return nil
}