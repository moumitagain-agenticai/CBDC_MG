package transaction

import (
    "context"
    "time"
)

// Repository defines the transaction repository interface
type Repository interface {
    // Create creates a new transaction
    Create(ctx context.Context, tx *Transaction) error

    // Update updates an existing transaction
    Update(ctx context.Context, tx *Transaction) error

    // GetByID gets a transaction by ID
    GetByID(ctx context.Context, id string) (*Transaction, error)

    // GetByIdempotencyKey gets a transaction by idempotency key
    GetByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)

    // List lists transactions with filters
    List(ctx context.Context, filter *TransactionFilter) ([]*Transaction, int64, error)

    // Delete deletes a transaction
    Delete(ctx context.Context, id string) error

    // UpdateStatus updates the status of a transaction
    UpdateStatus(ctx context.Context, id string, status TransactionStatus) error

    // UpdateState updates the state of a transaction
    UpdateState(ctx context.Context, id string, state TransactionState) error

    // GetPendingTransactions gets pending transactions
    GetPendingTransactions(ctx context.Context) ([]*Transaction, error)

    // GetTransactionsByStatus gets transactions by status
    GetTransactionsByStatus(ctx context.Context, status TransactionStatus, limit int) ([]*Transaction, error)
}

// TransactionFilter represents filters for listing transactions
type TransactionFilter struct {
    Status     TransactionStatus `json:"status"`
    State      TransactionState  `json:"state"`
    Source     string            `json:"source"`
    Target     string            `json:"target"`
    FromDate   *time.Time        `json:"fromDate"`
    ToDate     *time.Time        `json:"toDate"`
    Limit      int               `json:"limit"`
    Offset     int               `json:"offset"`
    SortBy     string            `json:"sortBy"`
    SortOrder  string            `json:"sortOrder"`
}