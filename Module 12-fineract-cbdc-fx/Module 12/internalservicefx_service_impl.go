package service

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/apache/fineract-cbdc-fx/internal/domain/conversion"
    "github.com/apache/fineract-cbdc-fx/internal/domain/quote"
    "github.com/apache/fineract-cbdc-fx/internal/domain/rate"
    "github.com/apache/fineract-cbdc-fx/internal/infrastructure/cache"
    "github.com/apache/fineract-cbdc-fx/internal/infrastructure/client"
    "github.com/apache/fineract-cbdc-fx/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-fx/pkg/metrics"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type FXServiceImpl struct {
    rateRepo       rate.Repository
    quoteRepo      quote.Repository
    conversionRepo conversion.Repository
    rateProvider   client.RateProvider
    rateCache      cache.RateCache
    logger         *zap.Logger
    config         *config.FXConfig
    mu             sync.RWMutex
}

func NewFXService(
    rateRepo rate.Repository,
    quoteRepo quote.Repository,
    conversionRepo conversion.Repository,
    rateProvider client.RateProvider,
    rateCache cache.RateCache,
    logger *zap.Logger,
    config *config.FXConfig,
) fx.Service {
    return &FXServiceImpl{
        rateRepo:       rateRepo,
        quoteRepo:      quoteRepo,
        conversionRepo: conversionRepo,
        rateProvider:   rateProvider,
        rateCache:      rateCache,
        logger:         logger,
        config:         config,
    }
}

