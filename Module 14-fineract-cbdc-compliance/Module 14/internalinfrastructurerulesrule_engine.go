package rules

import (
    "context"
    "fmt"
    "sync"

    "github.com/apache/fineract-cbdc-compliance/internal/domain/screening"
    "github.com/apache/fineract-cbdc-compliance/internal/infrastructure/config"

    "go.uber.org/zap"
)

// Rule represents a compliance rule
type Rule struct {
    ID          string                 `yaml:"id"`
    Name        string                 `yaml:"name"`
    Description string                 `yaml:"description"`
    Type        string                 `yaml:"type"`
    Condition   RuleCondition          `yaml:"condition"`
    Action      RuleAction             `yaml:"action"`
    Severity    string                 `yaml:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
    Enabled     bool                   `yaml:"enabled"`
}

// RuleCondition represents a rule condition
type RuleCondition struct {
    Field    string                 `yaml:"field"`
    Operator string                 `yaml:"operator"` // equals, contains, gt, lt, gte, lte
    Value    interface{}            `yaml:"value"`
    And      []RuleCondition        `yaml:"and,omitempty"`
    Or       []RuleCondition        `yaml:"or,omitempty"`
}

// RuleAction represents a rule action
type RuleAction struct {
    Type   string                 `yaml:"type"` // block, escalate, approve, review
    Result string                 `yaml:"result"`
    Params map[string]interface{} `yaml:"params"`
}

// RuleResult represents the result of rule evaluation
type RuleResult struct {
    AppliedRules          []string `json:"appliedRules"`
    ViolatedRules         []string `json:"violatedRules"`
    Score                 int      `json:"score"`
    HasCriticalViolations bool     `json:"hasCriticalViolations"`
    Recommendations       []string `json:"recommendations"`
}

// RuleEngine evaluates compliance rules
type RuleEngine struct {
    rules  []Rule
    logger *zap.Logger
    mu     sync.RWMutex
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(cfg *config.RulesConfig, logger *zap.Logger) *RuleEngine {
    engine := &RuleEngine{
        logger: logger,
    }

    // Load rules from configuration
    if err := engine.loadRules(cfg); err != nil {
        logger.Error("Failed to load rules", zap.Error(err))
    }

    return engine
}

// loadRules loads rules from configuration
func (e *RuleEngine) loadRules(cfg *config.RulesConfig) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // In production, this would load from YAML file or database
    e.rules = []Rule{
        {
            ID:          "AML_001",
            Name:        "Large Transaction",
            Description: "Flag transactions above threshold",
            Type:        "AML",
            Condition: RuleCondition{
                Field:    "amount",
                Operator: "gt",
                Value:    10000.0,
            },
            Action: RuleAction{
                Type:   "review",
                Result: "HIGH_RISK",
                Params: map[string]interface{}{
                    "reason": "Large transaction requires review",
                },
            },
            Severity: "HIGH",
            Enabled:  true,
        },
        {
            ID:          "SAN_001",
            Name:        "High Risk Country",
            Description: "Flag transactions to/from high risk countries",
            Type:        "SANCTIONS",
            Condition: RuleCondition{
                Field:    "targetCountry",
                Operator: "in",
                Value:    []string{"XX", "YY", "ZZ"},
            },
            Action: RuleAction{
                Type:   "block",
                Result: "BLOCKED",
                Params: map[string]interface{}{
                    "reason": "Transaction to high risk country",
                },
            },
            Severity: "CRITICAL",
            Enabled:  true,
        },
        {
            ID:          "COMP_001",
            Name:        "Customer Blacklist",
            Description: "Check customer against blacklist",
            Type:        "COMPLIANCE",
            Condition: RuleCondition{
                Field:    "customerId",
                Operator: "in_blacklist",
                Value:    nil,
            },
            Action: RuleAction{
                Type:   "block",
                Result: "BLOCKED",
                Params: map[string]interface{}{
                    "reason": "Customer is blacklisted",
                },
            },
            Severity: "CRITICAL",
            Enabled:  true,
        },
    }

    return nil
}

// ApplyRules applies rules to a screening
func (e *RuleEngine) ApplyRules(ctx context.Context, screening *screening.Screening) (*RuleResult, error) {
    e.mu.RLock()
    defer e.mu.RUnlock()

    result := &RuleResult{
        AppliedRules:    []string{},
        ViolatedRules:   []string{},
        Recommendations: []string{},
    }

    for _, rule := range e.rules {
        if !rule.Enabled {
            continue
        }

        // Evaluate condition
        matched, err := e.evaluateCondition(screening, &rule.Condition)
        if err != nil {
            e.logger.Warn("Failed to evaluate rule condition",
                zap.String("rule", rule.Name),
                zap.Error(err),
            )
            continue
        }

        if matched {
            result.AppliedRules = append(result.AppliedRules, rule.ID)

            // Check if rule is violated
            if rule.Action.Type == "block" || rule.Action.Type == "review" {
                result.ViolatedRules = append(result.ViolatedRules, rule.ID)
                result.Score += e.getScoreForSeverity(rule.Severity)

                if rule.Severity == "CRITICAL" {
                    result.HasCriticalViolations = true
                }

                result.Recommendations = append(result.Recommendations,
                    fmt.Sprintf("Rule %s: %s", rule.ID, rule.Description),
                )
            }
        }
    }

    return result, nil
}

// evaluateCondition evaluates a rule condition
func (e *RuleEngine) evaluateCondition(screening *screening.Screening, condition *RuleCondition) (bool, error) {
    value, err := e.getFieldValue(screening, condition.Field)
    if err != nil {
        return false, err
    }

    switch condition.Operator {
    case "equals":
        return value == condition.Value, nil
    case "contains":
        str, ok := value.(string)
        if !ok {
            return false, nil
        }
        val, ok := condition.Value.(string)
        if !ok {
            return false, nil
        }
        return strings.Contains(str, val), nil
    case "gt":
        return e.compareGreaterThan(value, condition.Value), nil
    case "gte":
        return e.compareGreaterThanOrEqual(value, condition.Value), nil
    case "lt":
        return e.compareLessThan(value, condition.Value), nil
    case "lte":
        return e.compareLessThanOrEqual(value, condition.Value), nil
    case "in":
        return e.compareIn(value, condition.Value), nil
    case "in_blacklist":
        return e.checkBlacklist(screening.CustomerID), nil
    default:
        return false, fmt.Errorf("unknown operator: %s", condition.Operator)
    }
}

// getFieldValue gets a field value from screening
func (e *RuleEngine) getFieldValue(screening *screening.Screening, field string) (interface{}, error) {
    switch field {
    case "amount":
        return screening.Amount.InexactFloat64(), nil
    case "customerId":
        return screening.CustomerID, nil
    case "customerCountry":
        return screening.CustomerCountry, nil
    case "sourceCountry":
        return screening.SourceCountry, nil
    case "targetCountry":
        return screening.TargetCountry, nil
    case "currency":
        return screening.Currency, nil
    default:
        return nil, fmt.Errorf("unknown field: %s", field)
    }
}

// compareGreaterThan compares two values
func (e *RuleEngine) compareGreaterThan(a, b interface{}) bool {
    aFloat, okA := a.(float64)
    bFloat, okB := b.(float64)
    if okA && okB {
        return aFloat > bFloat
    }
    return false
}

// compareGreaterThanOrEqual compares two values
func (e *RuleEngine) compareGreaterThanOrEqual(a, b interface{}) bool {
    aFloat, okA := a.(float64)
    bFloat, okB := b.(float64)
    if okA && okB {
        return aFloat >= bFloat
    }
    return false
}

// compareLessThan compares two values
func (e *RuleEngine) compareLessThan(a, b interface{}) bool {
    aFloat, okA := a.(float64)
    bFloat, okB := b.(float64)
    if okA && okB {
        return aFloat < bFloat
    }
    return false
}

// compareLessThanOrEqual compares two values
func (e *RuleEngine) compareLessThanOrEqual(a, b interface{}) bool {
    aFloat, okA := a.(float64)
    bFloat, okB := b.(float64)
    if okA && okB {
        return aFloat <= bFloat
    }
    return false
}

// compareIn checks if a value is in a list
func (e *RuleEngine) compareIn(value interface{}, list interface{}) bool {
    listSlice, ok := list.([]interface{})
    if !ok {
        return false
    }
    for _, item := range listSlice {
        if value == item {
            return true
        }
    }
    return false
}

// checkBlacklist checks if a customer is blacklisted
func (e *RuleEngine) checkBlacklist(customerID string) bool {
    // In production, this would check against a blacklist database
    // For now, return false
    return false
}

// getScoreForSeverity returns the score for a severity level
func (e *RuleEngine) getScoreForSeverity(severity string) int {
    switch severity {
    case "CRITICAL":
        return 40
    case "HIGH":
        return 20
    case "MEDIUM":
        return 10
    case "LOW":
        return 5
    default:
        return 0
    }
}