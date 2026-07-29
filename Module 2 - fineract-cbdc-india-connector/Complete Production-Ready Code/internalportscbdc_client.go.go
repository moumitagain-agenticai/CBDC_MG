package ports

import (
    "context"

    "github.com/fineract/cbdc/india-connector/internal/domain"
)

// CBDCClient defines the interface for interacting with the Indian CBDC API
type CBDCClient interface {
    // Core Operations
    Issue(ctx context.Context, req *IssueRequest) (*IssueResponse, error)
    Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)
    Lock(ctx context.Context, req *LockRequest) (*LockResponse, error)
    Burn(ctx context.Context, req *BurnRequest) (*BurnResponse, error)
    Redeem(ctx context.Context, req *RedeemRequest) (*RedeemResponse, error)

    // Query Operations
    GetBalance(ctx context.Context, req *BalanceRequest) (*BalanceResponse, error)
    GetTransactionStatus(ctx context.Context, txID string) (*TransactionStatusResponse, error)

    // Health Check
    HealthCheck(ctx context.Context) (*HealthResponse, error)
}

// IssueRequest represents a request to issue CBDC tokens
type IssueRequest struct {
    WalletID   string `json:"wallet_id"`
    Amount     string `json:"amount"`
    Currency   string `json:"currency"`
    ReferenceID string `json:"reference_id"`
}

// IssueResponse represents a response from issuing CBDC tokens
type IssueResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Message       string `json:"message"`
}

// TransferRequest represents a request to transfer CBDC tokens
type TransferRequest struct {
    SourceWallet      string `json:"source_wallet"`
    DestinationWallet string `json:"destination_wallet"`
    Amount            string `json:"amount"`
    Currency          string `json:"currency"`
    ReferenceID       string `json:"reference_id"`
}

// TransferResponse represents a response from transferring CBDC tokens
type TransferResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Message       string `json:"message"`
}

// LockRequest represents a request to lock CBDC tokens
type LockRequest struct {
    WalletID    string `json:"wallet_id"`
    Amount      string `json:"amount"`
    Currency    string `json:"currency"`
    DurationSec int    `json:"duration_sec"`
    ReferenceID string `json:"reference_id"`
}

// LockResponse represents a response from locking CBDC tokens
type LockResponse struct {
    LockID        string `json:"lock_id"`
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    ExpiresAt     string `json:"expires_at"`
}

// BurnRequest represents a request to burn CBDC tokens
type BurnRequest struct {
    WalletID    string `json:"wallet_id"`
    Amount      string `json:"amount"`
    Currency    string `json:"currency"`
    ReferenceID string `json:"reference_id"`
}

// BurnResponse represents a response from burning CBDC tokens
type BurnResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Message       string `json:"message"`
}

// RedeemRequest represents a request to redeem CBDC tokens
type RedeemRequest struct {
    WalletID    string `json:"wallet_id"`
    Amount      string `json:"amount"`
    Currency    string `json:"currency"`
    Account     string `json:"account"`
    ReferenceID string `json:"reference_id"`
}

// RedeemResponse represents a response from redeeming CBDC tokens
type RedeemResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Message       string `json:"message"`
}

// BalanceRequest represents a request to get balance
type BalanceRequest struct {
    WalletID string `json:"wallet_id"`
    Currency string `json:"currency"`
}

// BalanceResponse represents a balance response
type BalanceResponse struct {
    WalletID  string `json:"wallet_id"`
    Balance   string `json:"balance"`
    Currency  string `json:"currency"`
    Available string `json:"available"`
    Locked    string `json:"locked"`
}

// TransactionStatusResponse represents a transaction status response
type TransactionStatusResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Confirmations int    `json:"confirmations"`
    CompletedAt   string `json:"completed_at"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
    Status    string `json:"status"`
    Version   string `json:"version"`
    Timestamp string `json:"timestamp"`
}