// GetRate gets the current exchange rate for a currency pair
func (s *FXServiceImpl) GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error) {
    startTime := time.Now()
    defer func() {
        metrics.FXLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate currency codes
    if err := s.validateCurrencyCodes(baseCurrency, quoteCurrency); err != nil {
        metrics.FXRateErrors.Inc()
        return nil, err
    }

    // Check cache first
    cachedRate, err := s.rateCache.Get(ctx, baseCurrency, quoteCurrency)
    if err == nil && cachedRate != nil && cachedRate.IsActive() {
        metrics.FXCacheHits.Inc()
        return cachedRate, nil
    }

    // Fetch from provider
    rate, err := s.rateProvider.GetRate(ctx, baseCurrency, quoteCurrency)
    if err != nil {
        metrics.FXRateErrors.Inc()
        return nil, fmt.Errorf("failed to fetch rate from provider: %w", err)
    }

    // Save to cache
    if err := s.rateCache.Set(ctx, rate); err != nil {
        s.logger.Warn("Failed to cache rate", zap.Error(err))
    }

    // Save to database
    if err := s.rateRepo.Create(ctx, rate); err != nil {
        s.logger.Warn("Failed to save rate to database", zap.Error(err))
    }

    metrics.FXCacheMisses.Inc()
    return rate, nil
}

// RefreshRates refreshes all exchange rates from providers
func (s *FXServiceImpl) RefreshRates(ctx context.Context) error {
    s.logger.Info("Refreshing exchange rates")

    // Get all supported currency pairs
    pairs, err := s.rateProvider.GetSupportedPairs(ctx)
    if err != nil {
        return fmt.Errorf("failed to get supported pairs: %w", err)
    }

    var wg sync.WaitGroup
    errCh := make(chan error, len(pairs))

    for _, pair := range pairs {
        wg.Add(1)
        go func(base, quote string) {
            defer wg.Done()
            rate, err := s.rateProvider.GetRate(ctx, base, quote)
            if err != nil {
                errCh <- fmt.Errorf("failed to get rate for %s/%s: %w", base, quote, err)
                return
            }
            if err := s.rateCache.Set(ctx, rate); err != nil {
                s.logger.Warn("Failed to cache rate", zap.Error(err))
            }
            if err := s.rateRepo.Create(ctx, rate); err != nil {
                s.logger.Warn("Failed to save rate to database", zap.Error(err))
            }
        }(pair.Base, pair.Quote)
    }

    wg.Wait()
    close(errCh)

    var errors []error
    for err := range errCh {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        s.logger.Warn("Some rates failed to refresh", zap.Int("failures", len(errors)))
        // Return first error, but log all
        return errors[0]
    }

    return nil
}

// GetQuote gets a locked FX quote for a transaction
func (s *FXServiceImpl) GetQuote(ctx context.Context, req *fx.QuoteRequest) (*quote.FXQuote, error) {
    startTime := time.Now()
    defer func() {
        metrics.FXQuoteLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Get current rate
    rate, err := s.GetRate(ctx, req.BaseCurrency, req.QuoteCurrency)
    if err != nil {
        metrics.FXQuoteErrors.Inc()
        return nil, err
    }

    // Calculate quote amounts
    baseAmount := req.Amount
    quoteAmount := baseAmount.Mul(rate.AskRate)

    // Calculate spread
    spread := rate.AskRate.Sub(rate.BidRate)

    // Calculate markup
    markupPercent := s.config.DefaultMarkupPercent
    markupAmount := quoteAmount.Mul(markupPercent.Div(decimal.NewFromInt(100)))

    // Apply markup
    finalRate := rate.AskRate.Add(rate.AskRate.Mul(markupPercent.Div(decimal.NewFromInt(100))))

    // Create quote
    quote := &quote.FXQuote{
        ID:              uuid.New().String(),
        TransactionID:   req.TransactionID,
        BaseCurrency:    req.BaseCurrency,
        QuoteCurrency:   req.QuoteCurrency,
        BaseAmount:      baseAmount,
        QuoteAmount:     quoteAmount,
        Rate:            rate.AskRate,
        BidRate:         rate.BidRate,
        AskRate:         rate.AskRate,
        Spread:          spread,
        MarkupPercent:   markupPercent,
        MarkupAmount:    markupAmount,
        SlippagePercent: decimal.Zero,
        SlippageAmount:  decimal.Zero,
        FinalRate:       finalRate,
        Status:          quote.QuoteStatusActive,
        LockDuration:    req.LockDuration,
        ExpiresAt:       time.Now().Add(req.LockDuration),
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Validate quote
    if err := quote.Validate(); err != nil {
        return nil, err
    }

    // Save quote
    if err := s.quoteRepo.Create(ctx, quote); err != nil {
        metrics.FXQuoteErrors.Inc()
        return nil, fmt.Errorf("failed to save quote: %w", err)
    }

    metrics.FXQuotesCreated.Inc()
    return quote, nil
}

// LockQuote locks a quote for use in a transaction
func (s *FXServiceImpl) LockQuote(ctx context.Context, quoteID string) (*quote.FXQuote, error) {
    q, err := s.quoteRepo.GetByID(ctx, quoteID)
    if err != nil {
        return nil, err
    }

    if q == nil {
        return nil, quote.ErrQuoteNotFound
    }

    if !q.IsValid() {
        return nil, quote.ErrQuoteInvalid
    }

    if q.Status == quote.QuoteStatusLocked {
        return q, nil // Already locked
    }

    q.Status = quote.QuoteStatusLocked
    now := time.Now()
    q.LockedAt = &now
    q.UpdatedAt = now

    if err := s.quoteRepo.Update(ctx, q); err != nil {
        return nil, err
    }

    return q, nil
}

// ReleaseQuote releases a locked quote
func (s *FXServiceImpl) ReleaseQuote(ctx context.Context, quoteID string) error {
    q, err := s.quoteRepo.GetByID(ctx, quoteID)
    if err != nil {
        return err
    }

    if q == nil {
        return quote.ErrQuoteNotFound
    }

    if q.Status != quote.QuoteStatusLocked {
        return quote.ErrQuoteNotLocked
    }

    q.Status = quote.QuoteStatusCancelled
    q.UpdatedAt = time.Now()

    return s.quoteRepo.Update(ctx, q)
}

// ConvertCurrency performs a currency conversion
func (s *FXServiceImpl) ConvertCurrency(ctx context.Context, req *fx.ConversionRequest) (*conversion.Conversion, error) {
    startTime := time.Now()
    defer func() {
        metrics.FXConversionLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate request
    if err := s.validateConversionRequest(req); err != nil {
        return nil, err
    }

    // Get rate if not provided
    var rate decimal.Decimal
    if req.Rate != nil {
        rate = *req.Rate
    } else {
        exchangeRate, err := s.GetRate(ctx, req.FromCurrency, req.ToCurrency)
        if err != nil {
            return nil, err
        }
        rate = exchangeRate.AskRate
    }

    // Calculate conversion
    toAmount := req.Amount.Mul(rate)

    // Calculate fee
    feeAmount := toAmount.Mul(s.config.DefaultFeePercent.Div(decimal.NewFromInt(100)))
    if s.config.MinFee.GreaterThan(feeAmount) {
        feeAmount = s.config.MinFee
    }

    // Create conversion record
    conversion := &conversion.Conversion{
        ID:              uuid.New().String(),
        TransactionID:   req.TransactionID,
        FromCurrency:    req.FromCurrency,
        ToCurrency:      req.ToCurrency,
        FromAmount:      req.Amount,
        ToAmount:        toAmount,
        RateUsed:        rate,
        FeeAmount:       feeAmount,
        FeeCurrency:     req.ToCurrency,
        MarkupApplied:   s.config.DefaultMarkupPercent,
        SlippageApplied: decimal.Zero,
        Status:          conversion.ConversionStatusCompleted,
        CompletedAt:     timePtr(time.Now()),
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
        Metadata:        req.Metadata,
    }

    // Save conversion
    if err := s.conversionRepo.Create(ctx, conversion); err != nil {
        metrics.FXConversionErrors.Inc()
        return nil, fmt.Errorf("failed to save conversion: %w", err)
    }

    metrics.FXConversionsCompleted.Inc()
    return conversion, nil
}

// ConvertWithQuote converts currency using a locked quote
func (s *FXServiceImpl) ConvertWithQuote(ctx context.Context, quoteID string, amount decimal.Decimal) (*conversion.Conversion, error) {
    // Get quote
    q, err := s.quoteRepo.GetByID(ctx, quoteID)
    if err != nil {
        return nil, err
    }

    if q == nil {
        return nil, quote.ErrQuoteNotFound
    }

    if q.Status != quote.QuoteStatusLocked {
        return nil, quote.ErrQuoteNotLocked
    }

    if q.IsExpired() {
        return nil, quote.ErrQuoteExpired
    }

    // Calculate conversion
    rate := q.FinalRate
    toAmount := amount.Mul(rate)

    // Create conversion record
    conversion := &conversion.Conversion{
        ID:              uuid.New().String(),
        TransactionID:   q.TransactionID,
        QuoteID:         q.ID,
        FromCurrency:    q.BaseCurrency,
        ToCurrency:      q.QuoteCurrency,
        FromAmount:      amount,
        ToAmount:        toAmount,
        RateUsed:        rate,
        FeeAmount:       q.MarkupAmount,
        FeeCurrency:     q.QuoteCurrency,
        MarkupApplied:   q.MarkupPercent,
        SlippageApplied: q.SlippagePercent,
        Status:          conversion.ConversionStatusCompleted,
        CompletedAt:     timePtr(time.Now()),
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Save conversion
    if err := s.conversionRepo.Create(ctx, conversion); err != nil {
        return nil, fmt.Errorf("failed to save conversion: %w", err)
    }

    // Mark quote as used
    q.Status = quote.QuoteStatusUsed
    q.UsedAt = timePtr(time.Now())
    q.UpdatedAt = time.Now()
    if err := s.quoteRepo.Update(ctx, q); err != nil {
        s.logger.Warn("Failed to update quote status", zap.Error(err))
    }

    return conversion, nil
}

// Helper functions
func (s *FXServiceImpl) validateCurrencyCodes(base, quote string) error {
    if base == "" || len(base) != 3 {
        return rate.ErrInvalidCurrencyCode
    }
    if quote == "" || len(quote) != 3 {
        return rate.ErrInvalidCurrencyCode
    }
    if base == quote {
        return rate.ErrSameCurrency
    }
    return nil
}

func (s *FXServiceImpl) validateConversionRequest(req *fx.ConversionRequest) error {
    if req.FromCurrency == "" || len(req.FromCurrency) != 3 {
        return conversion.ErrInvalidCurrency
    }
    if req.ToCurrency == "" || len(req.ToCurrency) != 3 {
        return conversion.ErrInvalidCurrency
    }
    if req.FromCurrency == req.ToCurrency {
        return conversion.ErrSameCurrency
    }
    if req.Amount.IsNegative() || req.Amount.IsZero() {
        return conversion.ErrInvalidAmount
    }
    return nil
}

func timePtr(t time.Time) *time.Time {
    return &t
}

// Other methods...
// GetHistoricalRates, GetRateProviders, GetQuoteStatus, GetConversionStatus, 
// RollbackConversion, HealthCheck, GetMetrics