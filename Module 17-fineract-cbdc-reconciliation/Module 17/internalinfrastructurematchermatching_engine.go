package matcher

import (
    "context"
    "fmt"
    "time"

    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/matching"
    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/reconciliation"
    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/statement"

    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// MatchingEngine handles the matching process
type MatchingEngine struct {
    ruleBasedMatcher *RuleBasedMatcher
    fuzzyMatcher     *FuzzyMatcher
    matchingRepo     matching.Repository
    logger           *zap.Logger
    config           *MatchingConfig
}

// MatchingConfig holds matching configuration
type MatchingConfig struct {
    ExactMatchThreshold   float64 `mapstructure:"exact_match_threshold"`
    FuzzyMatchThreshold   float64 `mapstructure:"fuzzy_match_threshold"`
    PartialMatchThreshold float64 `mapstructure:"partial_match_threshold"`
    MaxDateDifference     int     `mapstructure:"max_date_difference"`
    MaxAmountDifference   float64 `mapstructure:"max_amount_difference"`
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(
    ruleBasedMatcher *RuleBasedMatcher,
    fuzzyMatcher *FuzzyMatcher,
    matchingRepo matching.Repository,
    logger *zap.Logger,
    config *MatchingConfig,
) *MatchingEngine {
    return &MatchingEngine{
        ruleBasedMatcher: ruleBasedMatcher,
        fuzzyMatcher:     fuzzyMatcher,
        matchingRepo:     matchingRepo,
        logger:           logger,
        config:           config,
    }
}

// Match executes the matching process
func (e *MatchingEngine) Match(ctx context.Context, rec *reconciliation.Reconciliation, stmt *statement.BankStatement) (*reconciliation.MatchingResult, error) {
    e.logger.Info("Starting matching process",
        zap.String("reconciliation_id", rec.ID),
        zap.Int("statement_entries", len(stmt.Entries)),
    )

    result := &reconciliation.MatchingResult{
        ReconciliationID: rec.ID,
        TotalEntries:     len(stmt.Entries),
        Details:          []reconciliation.MatchDetail{},
        CompletedAt:      time.Now(),
    }

    // Create matching map
    matchedBankEntries := make(map[string]bool)

    // Process each statement entry
    for _, entry := range stmt.Entries {
        // Rule-based matching (exact)
        match, err := e.ruleBasedMatcher.Match(ctx, entry, rec)
        if err != nil {
            e.logger.Warn("Rule-based matching failed", zap.Error(err))
            continue
        }

        if match != nil && match.IsMatch {
            matchedBankEntries[entry.ID] = true
            result.MatchedEntries++
            result.Details = append(result.Details, reconciliation.MatchDetail{
                SystemEntryID:  entry.ID,
                SystemAmount:   entry.Amount,
                SystemDate:     entry.Date,
                BankEntryID:    match.BankEntryID,
                BankAmount:     entry.Amount,
                BankDate:       entry.Date,
                MatchStatus:    "MATCHED",
                ConfidenceScore: 1.0,
            })
            continue
        }

        // Try fuzzy matching
        fuzzyMatch, err := e.fuzzyMatcher.Match(ctx, entry, rec, stmt)
        if err != nil {
            e.logger.Warn("Fuzzy matching failed", zap.Error(err))
            continue
        }

        if fuzzyMatch != nil && fuzzyMatch.IsMatch && fuzzyMatch.ConfidenceScore >= e.config.FuzzyMatchThreshold {
            matchedBankEntries[entry.ID] = true
            result.MatchedEntries++
            result.Details = append(result.Details, reconciliation.MatchDetail{
                SystemEntryID:  entry.ID,
                SystemAmount:   entry.Amount,
                SystemDate:     entry.Date,
                BankEntryID:    fuzzyMatch.BankEntryID,
                BankAmount:     entry.Amount,
                BankDate:       entry.Date,
                MatchStatus:    "FUZZY_MATCHED",
                ConfidenceScore: fuzzyMatch.ConfidenceScore,
            })
            continue
        }

        // Unmatched entry
        result.UnmatchedEntries++
        result.Details = append(result.Details, reconciliation.MatchDetail{
            SystemEntryID:  entry.ID,
            SystemAmount:   entry.Amount,
            SystemDate:     entry.Date,
            BankEntryID:    "",
            BankAmount:     decimal.Zero,
            BankDate:       time.Time{},
            MatchStatus:    "UNMATCHED",
            ConfidenceScore: 0,
        })
    }

    // Calculate balances
    result.SystemBalance = rec.OpeningBalance.Add(stmt.TotalCredit).Sub(stmt.TotalDebit)
    result.BankBalance = stmt.ClosingBalance
    result.Difference = result.SystemBalance.Sub(result.BankBalance)
    result.IsBalanced = result.Difference.IsZero()

    return result, nil
}