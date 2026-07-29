package ports

import (
    "context"
    "errors"

    "github.com/fineract/cbdc/india-connector/internal/domain"
)

// Common repository errors
var (
    ErrTransactionNotFound = errors.New("transaction not found")
    ErrWalletNotFound      = errors.New("wallet not found")
    ErrLockNotFound        = errors.New("lock not found")
    ErrDuplicateTransaction = errors.New("duplicate transaction")
)

// TransactionRepository defines the interface for transaction persistence
type TransactionRepository interface {
    // Save saves a transaction
    Save(ctx context.Context, tx *domain.Transaction) error

    // Update updates a transaction
    Update(ctx context.Context, tx *domain.Transaction) error

    // GetByID gets a transaction by ID
    GetByID(ctx context.Context, id string) (*domain.Transaction, error)

    // GetByTransactionID gets a transaction by transaction ID
    GetByTransactionID(ctx context.Context, transactionID string) (*domain.Transaction, error)

    // GetByReferenceID gets a transaction by reference ID
    GetByReferenceID(ctx context.Context, referenceID string) (*domain.Transaction, error)

    // List lists transactions with filters
    List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Transaction, int, error)

    // GetTransactionCount gets the count of transactions
    GetTransactionCount(ctx context.Context) (int, error)

    // Delete deletes a transaction
    Delete(ctx context.Context, id string) error
}

// WalletRepository defines the interface for wallet persistence
type WalletRepository interface {
    // Save saves a wallet
    Save(ctx context.Context, wallet *domain.Wallet) error

    // Update updates a wallet
    Update(ctx context.Context, wallet *domain.Wallet) error

    // GetByWalletID gets a wallet by wallet ID
    GetByWalletID(ctx context.Context, walletID string) (*domain.Wallet, error)

    // GetByCustomerID gets wallets by customer ID
    GetByCustomerID(ctx context.Context, customerID string) ([]*domain.Wallet, error)

    // List lists wallets with filters
    List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Wallet, int, error)

    // GetBalance gets the balance of a wallet
    GetBalance(ctx context.Context, walletID, currency string) (*domain.Balance, error)

    // UpdateBalance updates the balance of a wallet
    UpdateBalance(ctx context.Context, walletID, currency string, amount string, operation string) error
}

// LockRepository defines the interface for lock persistence
type LockRepository interface {
    // Save saves a lock
    Save(ctx context.Context, lock *domain.Lock) error

    // Update updates a lock
    Update(ctx context.Context, lock *domain.Lock) error

    // GetByLockID gets a lock by lock ID
    GetByLockID(ctx context.Context, lockID string) (*domain.Lock, error)

    // GetByTransactionID gets a lock by transaction ID
    GetByTransactionID(ctx context.Context, transactionID string) (*domain.Lock, error)

    // GetActiveLocks gets active locks for a wallet
    GetActiveLocks(ctx context.Context, walletID string) ([]*domain.Lock, error)

    // GetExpiredLocks gets expired locks
    GetExpiredLocks(ctx context.Context) ([]*domain.Lock, error)

    // Release releases a lock
    Release(ctx context.Context, lockID string) error
}