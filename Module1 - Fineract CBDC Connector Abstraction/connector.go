package connector

import (
	"context"
	"time"

	"github.com/fineract/cbdc/connector-abstraction/config"
	"github.com/fineract/cbdc/connector-abstraction/types"
)

// CBDCConnector defines the standard interface for all CBDC network connectors.
// All connectors (India, UAE, GiftCity, etc.) must implement this interface.
type CBDCConnector interface {
	// Basic Operations

	// Issue creates new CBDC tokens on the network.
	// Returns transaction ID and error if any.
	Issue(ctx context.Context, req *IssueRequest) (*IssueResponse, error)

	// Transfer moves CBDC tokens between wallets/accounts.
	// Returns transaction ID and error if any.
	Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)

	// Lock temporarily locks CBDC tokens for atomic settlement.
	// Returns lock ID and error if any.
	Lock(ctx context.Context, req *LockRequest) (*LockResponse, error)

	// Burn permanently destroys CBDC tokens.
	// Returns transaction ID and error if any.
	Burn(ctx context.Context, req *BurnRequest) (*BurnResponse, error)

	// Redeem converts CBDC back to fiat currency.
	// Returns transaction ID and error if any.
	Redeem(ctx context.Context, req *RedeemRequest) (*RedeemResponse, error)

	// Query Operations

	// GetBalance retrieves the current balance for a wallet/account.
	GetBalance(ctx context.Context, req *BalanceRequest) (*BalanceResponse, error)

	// GetTransaction retrieves transaction details by ID.
	GetTransaction(ctx context.Context, req *TransactionQueryRequest) (*TransactionResponse, error)

	// GetTransactionStatus retrieves the current status of a transaction.
	GetTransactionStatus(ctx context.Context, req *TransactionStatusRequest) (*TransactionStatusResponse, error)

	// Health & Monitoring

	// HealthCheck performs a health check on the connector and underlying network.
	HealthCheck(ctx context.Context) (*HealthResponse, error)

	// GetNetworkInfo retrieves current network information.
	GetNetworkInfo(ctx context.Context) (*NetworkInfoResponse, error)

	// Lifecycle Management

	// Initialize initializes the connector with the given configuration.
	Initialize(ctx context.Context, cfg *config.ConnectorConfig) error

	// Shutdown gracefully shuts down the connector.
	Shutdown(ctx context.Context) error

	// Status returns the current status of the connector.
	Status() ConnectorStatus
}

// ConnectorStatus represents the operational status of a connector.
type ConnectorStatus string

const (
	StatusUninitialized ConnectorStatus = "uninitialized"
	StatusInitializing  ConnectorStatus = "initializing"
	StatusReady         ConnectorStatus = "ready"
	StatusDegraded      ConnectorStatus = "degraded"
	StatusUnavailable   ConnectorStatus = "unavailable"
	StatusError         ConnectorStatus = "error"
)

// IssueRequest represents a request to issue new CBDC tokens.
type IssueRequest struct {
	// WalletID is the unique identifier of the wallet receiving the issued tokens.
	WalletID string `json:"walletId" validate:"required"`

	// Amount is the quantity of CBDC tokens to issue.
	Amount types.Amount `json:"amount" validate:"required,gt=0"`

	// Currency is the currency code (e.g., "INR", "AED").
	Currency types.CurrencyCode `json:"currency" validate:"required,len=3"`

	// ReferenceID is an external reference ID for reconciliation.
	ReferenceID string `json:"referenceId,omitempty"`

	// Metadata contains additional key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// TTL is the time-to-live for this request (optional).
	TTL *time.Duration `json:"ttl,omitempty"`
}

// IssueResponse represents the response to an issue request.
type IssueResponse struct {
	// TransactionID is the unique identifier of the transaction on the CBDC network.
	TransactionID string `json:"transactionId"`

	// BlockHash is the hash of the block containing the transaction.
	BlockHash string `json:"blockHash,omitempty"`

	// Timestamp is the time when the transaction was confirmed.
	Timestamp time.Time `json:"timestamp"`

	// Status is the final status of the transaction.
	Status types.TransactionStatus `json:"status"`

	// Fee is the transaction fee paid.
	Fee types.Amount `json:"fee,omitempty"`

	// ConfirmationCount is the number of confirmations.
	ConfirmationCount int `json:"confirmationCount,omitempty"`
}

// TransferRequest represents a request to transfer CBDC tokens.
type TransferRequest struct {
	// SourceWalletID is the wallet ID sending the tokens.
	SourceWalletID string `json:"sourceWalletId" validate:"required"`

	// DestinationWalletID is the wallet ID receiving the tokens.
	DestinationWalletID string `json:"destinationWalletId" validate:"required"`

	// Amount is the quantity of CBDC tokens to transfer.
	Amount types.Amount `json:"amount" validate:"required,gt=0"`

	// Currency is the currency code (e.g., "INR", "AED").
	Currency types.CurrencyCode `json:"currency" validate:"required,len=3"`

	// ReferenceID is an external reference ID for reconciliation.
	ReferenceID string `json:"referenceId,omitempty"`

	// PurposeCode identifies the purpose of the transfer.
	PurposeCode string `json:"purposeCode,omitempty"`

	// Metadata contains additional key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// TTL is the time-to-live for this request (optional).
	TTL *time.Duration `json:"ttl,omitempty"`
}

