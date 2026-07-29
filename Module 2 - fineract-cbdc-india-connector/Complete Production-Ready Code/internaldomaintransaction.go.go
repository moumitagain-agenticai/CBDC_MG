package domain

import (
    "time"

    "github.com/google/uuid"
)

// Transaction represents a CBDC transaction
type Transaction struct {
    ID              string          `json:"id" db:"id"`
    TransactionID   string          `json:"transaction_id" db:"transaction_id"`
    Type            TransactionType `json:"type" db:"type"`
    Status          TransactionStatus `json:"status" db:"status"`
    Amount          string          `json:"amount" db:"amount"`
    Currency        string          `json:"currency" db:"currency"`
    SourceWallet    string          `json:"source_wallet" db:"source_wallet"`
    DestinationWallet string        `json:"destination_wallet" db:"destination_wallet"`
    ReferenceID     string          `json:"reference_id" db:"reference_id"`
    Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
    ErrorCode       string          `json:"error_code,omitempty" db:"error_code"`
    ErrorMessage    string          `json:"error_message,omitempty" db:"error_message"`
    CreatedAt       time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
    CompletedAt     *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// TransactionType represents the type of CBDC transaction
type TransactionType string

const (
    TransactionTypeIssue   TransactionType = "issue"
    TransactionTypeTransfer TransactionType = "transfer"
    TransactionTypeLock     TransactionType = "lock"
    TransactionTypeBurn     TransactionType = "burn"
    TransactionTypeRedeem   TransactionType = "redeem"
)

// TransactionStatus represents the status of a CBDC transaction
type TransactionStatus string

const (
    TransactionStatusPending   TransactionStatus = "pending"
    TransactionStatusProcessing TransactionStatus = "processing"
    TransactionStatusCompleted  TransactionStatus = "completed"
    TransactionStatusFailed     TransactionStatus = "failed"
    TransactionStatusCancelled  TransactionStatus = "cancelled"
)

// TransactionRequest represents a request to perform a CBDC transaction
type TransactionRequest struct {
    Type             TransactionType `json:"type" validate:"required"`
    Amount           string          `json:"amount" validate:"required"`
    Currency         string          `json:"currency" validate:"required,len=3"`
    SourceWallet     string          `json:"source_wallet" validate:"required"`
    DestinationWallet string         `json:"destination_wallet,omitempty"`
    ReferenceID      string          `json:"reference_id" validate:"required"`
    Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionResponse represents the response from a CBDC transaction
type TransactionResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Message       string `json:"message,omitempty"`
    Details       map[string]interface{} `json:"details,omitempty"`
}

// NewTransaction creates a new transaction
func NewTransaction(req *TransactionRequest) *Transaction {
    now := time.Now()
    return &Transaction{
        ID:                uuid.New().String(),
        TransactionID:     uuid.New().String(),
        Type:              req.Type,
        Status:            TransactionStatusPending,
        Amount:            req.Amount,
        Currency:          req.Currency,
        SourceWallet:      req.SourceWallet,
        DestinationWallet: req.DestinationWallet,
        ReferenceID:       req.ReferenceID,
        Metadata:          req.Metadata,
        CreatedAt:         now,
        UpdatedAt:         now,
    }
}