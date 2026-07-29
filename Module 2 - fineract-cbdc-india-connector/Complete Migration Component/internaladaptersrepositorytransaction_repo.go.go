package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/fineract/cbdc/india-connector/internal/domain"
    "github.com/fineract/cbdc/india-connector/internal/ports"
    "github.com/google/uuid"
    "github.com/lib/pq"
)

// TransactionRepository implements the TransactionRepository interface
type TransactionRepository struct {
    db *sql.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
    return &TransactionRepository{db: db}
}

// Save saves a transaction
func (r *TransactionRepository) Save(ctx context.Context, tx *domain.Transaction) error {
    query := `
        INSERT INTO transactions (
            id, transaction_id, type, status, amount, currency,
            source_wallet, destination_wallet, reference_id, metadata,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

    metadataJSON, err := json.Marshal(tx.Metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    _, err = r.db.ExecContext(ctx, query,
        tx.ID,
        tx.TransactionID,
        string(tx.Type),
        string(tx.Status),
        tx.Amount,
        tx.Currency,
        tx.SourceWallet,
        tx.DestinationWallet,
        tx.ReferenceID,
        metadataJSON,
        tx.CreatedAt,
        tx.UpdatedAt,
    )

    if err != nil {
        return fmt.Errorf("failed to save transaction: %w", err)
    }

    return nil
}

// Update updates a transaction
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
    query := `
        UPDATE transactions SET
            status = $1,
            error_code = $2,
            error_message = $3,
            updated_at = $4,
            completed_at = $5
        WHERE id = $6
    `

    _, err := r.db.ExecContext(ctx, query,
        string(tx.Status),
        tx.ErrorCode,
        tx.ErrorMessage,
        time.Now(),
        tx.CompletedAt,
        tx.ID,
    )

    if err != nil {
        return fmt.Errorf("failed to update transaction: %w", err)
    }

    return nil
}

// GetByID gets a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
    query := `
        SELECT id, transaction_id, type, status, amount, currency,
               source_wallet, destination_wallet, reference_id, metadata,
               error_code, error_message, created_at, updated_at, completed_at
        FROM transactions
        WHERE id = $1
    `

    var tx domain.Transaction
    var metadataJSON []byte
    var destWallet sql.NullString
    var errorCode sql.NullString
    var errorMessage sql.NullString
    var completedAt pq.NullTime

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &tx.ID,
        &tx.TransactionID,
        &tx.Type,
        &tx.Status,
        &tx.Amount,
        &tx.Currency,
        &tx.SourceWallet,
        &destWallet,
        &tx.ReferenceID,
        &metadataJSON,
        &errorCode,
        &errorMessage,
        &tx.CreatedAt,
        &tx.UpdatedAt,
        &completedAt,
    )

    if err == sql.ErrNoRows {
        return nil, ports.ErrTransactionNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction: %w", err)
    }

    if destWallet.Valid {
        tx.DestinationWallet = destWallet.String
    }
    if errorCode.Valid {
        tx.ErrorCode = errorCode.String
    }
    if errorMessage.Valid {
        tx.ErrorMessage = errorMessage.String
    }
    if completedAt.Valid {
        tx.CompletedAt = &completedAt.Time
    }

    if len(metadataJSON) > 0 {
        if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
        }
    }

    return &tx, nil
}

// GetByTransactionID gets a transaction by transaction ID
func (r *TransactionRepository) GetByTransactionID(ctx context.Context, transactionID string) (*domain.Transaction, error) {
    query := `
        SELECT id, transaction_id, type, status, amount, currency,
               source_wallet, destination_wallet, reference_id, metadata,
               error_code, error_message, created_at, updated_at, completed_at
        FROM transactions
        WHERE transaction_id = $1
    `

    var tx domain.Transaction
    var metadataJSON []byte
    var destWallet sql.NullString
    var errorCode sql.NullString
    var errorMessage sql.NullString
    var completedAt pq.NullTime

    err := r.db.QueryRowContext(ctx, query, transactionID).Scan(
        &tx.ID,
        &tx.TransactionID,
        &tx.Type,
        &tx.Status,
        &tx.Amount,
        &tx.Currency,
        &tx.SourceWallet,
        &destWallet,
        &tx.ReferenceID,
        &metadataJSON,
        &errorCode,
        &errorMessage,
        &tx.CreatedAt,
        &tx.UpdatedAt,
        &completedAt,
    )

    if err == sql.ErrNoRows {
        return nil, ports.ErrTransactionNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction: %w", err)
    }

    if destWallet.Valid {
        tx.DestinationWallet = destWallet.String
    }
    if errorCode.Valid {
        tx.ErrorCode = errorCode.String
    }
    if errorMessage.Valid {
        tx.ErrorMessage = errorMessage.String
    }
    if completedAt.Valid {
        tx.CompletedAt = &completedAt.Time
    }

    if len(metadataJSON) > 0 {
        if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
        }
    }

    return &tx, nil
}

// GetByReferenceID gets a transaction by reference ID
func (r *TransactionRepository) GetByReferenceID(ctx context.Context, referenceID string) (*domain.Transaction, error) {
    query := `
        SELECT id, transaction_id, type, status, amount, currency,
               source_wallet, destination_wallet, reference_id, metadata,
               error_code, error_message, created_at, updated_at, completed_at
        FROM transactions
        WHERE reference_id = $1
    `

    var tx domain.Transaction
    var metadataJSON []byte
    var destWallet sql.NullString
    var errorCode sql.NullString
    var errorMessage sql.NullString
    var completedAt pq.NullTime

    err := r.db.QueryRowContext(ctx, query, referenceID).Scan(
        &tx.ID,
        &tx.TransactionID,
        &tx.Type,
        &tx.Status,
        &tx.Amount,
        &tx.Currency,
        &tx.SourceWallet,
        &destWallet,
        &tx.ReferenceID,
        &metadataJSON,
        &errorCode,
        &errorMessage,
        &tx.CreatedAt,
        &tx.UpdatedAt,
        &completedAt,
    )

    if err == sql.ErrNoRows {
        return nil, ports.ErrTransactionNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction: %w", err)
    }

    if destWallet.Valid {
        tx.DestinationWallet = destWallet.String
    }
    if errorCode.Valid {
        tx.ErrorCode = errorCode.String
    }
    if errorMessage.Valid {
        tx.ErrorMessage = errorMessage.String
    }
    if completedAt.Valid {
        tx.CompletedAt = &completedAt.Time
    }

    if len(metadataJSON) > 0 {
        if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
        }
    }

    return &tx, nil
}

// List lists transactions with filters
func (r *TransactionRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Transaction, int, error) {
    // Build query with filters
    query := `
        SELECT id, transaction_id, type, status, amount, currency,
               source_wallet, destination_wallet, reference_id, metadata,
               error_code, error_message, created_at, updated_at, completed_at
        FROM transactions
        WHERE 1=1
    `

    var args []interface{}
    argIndex := 1

    if wallet, ok := filters["wallet"]; ok {
        query += fmt.Sprintf(" AND (source_wallet = $%d OR destination_wallet = $%d)", argIndex, argIndex)
        args = append(args, wallet)
        argIndex++
    }

    if status, ok := filters["status"]; ok {
        query += fmt.Sprintf(" AND status = $%d", argIndex)
        args = append(args, status)
        argIndex++
    }

    if txType, ok := filters["type"]; ok {
        query += fmt.Sprintf(" AND type = $%d", argIndex)
        args = append(args, txType)
        argIndex++
    }

    if from, ok := filters["from"]; ok {
        query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
        args = append(args, from)
        argIndex++
    }

    if to, ok := filters["to"]; ok {
        query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
        args = append(args, to)
        argIndex++
    }

    // Get total count
    countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
    var total int
    err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get total count: %w", err)
    }

    // Add ordering and pagination
    query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
    args = append(args, limit, offset)

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
    }
    defer rows.Close()

    var transactions []*domain.Transaction

    for rows.Next() {
        var tx domain.Transaction
        var metadataJSON []byte
        var destWallet sql.NullString
        var errorCode sql.NullString
        var errorMessage sql.NullString
        var completedAt pq.NullTime

        err := rows.Scan(
            &tx.ID,
            &tx.TransactionID,
            &tx.Type,
            &tx.Status,
            &tx.Amount,
            &tx.Currency,
            &tx.SourceWallet,
            &destWallet,
            &tx.ReferenceID,
            &metadataJSON,
            &errorCode,
            &errorMessage,
            &tx.CreatedAt,
            &tx.UpdatedAt,
            &completedAt,
        )
        if err != nil {
            return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
        }

        if destWallet.Valid {
            tx.DestinationWallet = destWallet.String
        }
        if errorCode.Valid {
            tx.ErrorCode = errorCode.String
        }
        if errorMessage.Valid {
            tx.ErrorMessage = errorMessage.String
        }
        if completedAt.Valid {
            tx.CompletedAt = &completedAt.Time
        }

        if len(metadataJSON) > 0 {
            if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
                return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
            }
        }

        transactions = append(transactions, &tx)
    }

    if err := rows.Err(); err != nil {
        return nil, 0, fmt.Errorf("error iterating transactions: %w", err)
    }

    return transactions, total, nil
}

// GetTransactionCount gets the count of transactions
func (r *TransactionRepository) GetTransactionCount(ctx context.Context) (int, error) {
    query := "SELECT COUNT(*) FROM transactions"

    var count int
    err := r.db.QueryRowContext(ctx, query).Scan(&count)
    if err != nil {
        return 0, fmt.Errorf("failed to get transaction count: %w", err)
    }

    return count, nil
}

// Delete deletes a transaction (soft delete - just for cleanup)
func (r *TransactionRepository) Delete(ctx context.Context, id string) error {
    // In production, we might want to soft delete or archive
    // For now, we'll do a hard delete since this is a connector
    // and transactions should be immutable
    query := "DELETE FROM transactions WHERE id = $1"

    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete transaction: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rows == 0 {
        return ports.ErrTransactionNotFound
    }

    return nil
}