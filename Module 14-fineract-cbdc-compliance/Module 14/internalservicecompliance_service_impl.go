package service

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/apache/fineract-cbdc-compliance/internal/domain/audit"
    "github.com/apache/fineract-cbdc-compliance/internal/domain/compliance"
    "github.com/apache/fineract-cbdc-compliance/internal/domain/sanctions"
    "github.com/apache/fineract-cbdc-compliance/internal/domain/screening"
    "github.com/apache/fineract-cbdc-compliance/internal/infrastructure/client"
    "github.com/apache/fineract-cbdc-compliance/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-compliance/internal/infrastructure/rules"
    "github.com/apache/fineract-cbdc-compliance/pkg/metrics"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type ComplianceServiceImpl struct {
    screeningRepo  screening.Repository
    sanctionsRepo  sanctions.Repository
    complianceRepo compliance.Repository
    auditRepo      audit.Repository
    ofacClient     *client.OFACClient
    unClient       *client.UNClient
    ruleEngine     *rules.RuleEngine
    logger         *zap.Logger
    config         *config.ComplianceConfig
}

func NewComplianceService(
    screeningRepo screening.Repository,
    sanctionsRepo sanctions.Repository,
    complianceRepo compliance.Repository,
    auditRepo audit.Repository,
    ofacClient *client.OFACClient,
    unClient *client.UNClient,
    ruleEngine *rules.RuleEngine,
    logger *zap.Logger,
    config *config.ComplianceConfig,
) compliance.Service {
    return &ComplianceServiceImpl{
        screeningRepo:  screeningRepo,
        sanctionsRepo:  sanctionsRepo,
        complianceRepo: complianceRepo,
        auditRepo:      auditRepo,
        ofacClient:     ofacClient,
        unClient:       unClient,
        ruleEngine:     ruleEngine,
        logger:         logger,
        config:         config,
    }
}

// ScreenTransaction performs a full compliance screening on a transaction
func (s *ComplianceServiceImpl) ScreenTransaction(ctx context.Context, req *compliance.ScreeningRequest) (*screening.Screening, error) {
    startTime := time.Now()
    defer func() {
        metrics.ScreeningLatency.Observe(time.Since(startTime).Seconds())
    }()

    // Validate request
    if err := s.validateScreeningRequest(req); err != nil {
        metrics.ScreeningErrors.Inc()
        return nil, err
    }

    // Create screening record
    screening := &screening.Screening{
        ID:              uuid.New().String(),
        TransactionID:   req.TransactionID,
        CustomerID:      req.CustomerID,
        CustomerName:    req.CustomerName,
        CustomerCountry: req.CustomerCountry,
        Amount:          req.Amount,
        Currency:        req.Currency,
        SourceCountry:   req.SourceCountry,
        TargetCountry:   req.TargetCountry,
        Status:          screening.StatusPending,
        Type:            req.Type,
        Score:           0,
        Details:         make(map[string]interface{}),
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
        ExpiresAt:       time.Now().Add(24 * time.Hour),
    }

    // Save screening
    if err := s.screeningRepo.Create(ctx, screening); err != nil {
        metrics.ScreeningErrors.Inc()
        return nil, fmt.Errorf("failed to create screening: %w", err)
    }

    // Execute screening
    go func() {
        s.executeScreening(context.Background(), screening)
    }()

    metrics.ScreeningsInitiated.Inc()
    return screening, nil
}

// executeScreening executes the compliance screening process
func (s *ComplianceServiceImpl) executeScreening(ctx context.Context, screening *screening.Screening) {
    s.logger.Info("Executing compliance screening",
        zap.String("screening_id", screening.ID),
        zap.String("transaction_id", screening.TransactionID),
    )

    // Step 1: Check sanctions
    sanctionsResult, err := s.checkSanctionsAgainstLists(ctx, screening)
    if err != nil {
        s.failScreening(ctx, screening, "Sanctions check failed: "+err.Error())
        return
    }

    // Step 2: Apply compliance rules
    ruleResult, err := s.ruleEngine.ApplyRules(ctx, screening)
    if err != nil {
        s.failScreening(ctx, screening, "Rule engine failed: "+err.Error())
        return
    }

    // Step 3: Calculate overall score
    totalScore := s.calculateOverallScore(sanctionsResult, ruleResult)

    // Step 4: Determine final status
    finalStatus, result := s.determineFinalStatus(totalScore, sanctionsResult, ruleResult)

    // Update screening
    screening.Status = finalStatus
    screening.Result = result
    screening.Score = totalScore
    screening.Details["sanctions"] = sanctionsResult
    screening.Details["rules"] = ruleResult
    screening.Details["score_breakdown"] = map[string]int{
        "sanctions": sanctionsResult.Score,
        "rules":     ruleResult.Score,
        "total":     totalScore,
    }
    screening.MatchedSanctions = sanctionsResult.MatchedItems

    now := time.Now()
    screening.CompletedAt = &now
    screening.UpdatedAt = now

    if err := s.screeningRepo.Update(ctx, screening); err != nil {
        s.logger.Error("Failed to update screening",
            zap.String("screening_id", screening.ID),
            zap.Error(err),
        )
    }

    // Record audit trail
    s.recordAudit(ctx, screening)

    // Record metrics
    metrics.ScreeningsCompleted.Inc()
    if screening.Status == screening.StatusBlocked {
        metrics.ScreeningsBlocked.Inc()
    }

    s.logger.Info("Screening completed",
        zap.String("screening_id", screening.ID),
        zap.String("status", string(screening.Status)),
        zap.Int("score", screening.Score),
    )
}

