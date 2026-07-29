package service

import (
    "context"
    "fmt"
    "time"

    "github.com/fineract/cbdc/india-connector/internal/domain"
    "github.com/fineract/cbdc/india-connector/internal/ports"
    "github.com/fineract/cbdc/india-connector/pkg/metrics"
    "go.uber.org/zap"
)

// ConnectorService implements the core connector business logic
type ConnectorService struct {
    cbdcClient     ports.CBDCClient
    fineractClient ports.FineractClient
    repo           ports.TransactionRepository
    logger         *zap.Logger
    config         interface {
        GetRetryConfig() (maxAttempts int, initialDelay, maxDelay time.Duration, multiplier float64)
    }
}

// NewConnectorService creates a new connector service
func NewConnectorService(
    cbdcClient ports.CBDCClient,
    fineractClient ports.FineractClient,
    repo ports.TransactionRepository,
    logger *zap.Logger,
    config interface {
        GetRetryConfig() (maxAttempts int, initialDelay, maxDelay time.Duration, multiplier float64)
    },
) *ConnectorService {
    return &ConnectorService{
        cbdcClient:     cbdcClient,
        fineractClient: fineractClient,
        repo:           repo,
        logger:         logger,
        config:         config,
    }
}

// IssueCBDC issues new CBDC tokens
func (s *ConnectorService) IssueCBDC(ctx context.Context, req *domain.TransactionRequest) (*domain.TransactionResponse, error) {
    start := time.Now()
    s.logger.Info("issuing CBDC tokens",
        zap.String("wallet", req.SourceWallet),
        zap.String("amount", req.Amount),
        zap.String("currency", req.Currency),
    )

    // Create transaction record
    tx := domain.NewTransaction(req)
    if err := s.repo.Save(ctx, tx); err != nil {
        return nil, domain.NewDomainError(domain.ErrorInternal, "failed to save transaction")
    }

    // Prepare CBDC request
    cbdcReq := &ports.IssueRequest{
        WalletID:   req.SourceWallet,
        Amount:     req.Amount,
        Currency:   req.Currency,
        ReferenceID: tx.TransactionID,
    }

    // Execute with retry
    var cbdcResp *ports.IssueResponse
    err := s.executeWithRetry(ctx, func() error {
        var err error
        cbdcResp, err = s.cbdcClient.Issue(ctx, cbdcReq)
        return err
    })

    if err != nil {
        tx.Status = domain.TransactionStatusFailed
        tx.ErrorMessage = err.Error()
        s.repo.Update(ctx, tx)
        metrics.RecordTransaction("issue", "failed", time.Since(start))
        return nil, err
    }

    // Update transaction
    tx.Status = domain.TransactionStatusCompleted
    tx.TransactionID = cbdcResp.TransactionID
    now := time.Now()
    tx.CompletedAt = &now
    s.repo.Update(ctx, tx)

    // Notify Fineract
    if err := s.fineractClient.NotifyTransaction(ctx, tx); err != nil {
        s.logger.Warn("failed to notify Fineract", zap.Error(err))
    }

    metrics.RecordTransaction("issue", "success", time.Since(start))
    return &domain.TransactionResponse{
        TransactionID: cbdcResp.TransactionID,
        Status:        string(tx.Status),
        Message:       "CBDC tokens issued successfully",
    }, nil
}

// TransferCBDC transfers CBDC tokens between wallets
func (s *ConnectorService) TransferCBDC(ctx context.Context, req *domain.TransactionRequest) (*domain.TransactionResponse, error) {
    start := time.Now()
    s.logger.Info("transferring CBDC tokens",
        zap.String("source", req.SourceWallet),
        zap.String("destination", req.DestinationWallet),
        zap.String("amount", req.Amount),
        zap.String("currency", req.Currency),
    )

    // Create transaction record
    tx := domain.NewTransaction(req)
    if err := s.repo.Save(ctx, tx); err != nil {
        return nil, domain.NewDomainError(domain.ErrorInternal, "failed to save transaction")
    }

    // Prepare CBDC request
    cbdcReq := &ports.TransferRequest{
        SourceWallet:      req.SourceWallet,
        DestinationWallet: req.DestinationWallet,
        Amount:            req.Amount,
        Currency:          req.Currency,
        ReferenceID:       tx.TransactionID,
    }

    // Execute with retry
    var cbdcResp *ports.TransferResponse
    err := s.executeWithRetry(ctx, func() error {
        var err error
        cbdcResp, err = s.cbdcClient.Transfer(ctx, cbdcReq)
        return err
    })

    if err != nil {
        tx.Status = domain.TransactionStatusFailed
        tx.ErrorMessage = err.Error()
        s.repo.Update(ctx, tx)
        metrics.RecordTransaction("transfer", "failed", time.Since(start))
        return nil, err
    }

    // Update transaction
    tx.Status = domain.TransactionStatusCompleted
    tx.TransactionID = cbdcResp.TransactionID
    now := time.Now()
    tx.CompletedAt = &now
    s.repo.Update(ctx, tx)

    // Notify Fineract
    if err := s.fineractClient.NotifyTransaction(ctx, tx); err != nil {
        s.logger.Warn("failed to notify Fineract", zap.Error(err))
    }

    metrics.RecordTransaction("transfer", "success", time.Since(start))
    return &domain.TransactionResponse{
        TransactionID: cbdcResp.TransactionID,
        Status:        string(tx.Status),
        Message:       "CBDC tokens transferred successfully",
    }, nil
}

