//go:build integration

package integration

import (
    "database/sql"
    "testing"

    "github.com/fineract/cbdc/india-connector/internal/adapters/repository"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
    // Skip if no database connection
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test database
    db, err := setupTestDB()
    require.NoError(t, err)
    defer db.Close()

    // Run migrations
    err = repository.Migrate(db)
    require.NoError(t, err)

    // Check migration status
    status, err := repository.GetMigrationStatus(db)
    require.NoError(t, err)

    assert.Equal(t, 0, status.Pending, "all migrations should be applied")
    assert.Greater(t, status.Applied, 0, "migrations should be applied")

    // Validate all tables exist
    tables := []string{
        "transactions",
        "wallets",
        "locks",
        "audit_logs",
        "settlements",
        "webhook_events",
        "configs",
        "migrations",
    }

    for _, table := range tables {
        exists, err := tableExists(db, table)
        require.NoError(t, err)
        assert.True(t, exists, "table %s should exist", table)
    }

    // Validate configs were inserted
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM configs").Scan(&count)
    require.NoError(t, err)
    assert.Greater(t, count, 0, "configs should be populated")

    // Test rollback
    err = repository.Rollback(db, 1)
    require.NoError(t, err)

    status, err = repository.GetMigrationStatus(db)
    require.NoError(t, err)
    assert.Equal(t, 1, status.Pending, "one migration should be pending after rollback")

    // Run migrations again
    err = repository.Migrate(db)
    require.NoError(t, err)

    status, err = repository.GetMigrationStatus(db)
    require.NoError(t, err)
    assert.Equal(t, 0, status.Pending, "all migrations should be applied after re-run")
}

func TestTransactionRepository(t *testing.T) {
    // Skip if no database connection
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test database
    db, err := setupTestDB()
    require.NoError(t, err)
    defer db.Close()

    // Run migrations
    err = repository.Migrate(db)
    require.NoError(t, err)

    // Create repository
    repo := repository.NewTransactionRepository(db)

    // Test save and get
    // (Implementation would go here)
    t.Run("save and get transaction", func(t *testing.T) {
        // Test implementation
    })

    t.Run("get by transaction id", func(t *testing.T) {
        // Test implementation
    })

    t.Run("list transactions with filters", func(t *testing.T) {
        // Test implementation
    })
}

// Helper functions
func setupTestDB() (*sql.DB, error) {
    // Use test database configuration
    // This should connect to a test database
    connStr := "host=localhost port=5432 user=test_user password=test_pass dbname=test_cbdc sslmode=disable"
    return sql.Open("postgres", connStr)
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
    query := `
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.tables
            WHERE table_name = $1
        )
    `

    var exists bool
    err := db.QueryRow(query, tableName).Scan(&exists)
    return exists, err
}