// checkSanctionsAgainstLists checks a customer against sanctions lists
func (s *ComplianceServiceImpl) checkSanctionsAgainstLists(ctx context.Context, screening *screening.Screening) (*SanctionsCheckResult, error) {
    // Check OFAC
    ofacResult, err := s.ofacClient.Check(ctx, screening.CustomerName, screening.CustomerCountry)
    if err != nil {
        s.logger.Warn("OFAC check failed", zap.Error(err))
    }

    // Check UN
    unResult, err := s.unClient.Check(ctx, screening.CustomerName, screening.CustomerCountry)
    if err != nil {
        s.logger.Warn("UN check failed", zap.Error(err))
    }

    // Check internal sanctions list
    internalResult, err := s.checkInternalSanctions(ctx, screening.CustomerName)
    if err != nil {
        s.logger.Warn("Internal sanctions check failed", zap.Error(err))
    }

    // Combine results
    matchedItems := []sanctions.SanctionsList{}
    score := 0

    if ofacResult != nil && ofacResult.IsMatch {
        matchedItems = append(matchedItems, ofacResult.MatchedItems...)
        score += 50
    }

    if unResult != nil && unResult.IsMatch {
        matchedItems = append(matchedItems, unResult.MatchedItems...)
        score += 50
    }

    if internalResult != nil && internalResult.IsMatch {
        matchedItems = append(matchedItems, internalResult.MatchedItems...)
        score += 30
    }

    return &SanctionsCheckResult{
        IsMatch:      len(matchedItems) > 0,
        MatchedItems: matchedItems,
        Score:        score,
        Sources:      s.getMatchSources(ofacResult, unResult, internalResult),
    }, nil
}

// checkInternalSanctions checks against internal sanctions list
func (s *ComplianceServiceImpl) checkInternalSanctions(ctx context.Context, name string) (*SanctionsCheckResult, error) {
    // Query internal sanctions list
    entries, err := s.sanctionsRepo.FindByName(ctx, name)
    if err != nil {
        return nil, err
    }

    if len(entries) == 0 {
        return &SanctionsCheckResult{IsMatch: false, Score: 0}, nil
    }

    return &SanctionsCheckResult{
        IsMatch:      true,
        MatchedItems: entries,
        Score:        30,
    }, nil
}

// calculateOverallScore calculates the overall compliance score
func (s *ComplianceServiceImpl) calculateOverallScore(sanctionsResult *SanctionsCheckResult, ruleResult *rules.RuleResult) int {
    score := 0

    // Sanctions score (0-100)
    if sanctionsResult != nil {
        score += sanctionsResult.Score
    }

    // Rules score (0-100)
    if ruleResult != nil {
        score += ruleResult.Score
    }

    // Cap at 100
    if score > 100 {
        score = 100
    }

    return score
}

// determineFinalStatus determines the final status based on score
func (s *ComplianceServiceImpl) determineFinalStatus(score int, sanctionsResult *SanctionsCheckResult, ruleResult *rules.RuleResult) (screening.ScreeningStatus, string) {
    // High risk - blocked
    if score >= 80 {
        return screening.StatusBlocked, "High risk score - transaction blocked"
    }

    // Medium risk - escalated
    if score >= 50 {
        return screening.StatusEscalated, "Medium risk score - requires manual review"
    }

    // Check for sanctions match
    if sanctionsResult != nil && sanctionsResult.IsMatch {
        return screening.StatusBlocked, "Sanctions match detected - transaction blocked"
    }

    // Check for critical rule violations
    if ruleResult != nil && ruleResult.HasCriticalViolations {
        return screening.StatusBlocked, "Critical compliance rule violation"
    }

    // Approved
    return screening.StatusCompleted, "Approved"
}

// failScreening marks a screening as failed
func (s *ComplianceServiceImpl) failScreening(ctx context.Context, screening *screening.Screening, errorMsg string) {
    screening.Status = screening.StatusFailed
    screening.ErrorMessage = errorMsg
    now := time.Now()
    screening.CompletedAt = &now
    screening.UpdatedAt = now

    if err := s.screeningRepo.Update(ctx, screening); err != nil {
        s.logger.Error("Failed to update screening",
            zap.String("screening_id", screening.ID),
            zap.Error(err),
        )
    }
}

// recordAudit records an audit trail
func (s *ComplianceServiceImpl) recordAudit(ctx context.Context, screening *screening.Screening) {
    auditRecord := &audit.AuditTrail{
        ID:          uuid.New().String(),
        ScreeningID: screening.ID,
        Action:      "SCREENING_COMPLETED",
        Status:      string(screening.Status),
        Details: map[string]interface{}{
            "score":      screening.Score,
            "result":     screening.Result,
            "matched":    screening.MatchedSanctions,
            "type":       string(screening.Type),
            "customer":   screening.CustomerID,
        },
        CreatedAt: time.Now(),
    }

    if err := s.auditRepo.Create(ctx, auditRecord); err != nil {
        s.logger.Warn("Failed to record audit trail",
            zap.Error(err),
            zap.String("screening_id", screening.ID),
        )
    }
}