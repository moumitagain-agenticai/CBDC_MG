package types

import (
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
	HealthStatusUp       HealthStatus = "up"
	HealthStatusDown     HealthStatus = "down"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusUnknown  HealthStatus = "unknown"
)

// ComponentHealth represents the health of a component.
type ComponentHealth struct {
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CheckedAt time.Time              `json:"checkedAt"`
}

// EventType represents a connector event type.
type EventType string

const (
	EventTransactionStarted   EventType = "transaction_started"
	EventTransactionPending   EventType = "transaction_pending"
	EventTransactionConfirmed EventType = "transaction_confirmed"
	EventTransactionFailed    EventType = "transaction_failed"
	EventTransactionExpired   EventType = "transaction_expired"
	EventLockCreated          EventType = "lock_created"
	EventLockReleased         EventType = "lock_released"
	EventLockExpired          EventType = "lock_expired"
	EventHealthChanged        EventType = "health_changed"
	EventBalanceChanged       EventType = "balance_changed"
)

// FeatureType represents a connector feature type.
type FeatureType string

const (
	FeatureAtomicTransaction FeatureType = "atomic_transaction"
	FeatureBatchOperation    FeatureType = "batch_operation"
	FeatureWebhook           FeatureType = "webhook"
	FeatureLock              FeatureType = "lock"
	FeatureRedeem            FeatureType = "redeem"
	FeatureCrossChain        FeatureType = "cross_chain"
)

// NetworkType represents the type of network.
type NetworkType string

const (
	NetworkTypeBlockchain  NetworkType = "blockchain"
	NetworkTypeCentralized NetworkType = "centralized"
	NetworkTypeHybrid      NetworkType = "hybrid"
)
