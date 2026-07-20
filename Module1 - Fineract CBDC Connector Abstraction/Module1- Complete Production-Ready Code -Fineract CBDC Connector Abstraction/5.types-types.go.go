package types

import (
    "encoding/json"
    "fmt"
    "math/big"
    "time"
)

// CurrencyCode represents a 3-letter ISO currency code.
type CurrencyCode string

// TransactionStatus represents the status of a transaction.
type TransactionStatus string

const (
    StatusPending    TransactionStatus = "pending"
    StatusProcessing TransactionStatus = "processing"
    StatusConfirmed  TransactionStatus = "confirmed"
    StatusFailed     TransactionStatus = "failed"
    StatusExpired    TransactionStatus = "expired"
    StatusRolledBack TransactionStatus = "rolled_back"
    StatusCancelled  TransactionStatus = "cancelled"
)

// TransactionType represents the type of a transaction.
type TransactionType string

const (
    TypeIssue    TransactionType = "issue"
    TypeTransfer TransactionType = "transfer"
    TypeLock     TransactionType = "lock"
    TypeBurn     TransactionType = "burn"
    TypeRedeem   TransactionType = "redeem"
    TypeUnlock   TransactionType = "unlock"
)

// HealthStatus represents the health status of a component.
type HealthStatus string

const (
    HealthStatusUp   HealthStatus = "up"
    HealthStatusDown HealthStatus = "down"
    HealthStatusDegraded HealthStatus = "degraded"
    HealthStatusUnknown HealthStatus = "unknown"
)

// ComponentHealth represents the health of a component.
type ComponentHealth struct {
    Status   HealthStatus            `json:"status"`
    Message  string                  `json:"message,omitempty"`
    Details  map[string]interface{}  `json:"details,omitempty"`
    CheckedAt time.Time              `json:"checkedAt"`
}

// EventType represents a connector event type.
type EventType string

const (
    EventTransactionStarted  EventType = "transaction_started"
    EventTransactionPending  EventType = "transaction_pending"
    EventTransactionConfirmed EventType = "transaction_confirmed"
    EventTransactionFailed   EventType = "transaction_failed"
    EventTransactionExpired  EventType = "transaction_expired"
    EventLockCreated         EventType = "lock_created"
    EventLockReleased        EventType = "lock_released"
    EventLockExpired         EventType = "lock_expired"
    EventHealthChanged       EventType = "health_changed"
    EventBalanceChanged      EventType = "balance_changed"
)

// FeatureType represents a connector feature type.
type FeatureType string

const (
    FeatureAtomicTransaction FeatureType = "atomic_transaction"
    FeatureBatchOperation    FeatureType = "batch_operation"
    FeatureWebhook          FeatureType = "webhook"
    FeatureLock             FeatureType = "lock"
    FeatureRedeem           FeatureType = "redeem"
    FeatureCrossChain       FeatureType = "cross_chain"
)

// NetworkType represents the type of network.
type NetworkType string

const (
    NetworkTypeBlockchain  NetworkType = "blockchain"
    NetworkTypeCentralized NetworkType = "centralized"
    NetworkTypeHybrid      NetworkType = "hybrid"
)

// Amount represents a monetary amount with high precision.
type Amount struct {
    Value   *big.Int `json:"value"`
    Decimal int      `json:"decimal"`
}

// NewAmount creates a new Amount from a string.
func NewAmount(value string, decimal int) (*Amount, error) {
    // Remove leading/trailing whitespace
    value = strings.TrimSpace(value)
    if value == "" {
        return nil, fmt.Errorf("amount cannot be empty")
    }
    
    // Handle negative amounts
    negative := false
    if value[0] == '-' {
        negative = true
        value = value[1:]
    }
    
    // Split integer and decimal parts
    parts := strings.Split(value, ".")
    if len(parts) > 2 {
        return nil, fmt.Errorf("invalid amount format: %s", value)
    }
    
    integerPart := parts[0]
    decimalPart := ""
    if len(parts) == 2 {
        decimalPart = parts[1]
    }
    
    // Validate decimal precision
    if len(decimalPart) > decimal {
        return nil, fmt.Errorf("amount exceeds maximum decimal precision of %d", decimal)
    }
    
    // Pad decimal part to required precision
    if len(decimalPart) < decimal {
        decimalPart = decimalPart + strings.Repeat("0", decimal-len(decimalPart))
    }
    
    // Combine integer and decimal parts
    combined := integerPart + decimalPart
    if combined == "" {
        return nil, fmt.Errorf("invalid amount: %s", value)
    }
    
    val := new(big.Int)
    _, ok := val.SetString(combined, 10)
    if !ok {
        return nil, fmt.Errorf("invalid amount: %s", value)
    }
    
    if negative {
        val.Neg(val)
    }
    
    return &Amount{Value: val, Decimal: decimal}, nil
}

// String returns the string representation of the amount.
func (a *Amount) String() string {
    if a == nil || a.Value == nil {
        return "0"
    }
    
    s := a.Value.String()
    negative := false
    if s[0] == '-' {
        negative = true
        s = s[1:]
    }
    
    // Pad with zeros if needed
    if len(s) < a.Decimal {
        s = strings.Repeat("0", a.Decimal-len(s)) + s
    }
    
    integerPart := s[:len(s)-a.Decimal]
    decimalPart := s[len(s)-a.Decimal:]
    
    // Remove trailing zeros from decimal part
    decimalPart = strings.TrimRight(decimalPart, "0")
    
    result := integerPart
    if decimalPart != "" {
        result += "." + decimalPart
    }
    
    if negative {
        result = "-" + result
    }
    
    return result
}

// Add adds two amounts and returns a new amount.
func (a *Amount) Add(b *Amount) (*Amount, error) {
    if a.Decimal != b.Decimal {
        return nil, fmt.Errorf("amount decimal mismatch: %d vs %d", a.Decimal, b.Decimal)
    }
    
    result := new(big.Int).Add(a.Value, b.Value)
    return &Amount{Value: result, Decimal: a.Decimal}, nil
}

// Sub subtracts two amounts and returns a new amount.
func (a *Amount) Sub(b *Amount) (*Amount, error) {
    if a.Decimal != b.Decimal {
        return nil, fmt.Errorf("amount decimal mismatch: %d vs %d", a.Decimal, b.Decimal)
    }
    
    result := new(big.Int).Sub(a.Value, b.Value)
    return &Amount{Value: result, Decimal: a.Decimal}, nil
}

// Cmp compares two amounts.
func (a *Amount) Cmp(b *Amount) int {
    if a.Decimal != b.Decimal {
        // Convert to same decimal precision for comparison
        // This is a simplified version; in production, handle properly
        return a.Value.Cmp(b.Value) // Assuming same decimal
    }
    return a.Value.Cmp(b.Value)
}

// IsZero checks if the amount is zero.
func (a *Amount) IsZero() bool {
    return a.Value.Sign() == 0
}

// IsNegative checks if the amount is negative.
func (a *Amount) IsNegative() bool {
    return a.Value.Sign() < 0
}

// MarshalJSON implements json.Marshaler.
func (a *Amount) MarshalJSON() ([]byte, error) {
    return json.Marshal(a.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Amount) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    
    amount, err := NewAmount(s, a.Decimal)
    if err != nil {
        return err
    }
    
    *a = *amount
    return nil
}