// LockCBDC locks CBDC tokens
func (s *ConnectorService) LockCBDC(ctx context.Context, req *domain.TransactionRequest, durationSec int) (*domain.TransactionResponse, error) {
    start := time.Now()
    s.logger.Info("locking CBDC tokens",
        zap.String("wallet", req.SourceWallet),
        zap.String("amount", req.Amount),
        zap.String("currency", req.Currency),
        zap.Int("duration_sec", durationSec),
    )

    // Create transaction record
    tx := domain.NewTransaction(req)
    if err := s.repo.Save(ctx, tx); err != nil {
        return nil, domain.NewDomainError(domain.ErrorInternal, "failed to save transaction")
    }

    // Prepare CBDC request
    cbdcReq := &ports.LockRequest{
        WalletID:    req.SourceWallet,
        Amount:      req.Amount,
        Currency:    req.Currency,
        DurationSec: durationSec,
        ReferenceID: tx.TransactionID,
    }

    // Execute with retry
    var cbdcResp *ports.LockResponse
    err := s.executeWithRetry(ctx, func() error {
        var err error
        cbdcResp, err = s.cbdcClient.Lock(ctx, cbdcReq)
        return err
    })

    if err != nil {
        tx.Status = domain.TransactionStatusFailed
        tx.ErrorMessage = err.Error()
        s.repo.Update(ctx, tx)
        metrics.RecordTransaction("lock", "failed", time.Since(start))
        return nil, err
    }

    // Update transaction
    tx.Status = domain.TransactionStatusCompleted
    tx.TransactionID = cbdcResp.TransactionID
    now := time.Now()
    tx.CompletedAt = &now
    s.repo.Update(ctx, tx)

    metrics.RecordTransaction("lock", "success", time.Since(start))
    return &domain.TransactionResponse{
        TransactionID: cbdcResp.TransactionID,
        Status:        string(tx.Status),
        Message:       fmt.Sprintf("CBDC tokens locked until %s", cbdcResp.ExpiresAt),
        Details: map[string]interface{}{
            "lock_id": cbdcResp.LockID,
            "expires_at": cbdcResp.ExpiresAt,
        },
    }, nil
}

// executeWithRetry executes an operation with retry logic
func (s *ConnectorService) executeWithRetry(ctx context.Context, fn func() error) error {
    maxAttempts, initialDelay, maxDelay, multiplier := s.config.GetRetryConfig()

    var lastErr error
    delay := initialDelay

    for attempt := 0; attempt < maxAttempts; attempt++ {
        if attempt > 0 {
            s.logger.Debug("retrying operation",
                zap.Int("attempt", attempt+1),
                zap.Int("max_attempts", maxAttempts),
                zap.Duration("delay", delay),
            )
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * multiplier)
            if delay > maxDelay {
                delay = maxDelay
            }
        }

        err := fn()
        if err == nil {
            return nil
        }

        // Check if error is retryable
        if !s.isRetryableError(err) {
            return err
        }

        lastErr = err
    }

    return domain.NewDomainError(domain.ErrorRetryExhausted, "retry limit exceeded")
}

// isRetryableError checks if an error is retryable
func (s *ConnectorService) isRetryableError(err error) bool {
    if domainErr, ok := err.(*domain.DomainError); ok {
        switch domainErr.Code {
        case domain.ErrorTransactionTimeout,
            domain.ErrorRateLimit,
            domain.ErrorTransactionFailed:
            return true
        }
    }
    return false
}