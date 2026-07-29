package coordinator

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/apache/fineract-cbdc-settlement/internal/domain/settlement"
    "github.com/apache/fineract-cbdc-settlement/internal/infrastructure/client"
    "github.com/apache/fineract-cbdc-settlement/internal/infrastructure/config"

    "go.uber.org/zap"
)

// TwoPhaseCommitResult represents the result of a two-phase commit
type TwoPhaseCommitResult struct {
    SourceLockID      string
    TargetLockID      string
    BurnTransactionID string
    IssueTransactionID string
    BurnCompleted     bool
    IssueCompleted    bool
}

// TwoPhaseCoordinator coordinates two-phase commit for atomic settlements
type TwoPhaseCoordinator struct {
    indiaConnector *client.IndiaConnector
    uaeConnector   *client.UAEConnector
    fabricClient   *client.FabricClient
    logger         *zap.Logger
    config         *config.CoordinatorConfig
    mu             sync.RWMutex
}

// NewTwoPhaseCoordinator creates a new two-phase coordinator
func NewTwoPhaseCoordinator(
    indiaConnector *client.IndiaConnector,
    uaeConnector *client.UAEConnector,
    fabricClient *client.FabricClient,
    logger *zap.Logger,
    config *config.CoordinatorConfig,
) *TwoPhaseCoordinator {
    return &TwoPhaseCoordinator{
        indiaConnector: indiaConnector,
        uaeConnector:   uaeConnector,
        fabricClient:   fabricClient,
        logger:         logger,
        config:         config,
    }
}

// ExecuteTwoPhaseCommit executes a two-phase commit for atomic settlement
func (c *TwoPhaseCoordinator) ExecuteTwoPhaseCommit(ctx context.Context, settlement *settlement.Settlement) (*TwoPhaseCommitResult, error) {
    result := &TwoPhaseCommitResult{}

    c.logger.Info("Starting two-phase commit",
        zap.String("settlement_id", settlement.ID),
        zap.String("source", settlement.SourceNetwork),
        zap.String("target", settlement.TargetNetwork),
    )

    // Phase 1: Lock Funds (Target first, then Source)
    targetLock, err := c.lockTargetFunds(ctx, settlement)
    if err != nil {
        return result, fmt.Errorf("phase 1 failed - target lock: %w", err)
    }
    result.TargetLockID = targetLock

    sourceLock, err := c.lockSourceFunds(ctx, settlement)
    if err != nil {
        // Rollback target lock
        c.rollbackLock(ctx, settlement.TargetNetwork, targetLock)
        return result, fmt.Errorf("phase 1 failed - source lock: %w", err)
    }
    result.SourceLockID = sourceLock

    // Phase 2: Burn Source Funds
    burnTx, err := c.burnSourceFunds(ctx, settlement)
    if err != nil {
        // Rollback both locks
        c.rollbackLock(ctx, settlement.TargetNetwork, targetLock)
        c.rollbackLock(ctx, settlement.SourceNetwork, sourceLock)
        return result, fmt.Errorf("phase 2 failed - burn: %w", err)
    }
    result.BurnTransactionID = burnTx
    result.BurnCompleted = true

    // Phase 3: Issue Target Funds
    issueTx, err := c.issueTargetFunds(ctx, settlement)
    if err != nil {
        // Critical: Burn completed but issue failed - need compensation
        c.logger.Error("Burn completed but issue failed - compensation needed",
            zap.String("settlement_id", settlement.ID),
            zap.String("burn_tx", burnTx),
        )
        result.IssueCompleted = false

        // Rollback locks (target lock already used for issue, but release source)
        c.rollbackLock(ctx, settlement.SourceNetwork, sourceLock)

        return result, fmt.Errorf("phase 3 failed - issue: %w", err)
    }
    result.IssueTransactionID = issueTx
    result.IssueCompleted = true

    // Phase 4: Release Locks (Success path)
    c.releaseLock(ctx, settlement.TargetNetwork, targetLock)
    c.releaseLock(ctx, settlement.SourceNetwork, sourceLock)

    // Phase 5: Record on Fabric
    c.recordOnFabric(ctx, settlement, result)

    c.logger.Info("Two-phase commit completed successfully",
        zap.String("settlement_id", settlement.ID),
        zap.String("burn_tx", burnTx),
        zap.String("issue_tx", issueTx),
    )

    return result, nil
}

// lockTargetFunds locks funds on the target network
func (c *TwoPhaseCoordinator) lockTargetFunds(ctx context.Context, settlement *settlement.Settlement) (string, error) {
    // In production, this would call the actual CBDC connector
    lockID := fmt.Sprintf("lock_target_%s_%d", settlement.ID, time.Now().UnixNano())
    return lockID, nil
}

// lockSourceFunds locks funds on the source network
func (c *TwoPhaseCoordinator) lockSourceFunds(ctx context.Context, settlement *settlement.Settlement) (string, error) {
    lockID := fmt.Sprintf("lock_source_%s_%d", settlement.ID, time.Now().UnixNano())
    return lockID, nil
}

// burnSourceFunds burns funds on the source network
func (c *TwoPhaseCoordinator) burnSourceFunds(ctx context.Context, settlement *settlement.Settlement) (string, error) {
    burnTx := fmt.Sprintf("burn_%s_%d", settlement.ID, time.Now().UnixNano())
    return burnTx, nil
}

// issueTargetFunds issues funds on the target network
func (c *TwoPhaseCoordinator) issueTargetFunds(ctx context.Context, settlement *settlement.Settlement) (string, error) {
    issueTx := fmt.Sprintf("issue_%s_%d", settlement.ID, time.Now().UnixNano())
    return issueTx, nil
}

// rollbackLock rolls back a lock
func (c *TwoPhaseCoordinator) rollbackLock(ctx context.Context, network, lockID string) {
    c.logger.Warn("Rolling back lock",
        zap.String("network", network),
        zap.String("lock_id", lockID),
    )
    // In production, call connector.ReleaseLock()
}

// releaseLock releases a lock
func (c *TwoPhaseCoordinator) releaseLock(ctx context.Context, network, lockID string) {
    c.logger.Info("Releasing lock",
        zap.String("network", network),
        zap.String("lock_id", lockID),
    )
    // In production, call connector.ReleaseLock()
}

// recordOnFabric records the settlement on Fabric
func (c *TwoPhaseCoordinator) recordOnFabric(ctx context.Context, settlement *settlement.Settlement, result *TwoPhaseCommitResult) {
    record := map[string]interface{}{
        "settlement_id":     settlement.ID,
        "transaction_id":    settlement.TransactionID,
        "source_network":    settlement.SourceNetwork,
        "target_network":    settlement.TargetNetwork,
        "source_lock_id":    result.SourceLockID,
        "target_lock_id":    result.TargetLockID,
        "burn_transaction":  result.BurnTransactionID,
        "issue_transaction": result.IssueTransactionID,
        "source_amount":     settlement.SourceAmount.String(),
        "target_amount":     settlement.TargetAmount.String(),
        "source_currency":   settlement.SourceCurrency,
        "target_currency":   settlement.TargetCurrency,
        "timestamp":         time.Now().Unix(),
        "status":            "COMPLETED",
    }

    if err := c.fabricClient.WriteRecord(ctx, "Settlement", record); err != nil {
        c.logger.Error("Failed to record on Fabric",
            zap.Error(err),
            zap.String("settlement_id", settlement.ID),
        )
    }
}