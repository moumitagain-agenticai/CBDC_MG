package repository

import (
    "database/sql"
    "fmt"
    "strings"

    "go.uber.org/zap"
)

// Migration represents a database migration
type Migration struct {
    Version int
    Up      string
    Down    string
}

// GetMigrations returns all migrations in order
func GetMigrations() []Migration {
    return []Migration{
        {
            Version: 1,
            Up: `
                CREATE TABLE IF NOT EXISTS transactions (
                    id VARCHAR(36) PRIMARY KEY,
                    transaction_id VARCHAR(36) UNIQUE NOT NULL,
                    type VARCHAR(20) NOT NULL,
                    status VARCHAR(20) NOT NULL,
                    amount DECIMAL(20,4) NOT NULL,
                    currency VARCHAR(3) NOT NULL,
                    source_wallet VARCHAR(100) NOT NULL,
                    destination_wallet VARCHAR(100),
                    reference_id VARCHAR(100) NOT NULL,
                    metadata JSONB,
                    error_code VARCHAR(50),
                    error_message TEXT,
                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    completed_at TIMESTAMP WITH TIME ZONE
                );

                CREATE INDEX idx_transactions_type ON transactions(type);
                CREATE INDEX idx_transactions_status ON transactions(status);
                CREATE INDEX idx_transactions_reference_id ON transactions(reference_id);
                CREATE INDEX idx_transactions_created_at ON transactions(created_at);
                CREATE INDEX idx_transactions_source_wallet ON transactions(source_wallet);
                CREATE INDEX idx_transactions_destination_wallet ON transactions(destination_wallet);
            `,
            Down: `
                DROP TABLE IF EXISTS transactions;
            `,
        },
        {
            Version: 2,
            Up: `
                CREATE TABLE IF NOT EXISTS wallets (
                    id VARCHAR(36) PRIMARY KEY,
                    wallet_id VARCHAR(100) UNIQUE NOT NULL,
                    currency VARCHAR(3) NOT NULL,
                    balance DECIMAL(20,4) NOT NULL DEFAULT 0,
                    locked_balance DECIMAL(20,4) NOT NULL DEFAULT 0,
                    status VARCHAR(20) NOT NULL DEFAULT 'active',
                    customer_id VARCHAR(36),
                    metadata JSONB,
                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
                );

                CREATE INDEX idx_wallets_wallet_id ON wallets(wallet_id);
                CREATE INDEX idx_wallets_customer_id ON wallets(customer_id);
                CREATE INDEX idx_wallets_currency ON wallets(currency);
                CREATE UNIQUE INDEX idx_wallets_wallet_currency ON wallets(wallet_id, currency);
            `,
            Down: `
                DROP TABLE IF EXISTS wallets;
            `,
        },
        {
            Version: 3,
            Up: `
                CREATE TABLE IF NOT EXISTS locks (
                    id VARCHAR(36) PRIMARY KEY,
                    lock_id VARCHAR(36) UNIQUE NOT NULL,
                    transaction_id VARCHAR(36) NOT NULL REFERENCES transactions(transaction_id),
                    wallet_id VARCHAR(100) NOT NULL,
                    amount DECIMAL(20,4) NOT NULL,
                    currency VARCHAR(3) NOT NULL,
                    status VARCHAR(20) NOT NULL DEFAULT 'locked',
                    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
                    released_at TIMESTAMP WITH TIME ZONE,
                    metadata JSONB,
                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
                );

                CREATE INDEX idx_locks_lock_id ON locks(lock_id);
                CREATE INDEX idx_locks_transaction_id ON locks(transaction_id);
                CREATE INDEX idx_locks_wallet_id ON locks(wallet_id);
                CREATE INDEX idx_locks_status ON locks(status);
                CREATE INDEX idx_locks_expires_at ON locks(expires_at);
            `,
            Down: `
                DROP TABLE IF EXISTS locks;
            `,
        },
        {
            Version: 4,
            Up: `
                CREATE TABLE IF NOT EXISTS audit_logs (
                    id VARCHAR(36) PRIMARY KEY,
                    transaction_id VARCHAR(36) REFERENCES transactions(transaction_id),
                    action VARCHAR(50) NOT NULL,
                    entity_type VARCHAR(50) NOT NULL,
                    entity_id VARCHAR(100) NOT NULL,
                    user_id VARCHAR(100),
                    ip_address VARCHAR(45),
                    user_agent TEXT,
                    old_value JSONB,
                    new_value JSONB,
                    metadata JSONB,
                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
                );

                CREATE INDEX idx_audit_logs_transaction_id ON audit_logs(transaction_id);
                CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
                CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
                CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
                CREATE INDEX idx_audit_logs_action ON audit_logs(action);
            `,
            Down: `
                DROP TABLE IF EXISTS audit_logs;
            `,
        },
        {
            Version: 5,
            Up: `
                CREATE TABLE IF NOT EXISTS settlements (
                    id VARCHAR(36) PRIMARY KEY,
                    settlement_id VARCHAR(36) UNIQUE NOT NULL,
                    type VARCHAR(20) NOT NULL,
                    status VARCHAR(20) NOT NULL,
                    source_currency VARCHAR(3) NOT NULL,
                    source_amount DECIMAL(20,4) NOT NULL,
                    destination_currency VARCHAR(3) NOT NULL,
                    destination_amount DECIMAL(20,4) NOT NULL,
                    fx_rate DECIMAL(20,8),
                    reference_id VARCHAR(100),
                    transactions JSONB,
                    metadata JSONB,
                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    completed_at TIMESTAMP WITH TIME ZONE
                );

                CREATE INDEX idx_settlements_settlement_id ON settlements(settlement_id);
                CREATE INDEX idx_settlements_status ON settlements(status);
                CREATE INDEX idx_settlements_type ON settlements(type);
                CREATE INDEX idx_settlements_reference_id ON settlements(reference_id);
                CREATE INDEX idx_settlements_created_at ON settlements(created_at);
            `,
            Down: `
                DROP TABLE IF EXISTS settlements;
            `,
        },
        {
            Version: 6,
            Up: `
                CREATE TABLE IF NOT EXISTS webhook_events (
                    id VARCHAR(36) PRIMARY KEY,
                    event_id VARCHAR(36) UNIQUE NOT NULL,
                    type VARCHAR(50) NOT NULL,
                    status VARCHAR(20) NOT NULL DEFAULT 'pending',
                    payload JSONB NOT NULL,
                    headers JSONB,
                    attempts INT NOT NULL DEFAULT 0,
                    max_attempts INT NOT NULL DEFAULT 3,
                    last_attempt_at TIMESTAMP WITH TIME ZONE,
                    next_attempt_at TIMESTAMP WITH TIME ZONE,
                    completed_at TIMESTAMP WITH TIME ZONE,
                    error_message TEXT,
                    metadata JSONB,
                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
                );

                CREATE INDEX idx_webhook_events_event_id ON webhook_events(event_id);
                CREATE INDEX idx_webhook_events_status ON webhook_events(status);
                CREATE INDEX idx_webhook_events_type ON webhook_events(type);
                CREATE INDEX idx_webhook_events_next_attempt_at ON webhook_events(next_attempt_at);
                CREATE INDEX idx_webhook_events_created_at ON webhook_events(created_at);
            `,
            Down: `
                DROP TABLE IF EXISTS webhook_events;
            `,
        },
        {
            Version: 7,
            Up: `
                CREATE TABLE IF NOT EXISTS configs (
                    key VARCHAR(100) PRIMARY KEY,
                    value JSONB NOT NULL,
                    description TEXT,
                    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
                );

                INSERT INTO configs (key, value, description) VALUES
                    ('transaction.timeout', '{"seconds": 300}', 'Transaction timeout in seconds'),
                    ('transaction.max_amount', '{"amount": "1000000.00", "currency": "INR"}', 'Maximum transaction amount'),
                    ('wallet.daily_limit', '{"amount": "50000.00", "currency": "INR"}', 'Daily wallet transfer limit'),
                    ('lock.max_duration', '{"seconds": 3600}', 'Maximum lock duration in seconds'),
                    ('retry.max_attempts', '{"attempts": 3}', 'Maximum retry attempts for failed transactions'),
                    ('circuit_breaker.threshold', '{"failure_ratio": 0.5, "min_requests": 10}', 'Circuit breaker threshold configuration');
            `,
            Down: `
                DROP TABLE IF EXISTS configs;
            `,
        },
    }
}

