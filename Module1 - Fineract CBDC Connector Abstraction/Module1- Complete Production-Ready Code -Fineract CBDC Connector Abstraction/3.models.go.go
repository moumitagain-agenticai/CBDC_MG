package connector

import (
    "time"
    
    "github.com/fineract/cbdc/connector-abstraction/types"
)

// ConnectorMetadata contains metadata about a connector implementation.
type ConnectorMetadata struct {
    // Name is the display name of the connector.
    Name string `json:"name"`
    
    // Version is the version of the connector.
    Version string `json:"version"`
    
    // Description is a brief description of the connector.
    Description string `json:"description"`
    
    // SupportedCurrencies are the currencies supported by this connector.
    SupportedCurrencies []types.CurrencyCode `json:"supportedCurrencies"`
    
    // SupportedFeatures are the features supported by this connector.
    SupportedFeatures []types.FeatureType `json:"supportedFeatures"`
    
    // NetworkType is the type of network (e.g., "blockchain", "centralized").
    NetworkType types.NetworkType `json:"networkType"`
    
    // Provider is the network provider name.
    Provider string `json:"provider"`
}

// ConnectorCapabilities defines the capabilities of a connector.
type ConnectorCapabilities struct {
    // SupportsIssue indicates if issuing is supported.
    SupportsIssue bool `json:"supportsIssue"`
    
    // SupportsTransfer indicates if transferring is supported.
    SupportsTransfer bool `json:"supportsTransfer"`
    
    // SupportsLock indicates if locking is supported.
    SupportsLock bool `json:"supportsLock"`
    
    // SupportsBurn indicates if burning is supported.
    SupportsBurn bool `json:"supportsBurn"`
    
    // SupportsRedeem indicates if redeeming is supported.
    SupportsRedeem bool `json:"supportsRedeem"`
    
    // SupportsBatchOperations indicates if batch operations are supported.
    SupportsBatchOperations bool `json:"supportsBatchOperations"`
    
    // SupportsWebhooks indicates if webhook notifications are supported.
    SupportsWebhooks bool `json:"supportsWebhooks"`
    
    // SupportsAtomicTransactions indicates if atomic transactions are supported.
    SupportsAtomicTransactions bool `json:"supportsAtomicTransactions"`
    
    // MaxBatchSize is the maximum batch size for batch operations.
    MaxBatchSize int `json:"maxBatchSize,omitempty"`
    
    // MinConfirmation is the minimum confirmation required.
    MinConfirmation int `json:"minConfirmation,omitempty"`
    
    // MaxTransactionAmount is the maximum transaction amount.
    MaxTransactionAmount types.Amount `json:"maxTransactionAmount,omitempty"`
    
    // TransactionTimeout is the timeout for transactions.
    TransactionTimeout time.Duration `json:"transactionTimeout,omitempty"`
}

// Event represents a connector event.
type Event struct {
    // ID is the unique event ID.
    ID string `json:"id"`
    
    // Type is the event type.
    Type types.EventType `json:"type"`
    
    // TransactionID is the associated transaction ID (if any).
    TransactionID string `json:"transactionId,omitempty"`
    
    // Data contains the event data.
    Data map[string]interface{} `json:"data"`
    
    // Timestamp is when the event occurred.
    Timestamp time.Time `json:"timestamp"`
}

// EventHandler defines a handler for connector events.
type EventHandler interface {
    // HandleEvent handles a connector event.
    HandleEvent(ctx context.Context, event *Event) error
}

// EventSubscription represents a subscription to connector events.
type EventSubscription struct {
    // ID is the subscription ID.
    ID string `json:"id"`
    
    // EventTypes are the event types to subscribe to.
    EventTypes []types.EventType `json:"eventTypes"`
    
    // Handler is the event handler.
    Handler EventHandler `json:"-"`
    
    // CreatedAt is when the subscription was created.
    CreatedAt time.Time `json:"createdAt"`
}