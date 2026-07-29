package repository

import (
    "context"
    "fmt"
    "time"

    "github.com/apache/fineract-cbdc-orchestrator/internal/domain/transaction"
    "gorm.io/gorm"
)

// TransactionRepositoryImpl implements the transaction repository interface
type TransactionRepositoryImpl struct {
    db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) transaction.Repository {
    return &TransactionRepositoryImpl{db: db}
}

// Create creates a new transaction
func (r *TransactionRepositoryImpl) Create(ctx context.Context, tx *transaction.Transaction) error {
    return r.db.WithContext(ctx).Create(tx).Error
}

// Update updates an existing transaction
func (r *TransactionRepositoryImpl) Update(ctx context.Context, tx *transaction.Transaction) error {
    return r.db.WithContext(ctx).Save(tx).Error
}

// GetByID gets a transaction by ID
func (r *TransactionRepositoryImpl) GetByID(ctx context.Context, id string) (*transaction.Transaction, error) {
    var tx transaction.Transaction
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&tx).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }
    return &tx, nil
}

// GetByIdempotencyKey gets a transaction by idempotency key
func (r *TransactionRepositoryImpl) GetByIdempotencyKey(ctx context.Context, key string) (*transaction.Transaction, error) {
    if key == "" {
        return nil, nil
    }
    var tx transaction.Transaction
    err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&tx).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }
    return &tx, nil
}

// List lists transactions with filters
func (r *TransactionRepositoryImpl) List(ctx context.Context, filter *transaction.TransactionFilter) ([]*transaction.Transaction, int64, error) {
    query := r.db.WithContext(ctx).Model(&transaction.Transaction{})

    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.State != "" {
        query = query.Where("state = ?", filter.State)
    }
    if filter.Source != "" {
        query = query.Where("source_country = ?", filter.Source)
    }
    if filter.Target != "" {
        query = query.Where("target_country = ?", filter.Target)
    }
    if filter.FromDate != nil {
        query = query.Where("created_at >= ?", filter.FromDate)
    }
    if filter.ToDate != nil {
        query = query.Where("created_at <= ?", filter.ToDate)
    }

    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if filter.SortBy != "" {
        order := filter.SortBy
        if filter.SortOrder == "DESC" {
            order += " DESC"
        }
        query = query.Order(order)
    } else {
        query = query.Order("created_at DESC")
    }

    if filter.Limit > 0 {
        query = query.Limit(filter.Limit)
    }
    if filter.Offset > 0 {
        query = query.Offset(filter.Offset)
    }

    var transactions []*transaction.Transaction
    if err := query.Find(&transactions).Error; err != nil {
        return nil, 0, err
    }

    return transactions, total, nil
}

// Delete deletes a transaction
func (r *TransactionRepositoryImpl) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&transaction.Transaction{}, "id = ?", id).Error
}

// UpdateStatus updates the status of a transaction
func (r *TransactionRepositoryImpl) UpdateStatus(ctx context.Context, id string, status transaction.TransactionStatus) error {
    return r.db.WithContext(ctx).Model(&transaction.Transaction{}).
        Where("id = ?", id).
        Update("status", status).Error
}

// UpdateState updates the state of a transaction
func (r *TransactionRepositoryImpl) UpdateState(ctx context.Context, id string, state transaction.TransactionState) error {
    return r.db.WithContext(ctx).Model(&transaction.Transaction{}).
        Where("id = ?", id).
        Update("state", state).Error
}

// GetPendingTransactions gets pending transactions
func (r *TransactionRepositoryImpl) GetPendingTransactions(ctx context.Context) ([]*transaction.Transaction, error) {
    var transactions []*transaction.Transaction
    err := r.db.WithContext(ctx).
        Where("status IN ?", []transaction.TransactionStatus{
            transaction.StatusPending,
            transaction.StatusComplianceCheck,
            transaction.StatusFXProcessing,
            transaction.StatusLocking,
            transaction.StatusSettling,
            transaction.StatusReleasingLock,
        }).
        Order("created_at ASC").
        Find(&transactions).Error
    return transactions, err
}

// GetTransactionsByStatus gets transactions by status
func (r *TransactionRepositoryImpl) GetTransactionsByStatus(ctx context.Context, status transaction.TransactionStatus, limit int) ([]*transaction.Transaction, error) {
    var transactions []*transaction.Transaction
    query := r.db.WithContext(ctx).Where("status = ?", status)
    if limit > 0 {
        query = query.Limit(limit)
    }
    err := query.Order("created_at ASC").Find(&transactions).Error
    return transactions, err
}