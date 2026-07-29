package service

import (
    "context"
    "fmt"
    "time"

    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/exception"
    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/matching"
    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/reconciliation"
    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/statement"
    "github.com/apache/fineract-cbdc-reconciliation/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-reconciliation/internal/infrastructure/matcher"
    "github.com/apache/fineract-cbdc-reconciliation/internal/infrastructure/parser"
    "github.com/apache/fineract-cbdc-reconciliation/pkg/metrics"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type ReconciliationServiceImpl struct {
    reconciliationRepo reconciliation.Repository
    statementRepo      statement.Repository
    exceptionRepo      exception.Repository
    matchingRepo       matching.Repository
    matchingEngine     *matcher.MatchingEngine
    statementParser    *parser.StatementParser
    logger             *zap.Logger
    config             *config.ReconciliationConfig
}

func NewReconciliationService(
    reconciliationRepo reconciliation.Repository,
    statementRepo statement.Repository,
    exceptionRepo exception.Repository,
    matchingEngine *matcher.MatchingEngine,
    statementParser *parser.StatementParser,
    logger *zap.Logger,
    config *config.ReconciliationConfig,
) reconciliation.Service {
    return &ReconciliationServiceImpl{
        reconciliationRepo: reconciliationRepo,
        statementRepo:      statementRepo,
        exceptionRepo:      exceptionRepo,
        matchingEngine:     matchingEngine,
        statementParser:    statementParser,
        logger:             logger,
        config:             config,
    }
}

// CreateReconciliation creates a new reconciliation
func (s *ReconciliationServiceImpl) CreateReconciliation(ctx context.Context, req *reconciliation.CreateReconciliationRequest) (*reconciliation.Reconciliation, error) {
    startTime := time.Now()
    defer func() {
        metrics.ReconciliationLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate request
    if err := s.validateCreateReconciliationRequest(req); err != nil {
        metrics.ReconciliationErrors.Inc()
        return nil, err
    }

    // Create reconciliation
    rec := &reconciliation.Reconciliation{
        ID:             uuid.New().String(),
        Name:           req.Name,
        Type:           req.Type,
        Status:         reconciliation.StatusPending,
        AccountID:      req.AccountID,
        AccountNumber:  req.AccountNumber,
        Currency:       req.Currency,
        StartDate:      req.StartDate,
        EndDate:        req.EndDate,
        OpeningBalance: req.OpeningBalance,
        ClosingBalance: decimal.Zero,
        SystemBalance:  decimal.Zero,
        BankBalance:    decimal.Zero,
        Difference:     decimal.Zero,
        TotalEntries:   0,
        MatchedEntries: 0,
        UnmatchedEntries: 0,
        TenantID:       req.TenantID,
        Metadata:       req.Metadata,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    if err := s.reconciliationRepo.Create(ctx, rec); err != nil {
        metrics.ReconciliationErrors.Inc()
        return nil, fmt.Errorf("failed to create reconciliation: %w", err)
    }

    metrics.ReconciliationsInitiated.Inc()
    return rec, nil
}

// UploadStatement uploads and processes a bank statement
func (s *ReconciliationServiceImpl) UploadStatement(ctx context.Context, req *reconciliation.UploadStatementRequest) (*statement.BankStatement, error) {
    startTime := time.Now()
    defer func() {
        metrics.StatementUploadLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Get reconciliation
    rec, err := s.reconciliationRepo.GetByID(ctx, req.ReconciliationID)
    if err != nil {
        return nil, err
    }
    if rec == nil {
        return nil, reconciliation.ErrReconciliationNotFound
    }

    // Parse statement
    stmt, err := s.statementParser.Parse(ctx, req.FileContent, req.StatementType)
    if err != nil {
        metrics.StatementUploadErrors.Inc()
        return nil, fmt.Errorf("failed to parse statement: %w", err)
    }

    // Create statement record
    bankStatement := &statement.BankStatement{
        ID:               uuid.New().String(),
        ReconciliationID: req.ReconciliationID,
        AccountID:        rec.AccountID,
        AccountNumber:    rec.AccountNumber,
        Currency:         rec.Currency,
        StatementDate:    stmt.StatementDate,
        StatementType:    req.StatementType,
        OpeningBalance:   stmt.OpeningBalance,
        ClosingBalance:   stmt.ClosingBalance,
        TotalDebit:       stmt.TotalDebit,
        TotalCredit:      stmt.TotalCredit,
        Entries:          stmt.Entries,
        Status:           statement.StatusReceived,
        FileName:         req.FileName,
        FileContent:      req.FileContent,
        Metadata:         req.Metadata,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }

    if err := s.statementRepo.Create(ctx, bankStatement); err != nil {
        metrics.StatementUploadErrors.Inc()
        return nil, fmt.Errorf("failed to save statement: %w", err)
    }

    // Update reconciliation status
    rec.Status = reconciliation.StatusProcessing
    rec.BankBalance = stmt.ClosingBalance
    rec.UpdatedAt = time.Now()

    if err := s.reconciliationRepo.Update(ctx, rec); err != nil {
        s.logger.Warn("Failed to update reconciliation status", zap.Error(err))
    }

    metrics.StatementsUploaded.Inc()
    return bankStatement, nil
}

// ExecuteMatching executes the matching process
func (s *ReconciliationServiceImpl) ExecuteMatching(ctx context.Context, reconciliationID string) (*reconciliation.MatchingResult, error) {
    startTime := time.Now()
    defer func() {
        metrics.MatchingLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Get reconciliation
    rec, err := s.reconciliationRepo.GetByID(ctx, reconciliationID)
    if err != nil {
        return nil, err
    }
    if rec == nil {
        return nil, reconciliation.ErrReconciliationNotFound
    }

    // Get statement
    statements, err := s.statementRepo.List(ctx, &reconciliation.StatementFilter{
        ReconciliationID: reconciliationID,
        Status:           statement.StatusReceived,
        Limit:            1,
    })
    if err != nil {
        return nil, err
    }

    if len(statements) == 0 {
        return nil, reconciliation.ErrStatementNotFound
    }

    statement := statements[0]

    // Execute matching
    results, err := s.matchingEngine.Match(ctx, rec, statement)
    if err != nil {
        metrics.MatchingErrors.Inc()
        return nil, fmt.Errorf("failed to execute matching: %w", err)
    }

    // Update reconciliation with results
    rec.MatchedEntries = results.MatchedEntries
    rec.UnmatchedEntries = results.UnmatchedEntries
    rec.TotalEntries = results.TotalEntries
    rec.SystemBalance = results.SystemBalance
    rec.BankBalance = results.BankBalance
    rec.Difference = results.Difference

    if rec.IsBalanced() {
        rec.Status = reconciliation.StatusCompleted
        now := time.Now()
        rec.CompletedAt = &now
    } else {
        rec.Status = reconciliation.StatusReview
    }

    rec.UpdatedAt = time.Now()

    if err := s.reconciliationRepo.Update(ctx, rec); err != nil {
        s.logger.Warn("Failed to update reconciliation with matching results", zap.Error(err))
    }

    // Create exceptions for unmatched entries
    if err := s.createExceptions(ctx, rec, results.UnmatchedDetails); err != nil {
        s.logger.Warn("Failed to create exceptions", zap.Error(err))
    }

    metrics.MatchesCompleted.Inc()
    return results, nil
}

// createExceptions creates exception items for unmatched entries
func (s *ReconciliationServiceImpl) createExceptions(ctx context.Context, rec *reconciliation.Reconciliation, unmatched []matching.MatchDetail) error {
    for _, detail := range unmatched {
        exception := &exception.ExceptionItem{
            ID:                uuid.New().String(),
            ReconciliationID:  rec.ID,
            Type:              exception.TypeUnmatched,
            Status:            exception.StatusOpen,
            Priority:          "MEDIUM",
            Description:       fmt.Sprintf("Unmatched entry: %s", detail.SystemEntryID),
            Amount:            detail.SystemAmount,
            Currency:          rec.Currency,
            SystemTransactionID: detail.SystemEntryID,
            BankTransactionID: detail.BankEntryID,
            Date:              detail.SystemDate,
            TenantID:          rec.TenantID,
            CreatedAt:         time.Now(),
            UpdatedAt:         time.Now(),
        }

        if err := s.exceptionRepo.Create(ctx, exception); err != nil {
            s.logger.Warn("Failed to create exception", zap.Error(err))
        }
    }

    return nil
}