// Migrate runs all pending migrations
func Migrate(db *sql.DB) error {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Create migrations table if it doesn't exist
    if err := createMigrationsTable(db); err != nil {
        return fmt.Errorf("failed to create migrations table: %w", err)
    }

    // Get current version
    currentVersion, err := getCurrentVersion(db)
    if err != nil {
        return fmt.Errorf("failed to get current version: %w", err)
    }

    migrations := GetMigrations()
    targetVersion := len(migrations)

    logger.Info("running database migrations",
        zap.Int("current_version", currentVersion),
        zap.Int("target_version", targetVersion),
    )

    // Run migrations
    for i := currentVersion; i < targetVersion; i++ {
        migration := migrations[i]
        version := i + 1

        logger.Info("running migration",
            zap.Int("version", version),
            zap.String("up_sql", truncateSQL(migration.Up)),
        )

        // Begin transaction
        tx, err := db.Begin()
        if err != nil {
            return fmt.Errorf("failed to begin transaction for migration %d: %w", version, err)
        }

        // Execute migration
        if _, err := tx.Exec(migration.Up); err != nil {
            tx.Rollback()
            logger.Error("migration failed",
                zap.Int("version", version),
                zap.Error(err),
            )
            return fmt.Errorf("failed to execute migration %d: %w", version, err)
        }

        // Record migration
        if _, err := tx.Exec(`
            INSERT INTO migrations (version, applied_at, up_sql)
            VALUES ($1, NOW(), $2)
        `, version, migration.Up); err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to record migration %d: %w", version, err)
        }

        // Commit transaction
        if err := tx.Commit(); err != nil {
            return fmt.Errorf("failed to commit migration %d: %w", version, err)
        }

        logger.Info("migration completed", zap.Int("version", version))
    }

    logger.Info("all migrations completed successfully")
    return nil
}

