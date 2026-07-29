package service

import (
    "context"
    "fmt"
    "sort"
    "time"

    "github.com/apache/fineract-cbdc-fee/internal/domain/calculation"
    "github.com/apache/fineract-cbdc-fee/internal/domain/corridor"
    "github.com/apache/fineract-cbdc-fee/internal/domain/fee"
    "github.com/apache/fineract-cbdc-fee/internal/infrastructure/cache"
    "github.com/apache/fineract-cbdc-fee/internal/infrastructure/calculator"
    "github.com/apache/fineract-cbdc-fee/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-fee/pkg/metrics"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type FeeServiceImpl struct {
    feeRepo           fee.Repository
    calculationRepo   calculation.Repository
    corridorRepo      corridor.Repository
    feeCache          cache.FeeCache
    calculatorFactory *calculator.CalculatorFactory
    logger            *zap.Logger
    config            *config.FeeConfig
}

func NewFeeService(
    feeRepo fee.Repository,
    calculationRepo calculation.Repository,
    corridorRepo corridor.Repository,
    feeCache cache.FeeCache,
    calculatorFactory *calculator.CalculatorFactory,
    logger *zap.Logger,
    config *config.FeeConfig,
) fee.Service {
    return &FeeServiceImpl{
        feeRepo:           feeRepo,
        calculationRepo:   calculationRepo,
        corridorRepo:      corridorRepo,
        feeCache:          feeCache,
        calculatorFactory: calculatorFactory,
        logger:            logger,
        config:            config,
    }
}

// CalculateFee calculates fees for a transaction
func (s *FeeServiceImpl) CalculateFee(ctx context.Context, req *fee.CalculationRequest) (*fee.CalculationResult, error) {
    startTime := time.Now()
    defer func() {
        metrics.FeeCalculationLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate request
    if err := s.validateCalculationRequest(req); err != nil {
        metrics.FeeCalculationErrors.Inc()
        return nil, err
    }

    // Get corridor
    var corridor *corridor.FeeCorridor
    var err error

    if req.CorridorCode != "" {
        corridor, err = s.corridorRepo.GetByCode(ctx, req.CorridorCode)
        if err != nil {
            metrics.FeeCalculationErrors.Inc()
            return nil, fmt.Errorf("failed to get corridor: %w", err)
        }
    } else {
        // Find matching corridor
        corridor, err = s.findMatchingCorridor(ctx, req)
        if err != nil {
            metrics.FeeCalculationErrors.Inc()
            return nil, fmt.Errorf("failed to find matching corridor: %w", err)
        }
    }

    // Get applicable fees
    var fees []*fee.Fee

    if len(req.FeeCodes) > 0 {
        fees, err = s.getFeesByCodes(ctx, req.FeeCodes)
        if err != nil {
            metrics.FeeCalculationErrors.Inc()
            return nil, fmt.Errorf("failed to get fees: %w", err)
        }
    } else {
        fees, err = s.findApplicableFees(ctx, req, corridor)
        if err != nil {
            metrics.FeeCalculationErrors.Inc()
            return nil, fmt.Errorf("failed to find applicable fees: %w", err)
        }
    }

    // Sort fees by priority
    sort.Slice(fees, func(i, j int) bool {
        return fees[i].Priority < fees[j].Priority
    })

    // Calculate each fee
    totalFee := decimal.Zero
    feeBreakdown := []fee.FeeBreakdown{}

    for _, f := range fees {
        // Check cache first
        var feeAmount decimal.Decimal
        var err error

        cachedAmount, err := s.feeCache.Get(ctx, f.ID, req.Amount)
        if err == nil && cachedAmount != nil {
            feeAmount = *cachedAmount
        } else {
            // Calculate fee
            feeAmount, err = f.Calculate(req.Amount)
            if err != nil {
                s.logger.Warn("Failed to calculate fee",
                    zap.String("fee_id", f.ID),
                    zap.Error(err),
                )
                continue
            }

            // Cache fee calculation
            if err := s.feeCache.Set(ctx, f.ID, req.Amount, feeAmount); err != nil {
                s.logger.Warn("Failed to cache fee calculation", zap.Error(err))
            }
        }

        totalFee = totalFee.Add(feeAmount)
        feeBreakdown = append(feeBreakdown, fee.FeeBreakdown{
            FeeID:       f.ID,
            FeeCode:     f.Code,
            FeeName:     f.Name,
            FeeType:     f.Type,
            Amount:      feeAmount,
            Description: fmt.Sprintf("%s fee (%s)", f.Structure, f.Type),
        })
    }

    // Apply corridor discount
    if corridor != nil {
        corridorFee := corridor.CalculateFee(req.Amount)
        totalFee = totalFee.Add(corridorFee)
    }

    // Create calculation record
    calc := &calculation.Calculation{
        ID:            uuid.New().String(),
        TransactionID: req.TransactionID,
        Amount:        req.Amount,
        TotalFee:      totalFee,
        Currency:      req.Currency,
        FeeBreakdown:  feeBreakdown,
        CorridorID:    func() string { if corridor != nil { return corridor.ID } else { return "" } }(),
        Status:        calculation.StatusCompleted,
        CompletedAt:   timePtr(time.Now()),
        Metadata:      req.Metadata,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    if err := s.calculationRepo.Create(ctx, calc); err != nil {
        s.logger.Warn("Failed to save calculation", zap.Error(err))
        // Non-critical error, continue
    }

    metrics.FeeCalculations.Inc()

    result := &fee.CalculationResult{
        TotalFee:        totalFee,
        Currency:        req.Currency,
        FeeBreakdown:    feeBreakdown,
        CorridorApplied: corridor,
        Timestamp:       time.Now(),
        Metadata:        req.Metadata,
    }

    return result, nil
}

// findMatchingCorridor finds a matching fee corridor
func (s *FeeServiceImpl) findMatchingCorridor(ctx context.Context, req *fee.CalculationRequest) (*corridor.FeeCorridor, error) {
    corridors, err := s.corridorRepo.List(ctx, &corridor.CorridorFilter{
        SourceCountry:  req.SourceCountry,
        TargetCountry:  req.TargetCountry,
        SourceCurrency: req.SourceCurrency,
        TargetCurrency: req.TargetCurrency,
        IsActive:       boolPtr(true),
        Limit:          10,
    })
    if err != nil {
        return nil, err
    }

    if len(corridors) == 0 {
        return nil, nil
    }

    // Return highest priority corridor
    sort.Slice(corridors, func(i, j int) bool {
        return corridors[i].Priority < corridors[j].Priority
    })

    return corridors[0], nil
}

// findApplicableFees finds fees applicable to the request
func (s *FeeServiceImpl) findApplicableFees(ctx context.Context, req *fee.CalculationRequest, corridor *corridor.FeeCorridor) ([]*fee.Fee, error) {
    corridorID := ""
    if corridor != nil {
        corridorID = corridor.ID
    }

    filter := &fee.FeeFilter{
        SourceCountry:  req.SourceCountry,
        TargetCountry:  req.TargetCountry,
        SourceCurrency: req.SourceCurrency,
        TargetCurrency: req.TargetCurrency,
        IsActive:       boolPtr(true),
        Limit:          100,
    }

    if corridorID != "" {
        filter.CorridorID = corridorID
    }

    return s.feeRepo.ListByFilter(ctx, filter)
}

// getFeesByCodes gets fees by their codes
func (s *FeeServiceImpl) getFeesByCodes(ctx context.Context, codes []string) ([]*fee.Fee, error) {
    var result []*fee.Fee

    for _, code := range codes {
        f, err := s.feeRepo.GetByCode(ctx, code)
        if err != nil {
            s.logger.Warn("Failed to get fee by code", zap.String("code", code), zap.Error(err))
            continue
        }
        if f != nil && f.IsActive {
            result = append(result, f)
        }
    }

    return result, nil
}

// validateCalculationRequest validates a calculation request
func (s *FeeServiceImpl) validateCalculationRequest(req *fee.CalculationRequest) error {
    if req.TransactionID == "" {
        return fee.ErrTransactionIDRequired
    }
    if req.Amount.IsNegative() || req.Amount.IsZero() {
        return fee.ErrInvalidAmount
    }
    if req.Currency == "" {
        return fee.ErrCurrencyRequired
    }
    return nil
}

// Helper functions
func timePtr(t time.Time) *time.Time {
    return &t
}

func boolPtr(b bool) *bool {
    return &b
}