// TransferResponse represents the response to a transfer request.
type TransferResponse struct {
	// TransactionID is the unique identifier of the transaction.
	TransactionID string `json:"transactionId"`

	// BlockHash is the hash of the block containing the transaction.
	BlockHash string `json:"blockHash,omitempty"`

	// Timestamp is the time when the transaction was confirmed.
	Timestamp time.Time `json:"timestamp"`

	// Status is the final status of the transaction.
	Status types.TransactionStatus `json:"status"`

	// Fee is the transaction fee paid.
	Fee types.Amount `json:"fee,omitempty"`

	// ConfirmationCount is the number of confirmations.
	ConfirmationCount int `json:"confirmationCount,omitempty"`
}

// LockRequest represents a request to lock CBDC tokens.
type LockRequest struct {
	// WalletID is the wallet ID whose tokens will be locked.
	WalletID string `json:"walletId" validate:"required"`

	// Amount is the quantity of CBDC tokens to lock.
	Amount types.Amount `json:"amount" validate:"required,gt=0"`

	// Currency is the currency code (e.g., "INR", "AED").
	Currency types.CurrencyCode `json:"currency" validate:"required,len=3"`

	// LockDuration is how long the tokens should remain locked.
	LockDuration time.Duration `json:"lockDuration" validate:"required,gt=0"`

	// LockReason describes why the tokens are being locked.
	LockReason string `json:"lockReason,omitempty"`

	// ReferenceID is an external reference ID.
	ReferenceID string `json:"referenceId,omitempty"`

	// Metadata contains additional key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LockResponse represents the response to a lock request.
type LockResponse struct {
	// LockID is the unique identifier of the lock operation.
	LockID string `json:"lockId"`

	// TransactionID is the identifier of the underlying transaction.
	TransactionID string `json:"transactionId"`

	// Status is the status of the lock operation.
	Status types.TransactionStatus `json:"status"`

	// ExpiresAt is the time when the lock will expire.
	ExpiresAt time.Time `json:"expiresAt"`

	// Fee is the transaction fee paid.
	Fee types.Amount `json:"fee,omitempty"`
}

// BurnRequest represents a request to burn CBDC tokens.
type BurnRequest struct {
	// WalletID is the wallet ID whose tokens will be burned.
	WalletID string `json:"walletId" validate:"required"`

	// Amount is the quantity of CBDC tokens to burn.
	Amount types.Amount `json:"amount" validate:"required,gt=0"`

	// Currency is the currency code (e.g., "INR", "AED").
	Currency types.CurrencyCode `json:"currency" validate:"required,len=3"`

	// ReferenceID is an external reference ID.
	ReferenceID string `json:"referenceId,omitempty"`

	// BurnReason describes why the tokens are being burned.
	BurnReason string `json:"burnReason,omitempty"`

	// Metadata contains additional key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BurnResponse represents the response to a burn request.
type BurnResponse struct {
	// TransactionID is the unique identifier of the transaction.
	TransactionID string `json:"transactionId"`

	// BlockHash is the hash of the block containing the transaction.
	BlockHash string `json:"blockHash,omitempty"`

	// Timestamp is the time when the transaction was confirmed.
	Timestamp time.Time `json:"timestamp"`

	// Status is the final status of the transaction.
	Status types.TransactionStatus `json:"status"`

	// Fee is the transaction fee paid.
	Fee types.Amount `json:"fee,omitempty"`
}

// RedeemRequest represents a request to redeem CBDC tokens.
type RedeemRequest struct {
	// WalletID is the wallet ID whose tokens will be redeemed.
	WalletID string `json:"walletId" validate:"required"`

	// Amount is the quantity of CBDC tokens to redeem.
	Amount types.Amount `json:"amount" validate:"required,gt=0"`

	// Currency is the currency code (e.g., "INR", "AED").
	Currency types.CurrencyCode `json:"currency" validate:"required,len=3"`

	// FiatAccount is the account where fiat funds will be deposited.
	FiatAccount string `json:"fiatAccount" validate:"required"`

	// ReferenceID is an external reference ID.
	ReferenceID string `json:"referenceId,omitempty"`

	// Metadata contains additional key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RedeemResponse represents the response to a redeem request.
type RedeemResponse struct {
	// TransactionID is the unique identifier of the transaction.
	TransactionID string `json:"transactionId"`

	// BlockHash is the hash of the block containing the transaction.
	BlockHash string `json:"blockHash,omitempty"`

	// Timestamp is the time when the transaction was confirmed.
	Timestamp time.Time `json:"timestamp"`

	// Status is the final status of the transaction.
	Status types.TransactionStatus `json:"status"`

	// Fee is the transaction fee paid.
	Fee types.Amount `json:"fee,omitempty"`

	// FiatAmount is the amount of fiat currency received.
	FiatAmount types.Amount `json:"fiatAmount,omitempty"`
}

// BalanceRequest represents a request to check balance.
type BalanceRequest struct {
	// WalletID is the wallet ID to check.
	WalletID string `json:"walletId" validate:"required"`

	// Currency is the currency code (optional).
	Currency types.CurrencyCode `json:"currency,omitempty"`

	// IncludeReserved includes reserved/locked balance if true.
	IncludeReserved bool `json:"includeReserved,omitempty"`
}

// BalanceResponse represents a balance response.
type BalanceResponse struct {
	// WalletID is the wallet ID.
	WalletID string `json:"walletId"`

	// Available is the available balance.
	Available types.Amount `json:"available"`

	// Reserved is the reserved/locked balance.
	Reserved types.Amount `json:"reserved,omitempty"`

	// Total is the total balance (available + reserved).
	Total types.Amount `json:"total"`

	// Currency is the currency code.
	Currency types.CurrencyCode `json:"currency"`

	// UpdatedAt is the last update time.
	UpdatedAt time.Time `json:"updatedAt"`
}

// TransactionQueryRequest represents a request to query a transaction.
type TransactionQueryRequest struct {
	// TransactionID is the ID of the transaction to query.
	TransactionID string `json:"transactionId" validate:"required"`
}

// TransactionResponse represents a transaction response.
type TransactionResponse struct {
	// TransactionID is the unique identifier of the transaction.
	TransactionID string `json:"transactionId"`

	// Type is the type of the transaction.
	Type types.TransactionType `json:"type"`

	// Status is the current status of the transaction.
	Status types.TransactionStatus `json:"status"`

	// From is the source wallet/account.
	From string `json:"from,omitempty"`

	// To is the destination wallet/account.
	To string `json:"to,omitempty"`

	// Amount is the transaction amount.
	Amount types.Amount `json:"amount"`

	// Currency is the currency code.
	Currency types.CurrencyCode `json:"currency"`

	// Fee is the transaction fee.
	Fee types.Amount `json:"fee,omitempty"`

	// Timestamp is the transaction timestamp.
	Timestamp time.Time `json:"timestamp"`

	// Confirmations is the number of confirmations.
	Confirmations int `json:"confirmations,omitempty"`

	// BlockHash is the block hash if confirmed.
	BlockHash string `json:"blockHash,omitempty"`

	// Metadata contains additional data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionStatusRequest represents a request to check transaction status.
type TransactionStatusRequest struct {
	// TransactionID is the ID of the transaction.
	TransactionID string `json:"transactionId" validate:"required"`
}

// TransactionStatusResponse represents a transaction status response.
type TransactionStatusResponse struct {
	// TransactionID is the unique identifier of the transaction.
	TransactionID string `json:"transactionId"`

	// Status is the current status.
	Status types.TransactionStatus `json:"status"`

	// Confirmations is the number of confirmations.
	Confirmations int `json:"confirmations,omitempty"`

	// RequiredConfirmations is the required number of confirmations.
	RequiredConfirmations int `json:"requiredConfirmations,omitempty"`

	// IsFinal indicates if this is a final status.
	IsFinal bool `json:"isFinal"`

	// UpdatedAt is the last update time.
	UpdatedAt time.Time `json:"updatedAt"`
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	// Status is the overall health status.
	Status types.HealthStatus `json:"status"`

	// Components contains detailed health of each component.
	Components map[string]types.ComponentHealth `json:"components,omitempty"`

	// Version is the connector version.
	Version string `json:"version"`

	// Timestamp is the health check timestamp.
	Timestamp time.Time `json:"timestamp"`

	// Details contains additional health information.
	Details map[string]interface{} `json:"details,omitempty"`
}

// NetworkInfoResponse represents a network information response.
type NetworkInfoResponse struct {
	// NetworkID is the network identifier.
	NetworkID string `json:"networkId"`

	// NetworkName is the network name.
	NetworkName string `json:"networkName"`

	// ChainID is the chain ID (for blockchain networks).
	ChainID string `json:"chainId,omitempty"`

	// BlockHeight is the current block height.
	BlockHeight uint64 `json:"blockHeight,omitempty"`

	// BlockTime is the average block time.
	BlockTime time.Duration `json:"blockTime,omitempty"`

	// IsSyncing indicates if the node is syncing.
	IsSyncing bool `json:"isSyncing"`

	// ConnectedPeers is the number of connected peers.
	ConnectedPeers int `json:"connectedPeers,omitempty"`

	// LatestBlockHash is the latest block hash.
	LatestBlockHash string `json:"latestBlockHash,omitempty"`

	// Features lists supported features.
	Features []string `json:"features,omitempty"`
}
