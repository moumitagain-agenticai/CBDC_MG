package revaluator

import (
    "context"
    "time"

    "github.com/apache/fineract-cbdc-accounting/internal/domain/position"
    "github.com/apache/fineract-cbdc-accounting/internal/domain/revaluation"
    "github.com/apache/fineract-cbdc-accounting/internal/infrastructure/client"
    "github.com/apache/fineract-cbdc-accounting/internal/infrastructure/config"

    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// RevaluationEngine handles currency revaluations
type RevaluationEngine struct {
    positionRepo position.Repository
    fxClient     *client.FXClient
    logger       *zap.Logger
    config       *config.RevaluationConfig
}

// NewRevaluationEngine creates a new revaluation engine
func NewRevaluationEngine(
    positionRepo position.Repository,
    fxClient *client.FXClient,
    logger *zap.Logger,
    config *config.RevaluationConfig,
) *RevaluationEngine {
    return &RevaluationEngine{
        positionRepo: positionRepo,
        fxClient:     fxClient,
        logger:       logger,
        config:       config,
    }
}

// RevalueAllCurrencies revalues all currencies
func (e *RevaluationEngine) RevalueAllCurrencies(ctx context.Context, tenantID string) ([]*revaluation.Revaluation, error) {
    e.logger.Info("Starting revaluation for all currencies",
        zap.String("tenant_id", tenantID),
    )

    // Get all positions
    positions, err := e.positionRepo.List(ctx, &position.PositionFilter{
        TenantID: tenantID,
        Status:   position.StatusOpen,
        Limit:    1000,
    })
    if err != nil {
        return nil, err
    }

    var revaluations []*revaluation.Revaluation

    for _, pos := range positions {
        // Get current rate from FX client
        rate, err := e.fxClient.GetRate(ctx, "USD", pos.Currency)
        if err != nil {
            e.logger.Warn("Failed to get rate for currency",
                zap.String("currency", pos.Currency),
                zap.Error(err),
            )
            continue
        }

        // Create revaluation request
        req := &accounting.RevaluationRequest{
            Currency:        pos.Currency,
            TenantID:        tenantID,
            NewRate:         rate,
            RevaluationDate: time.Now(),
            ReferenceID:     "SCHEDULED_REVALUATION",
        }

        // Execute revaluation
        reval, err := e.executeRevaluation(ctx, req)
        if err != nil {
            e.logger.Warn("Failed to revalue currency",
                zap.String("currency", pos.Currency),
                zap.Error(err),
            )
            continue
        }

        revaluations = append(revaluations, reval)
    }

    return revaluations, nil
}