// Rollback rolls back migrations
func Rollback(db *sql.DB, steps int) error {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Get current version
    currentVersion, err := getCurrentVersion(db)
    if err != nil {
        return fmt.Errorf("failed to get current version: %w", err)
    }

    if currentVersion == 0 {
        logger.Info("no migrations to rollback")
        return nil
    }

    startVersion := currentVersion
    endVersion := startVersion - steps
    if endVersion < 0 {
        endVersion = 0
    }

    logger.Info("rolling back migrations",
        zap.Int("current_version", startVersion),
        zap.Int("target_version", endVersion),
        zap.Int("steps", steps),
    )

    migrations := GetMigrations()

    for version := startVersion; version > endVersion; version-- {
        migration := migrations[version-1]

        logger.Info("rolling back migration",
            zap.Int("version", version),
            zap.String("down_sql", truncateSQL(migration.Down)),
        )

        // Begin transaction
        tx, err := db.Begin()
        if err != nil {
            return fmt.Errorf("failed to begin transaction for rollback: %w", err)
        }

        // Execute rollback
        if _, err := tx.Exec(migration.Down); err != nil {
            tx.Rollback()
            logger.Error("rollback failed",
                zap.Int("version", version),
                zap.Error(err),
            )
            return fmt.Errorf("failed to rollback migration %d: %w", version, err)
        }

        // Remove migration record
        if _, err := tx.Exec(`
            DELETE FROM migrations WHERE version = $1
        `, version); err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to remove migration record %d: %w", version, err)
        }

        if err := tx.Commit(); err != nil {
            return fmt.Errorf("failed to commit rollback: %w", err)
        }

        logger.Info("rollback completed", zap.Int("version", version))
    }

    logger.Info("all rollbacks completed successfully")
    return nil
}

// createMigrationsTable creates the migrations table
func createMigrationsTable(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id SERIAL PRIMARY KEY,
            version INT NOT NULL UNIQUE,
            applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            up_sql TEXT,
            down_sql TEXT
        )
    `)
    return err
}

// getCurrentVersion gets the current migration version
func getCurrentVersion(db *sql.DB) (int, error) {
    var version int
    err := db.QueryRow(`
        SELECT COALESCE(MAX(version), 0) FROM migrations
    `).Scan(&version)
    if err != nil {
        return 0, err
    }
    return version, nil
}

// truncateSQL truncates SQL for logging
func truncateSQL(sql string) string {
    if len(sql) > 200 {
        return sql[:200] + "..."
    }
    return strings.TrimSpace(sql)
}

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
    CurrentVersion int `json:"current_version"`
    TargetVersion  int `json:"target_version"`
    Pending        int `json:"pending"`
    Applied        int `json:"applied"`
}

// GetMigrationStatus gets the migration status
func GetMigrationStatus(db *sql.DB) (*MigrationStatus, error) {
    currentVersion, err := getCurrentVersion(db)
    if err != nil {
        return nil, err
    }

    migrations := GetMigrations()
    targetVersion := len(migrations)

    return &MigrationStatus{
        CurrentVersion: currentVersion,
        TargetVersion:  targetVersion,
        Pending:        targetVersion - currentVersion,
        Applied:        currentVersion,
    }, nil
}

// ValidateMigrations validates that all migrations are applied
func ValidateMigrations(db *sql.DB) error {
    status, err := GetMigrationStatus(db)
    if err != nil {
        return err
    }

    if status.Pending > 0 {
        return fmt.Errorf("database has %d pending migrations", status.Pending)
    }

    return nil
}