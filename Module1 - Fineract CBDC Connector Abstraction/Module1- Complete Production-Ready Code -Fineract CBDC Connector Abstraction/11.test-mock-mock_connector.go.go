package mock

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/google/uuid"
    "github.com/fineract/cbdc/connector-abstraction"
    "github.com/fineract/cbdc/connector-abstraction/config"
    "github.com/fineract/cbdc/connector-abstraction/types"
)

// MockConnector is a mock implementation of the CBDCConnector interface.
type MockConnector struct {
    mu sync.RWMutex
    
    // Configuration
    config *config.ConnectorConfig
    status connector.ConnectorStatus
    
    // State
    balances map[string]map[types.CurrencyCode]*types.Amount
    locks    map[string]*mockLock
    txs      map[string]*mockTransaction
    
    // Behavior controls
    simulateError         bool
    errorCode             connector.ErrorCode
    simulateLatency       time.Duration
    failOnOperation       map[string]bool
    allowedOperations     map[string]bool
    transactionHistory    []*connector.TransactionResponse
}

// mockLock represents a mock lock.
type mockLock struct {
    ID        string
    WalletID  string
    Amount    types.Amount
    Currency  types.CurrencyCode
    ExpiresAt time.Time
    Status    string
}

// mockTransaction represents a mock transaction.
type mockTransaction struct {
    ID        string
    Type      types.TransactionType
    Status    types.TransactionStatus
    From      string
    To        string
    Amount    types.Amount
    Currency  types.CurrencyCode
    Fee       types.Amount
    Timestamp time.Time
    Metadata  map[string]interface{}
}

// NewMockConnector creates a new mock connector.
func NewMockConnector(cfg *config.ConnectorConfig) *MockConnector {
    return &MockConnector{
        config:      cfg,
        status:      connector.StatusUninitialized,
        balances:    make(map[string]map[types.CurrencyCode]*types.Amount),
        locks:       make(map[string]*mockLock),
        txs:         make(map[string]*mockTransaction),
        failOnOperation: make(map[string]bool),
        allowedOperations: make(map[string]bool),
    }
}

// Initialize implements the connector interface.
func (m *MockConnector) Initialize(ctx context.Context, cfg *config.ConnectorConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.config = cfg
    m.status = connector.StatusReady
    return nil
}

// Shutdown implements the connector interface.
func (m *MockConnector) Shutdown(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.status = connector.StatusUninitialized
    return nil
}

// Status implements the connector interface.
func (m *MockConnector) Status() connector.ConnectorStatus {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.status
}

// HealthCheck implements the connector interface.
func (m *MockConnector) HealthCheck(ctx context.Context) (*connector.HealthResponse, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if m.simulateError {
        return nil, connector.NewError(connector.ErrInternal, "simulated health check error", nil)
    }
    
    return &connector.HealthResponse{
        Status: types.HealthStatusUp,
        Components: map[string]types.ComponentHealth{
            "network": {
                Status: types.HealthStatusUp,
                CheckedAt: time.Now(),
            },
            "api": {
                Status: types.HealthStatusUp,
                CheckedAt: time.Now(),
            },
        },
        Version: "1.0.0",
        Timestamp: time.Now(),
    }, nil
}

// GetNetworkInfo implements the connector interface.
func (m *MockConnector) GetNetworkInfo(ctx context.Context) (*connector.NetworkInfoResponse, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if m.simulateError {
        return nil, connector.NewError(connector.ErrInternal, "simulated network info error", nil)
    }
    
    return &connector.NetworkInfoResponse{
        NetworkID:   "mock-network",
        NetworkName: "Mock CBDC Network",
        BlockHeight: 1000000,
        IsSyncing:   false,
        ConnectedPeers: 10,
        Features:    []string{"issue", "transfer", "lock", "burn", "redeem"},
    }, nil
}

// Issue implements the connector interface.
func (m *MockConnector) Issue(ctx context.Context, req *connector.IssueRequest) (*connector.IssueResponse, error) {
    if err := m.checkOperation("issue"); err != nil {
        return nil, err
    }
    
    if err := m.simulateOperation(ctx); err != nil {
        return nil, err
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Create the transaction
    txID := uuid.New().String()
    tx := &mockTransaction{
        ID:        txID,
        Type:      types.TypeIssue,
        Status:    types.StatusConfirmed,
        To:        req.WalletID,
        Amount:    req.Amount,
        Currency:  req.Currency,
        Timestamp: time.Now(),
        Metadata:  req.Metadata,
    }
    m.txs[txID] = tx
    
    // Update balance
    if _, ok := m.balances[req.WalletID]; !ok {
        m.balances[req.WalletID] = make(map[types.CurrencyCode]*types.Amount)
    }
    
    if m.balances[req.WalletID][req.Currency] == nil {
        m.balances[req.WalletID][req.Currency] = &types.Amount{Value: nil, Decimal: 0}
    }
    
    // Add to balance
    currentBalance := m.balances[req.WalletID][req.Currency]
    newBalance, err := currentBalance.Add(&req.Amount)
    if err != nil {
        return nil, connector.NewError(connector.ErrInternal, "failed to add to balance", err)
    }
    m.balances[req.WalletID][req.Currency] = newBalance
    
    return &connector.IssueResponse{
        TransactionID: txID,
        Timestamp:     tx.Timestamp,
        Status:        types.StatusConfirmed,
        Fee:           types.Amount{Value: nil, Decimal: 0},
    }, nil
}

// Transfer implements the connector interface.
func (m *MockConnector) Transfer(ctx context.Context, req *connector.TransferRequest) (*connector.TransferResponse, error) {
    if err := m.checkOperation("transfer"); err != nil {
        return nil, err
    }
    
    if err := m.simulateOperation(ctx); err != nil {
        return nil, err
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Check source balance
    if m.balances[req.SourceWalletID] == nil ||
        m.balances[req.SourceWalletID][req.Currency] == nil {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "source wallet has insufficient balance", nil)
    }
    
    // Ensure destination wallet exists
    if m.balances[req.DestinationWalletID] == nil {
        m.balances[req.DestinationWalletID] = make(map[types.CurrencyCode]*types.Amount)
    }
    if m.balances[req.DestinationWalletID][req.Currency] == nil {
        m.balances[req.DestinationWalletID][req.Currency] = &types.Amount{Value: nil, Decimal: 0}
    }
    
    // Check if amount is available
    if m.balances[req.SourceWalletID][req.Currency].Cmp(&req.Amount) < 0 {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "insufficient balance", nil)
    }
    
    // Execute transfer
    sourceBalance := m.balances[req.SourceWalletID][req.Currency]
    destBalance := m.balances[req.DestinationWalletID][req.Currency]
    
    newSourceBalance, err := sourceBalance.Sub(&req.Amount)
    if err != nil {
        return nil, connector.NewError(connector.ErrInternal, "failed to update source balance", err)
    }
    
    newDestBalance, err := destBalance.Add(&req.Amount)
    if err != nil {
        return nil, connector.NewError(connector.ErrInternal, "failed to update destination balance", err)
    }
    
    m.balances[req.SourceWalletID][req.Currency] = newSourceBalance
    m.balances[req.DestinationWalletID][req.Currency] = newDestBalance
    
    // Create transaction record
    txID := uuid.New().String()
    tx := &mockTransaction{
        ID:        txID,
        Type:      types.TypeTransfer,
        Status:    types.StatusConfirmed,
        From:      req.SourceWalletID,
        To:        req.DestinationWalletID,
        Amount:    req.Amount,
        Currency:  req.Currency,
        Fee:       types.Amount{Value: nil, Decimal: 0},
        Timestamp: time.Now(),
        Metadata:  req.Metadata,
    }
    m.txs[txID] = tx
    
    return &connector.TransferResponse{
        TransactionID: txID,
        Timestamp:     tx.Timestamp,
        Status:        types.StatusConfirmed,
        Fee:           types.Amount{Value: nil, Decimal: 0},
    }, nil
}

// Lock implements the connector interface.
func (m *MockConnector) Lock(ctx context.Context, req *connector.LockRequest) (*connector.LockResponse, error) {
    if err := m.checkOperation("lock"); err != nil {
        return nil, err
    }
    
    if err := m.simulateOperation(ctx); err != nil {
        return nil, err
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Check balance
    if m.balances[req.WalletID] == nil ||
        m.balances[req.WalletID][req.Currency] == nil {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "wallet has insufficient balance", nil)
    }
    
    // Check if amount is available
    if m.balances[req.WalletID][req.Currency].Cmp(&req.Amount) < 0 {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "insufficient balance", nil)
    }
    
    // Create lock
    lockID := uuid.New().String()
    lock := &mockLock{
        ID:        lockID,
        WalletID:  req.WalletID,
        Amount:    req.Amount,
        Currency:  req.Currency,
        ExpiresAt: time.Now().Add(req.LockDuration),
        Status:    "locked",
    }
    m.locks[lockID] = lock
    
    // Create transaction record
    txID := uuid.New().String()
    tx := &mockTransaction{
        ID:        txID,
        Type:      types.TypeLock,
        Status:    types.StatusConfirmed,
        From:      req.WalletID,
        To:        req.WalletID,
        Amount:    req.Amount,
        Currency:  req.Currency,
        Timestamp: time.Now(),
        Metadata:  req.Metadata,
    }
    m.txs[txID] = tx
    
    return &connector.LockResponse{
        LockID:        lockID,
        TransactionID: txID,
        Status:        types.StatusConfirmed,
        ExpiresAt:     lock.ExpiresAt,
        Fee:           types.Amount{Value: nil, Decimal: 0},
    }, nil
}

// Burn implements the connector interface.
func (m *MockConnector) Burn(ctx context.Context, req *connector.BurnRequest) (*connector.BurnResponse, error) {
    if err := m.checkOperation("burn"); err != nil {
        return nil, err
    }
    
    if err := m.simulateOperation(ctx); err != nil {
        return nil, err
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Check balance
    if m.balances[req.WalletID] == nil ||
        m.balances[req.WalletID][req.Currency] == nil {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "wallet has insufficient balance", nil)
    }
    
    // Check if amount is available
    if m.balances[req.WalletID][req.Currency].Cmp(&req.Amount) < 0 {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "insufficient balance", nil)
    }
    
    // Update balance
    currentBalance := m.balances[req.WalletID][req.Currency]
    newBalance, err := currentBalance.Sub(&req.Amount)
    if err != nil {
        return nil, connector.NewError(connector.ErrInternal, "failed to update balance", err)
    }
    m.balances[req.WalletID][req.Currency] = newBalance
    
    // Create transaction record
    txID := uuid.New().String()
    tx := &mockTransaction{
        ID:        txID,
        Type:      types.TypeBurn,
        Status:    types.StatusConfirmed,
        From:      req.WalletID,
        Amount:    req.Amount,
        Currency:  req.Currency,
        Timestamp: time.Now(),
        Metadata:  req.Metadata,
    }
    m.txs[txID] = tx
    
    return &connector.BurnResponse{
        TransactionID: txID,
        Timestamp:     tx.Timestamp,
        Status:        types.StatusConfirmed,
        Fee:           types.Amount{Value: nil, Decimal: 0},
    }, nil
}

// Redeem implements the connector interface.
func (m *MockConnector) Redeem(ctx context.Context, req *connector.RedeemRequest) (*connector.RedeemResponse, error) {
    if err := m.checkOperation("redeem"); err != nil {
        return nil, err
    }
    
    if err := m.simulateOperation(ctx); err != nil {
        return nil, err
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Check balance
    if m.balances[req.WalletID] == nil ||
        m.balances[req.WalletID][req.Currency] == nil {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "wallet has insufficient balance", nil)
    }
    
    // Check if amount is available
    if m.balances[req.WalletID][req.Currency].Cmp(&req.Amount) < 0 {
        return nil, connector.NewError(connector.ErrInsufficientBalance, "insufficient balance", nil)
    }
    
    // Update balance
    currentBalance := m.balances[req.WalletID][req.Currency]
    newBalance, err := currentBalance.Sub(&req.Amount)
    if err != nil {
        return nil, connector.NewError(connector.ErrInternal, "failed to update balance", err)
    }
    m.balances[req.WalletID][req.Currency] = newBalance
    
    // Create transaction record
    txID := uuid.New().String()
    tx := &mockTransaction{
        ID:        txID,
        Type:      types.TypeRedeem,
        Status:    types.StatusConfirmed,
        From:      req.WalletID,
        Amount:    req.Amount,
        Currency:  req.Currency,
        Timestamp: time.Now(),
        Metadata:  req.Metadata,
    }
    m.txs[txID] = tx
    
    return &connector.RedeemResponse{
        TransactionID: txID,
        Timestamp:     tx.Timestamp,
        Status:        types.StatusConfirmed,
        Fee:           types.Amount{Value: nil, Decimal: 0},
    }, nil
}

// GetBalance implements the connector interface.
func (m *MockConnector) GetBalance(ctx context.Context, req *connector.BalanceRequest) (*connector.BalanceResponse, error) {
    if err := m.checkOperation("balance"); err != nil {
        return nil, err
    }
    
    if err := m.simulateOperation(ctx); err != nil {
        return nil, err
    }
    
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if m.balances[req.WalletID] == nil ||
        m.balances[req.WalletID][req.Currency] == nil {
        return &connector.BalanceResponse{
            WalletID: req.WalletID,
            Available: types.Amount{Value: nil, Decimal: 0},
            Total: types.Amount{Value: nil, Decimal: 0},
            Currency: req.Currency,
            UpdatedAt: time.Now(),
        }, nil
    }
    
    balance := m.balances[req.WalletID][req.Currency]
    return &connector.BalanceResponse{
        WalletID: req.WalletID,
        Available: *balance,
        Total: *balance,
        Currency: req.Currency,
        UpdatedAt: time.Now(),
    }, nil
}

// GetTransaction implements the connector interface.
func (m *MockConnector) GetTransaction(ctx context.Context, req *connector.TransactionQueryRequest) (*connector.TransactionResponse, error) {
    if err := m.checkOperation("get_transaction"); err != nil {
        return nil, err
    }
    
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    tx, ok := m.txs[req.TransactionID]
    if !ok {
        return nil, connector.NewError(connector.ErrTransactionNotFound, "transaction not found", nil)
    }
    
    return &connector.TransactionResponse{
        TransactionID: tx.ID,
        Type:          tx.Type,
        Status:        tx.Status,
        From:          tx.From,
        To:            tx.To,
        Amount:        tx.Amount,
        Currency:      tx.Currency,
        Fee:           tx.Fee,
        Timestamp:     tx.Timestamp,
        Metadata:      tx.Metadata,
    }, nil
}

// GetTransactionStatus implements the connector interface.
func (m *MockConnector) GetTransactionStatus(ctx context.Context, req *connector.TransactionStatusRequest) (*connector.TransactionStatusResponse, error) {
    if err := m.checkOperation("transaction_status"); err != nil {
        return nil, err
    }
    
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    tx, ok := m.txs[req.TransactionID]
    if !ok {
        return nil, connector.NewError(connector.ErrTransactionNotFound, "transaction not found", nil)
    }
    
    return &connector.TransactionStatusResponse{
        TransactionID: tx.ID,
        Status:        tx.Status,
        IsFinal:       true,
        UpdatedAt:     tx.Timestamp,
    }, nil
}

// checkOperation checks if an operation is allowed.
func (m *MockConnector) checkOperation(operation string) error {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if m.simulateError {
        return connector.NewError(m.errorCode, "simulated error for "+operation, nil)
    }
    
    if fail, ok := m.failOnOperation[operation]; ok && fail {
        return connector.NewError(connector.ErrInternal, "simulated failure for "+operation, nil)
    }
    
    if len(m.allowedOperations) > 0 {
        if allowed, ok := m.allowedOperations[operation]; !ok || !allowed {
            return connector.NewError(connector.ErrFeatureNotSupported, "operation not allowed: "+operation, nil)
        }
    }
    
    return nil
}

// simulateOperation simulates operation latency and errors.
func (m *MockConnector) simulateOperation(ctx context.Context) error {
    m.mu.RLock()
    latency := m.simulateLatency
    m.mu.RUnlock()
    
    if latency > 0 {
        select {
        case <-time.After(latency):
        case <-ctx.Done():
            return connector.NewError(connector.ErrTimeout, "operation timed out", ctx.Err())
        }
    }
    
    return nil
}

// SetSimulateError enables error simulation.
func (m *MockConnector) SetSimulateError(code connector.ErrorCode) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.simulateError = true
    m.errorCode = code
}

// ClearSimulateError disables error simulation.
func (m *MockConnector) ClearSimulateError() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.simulateError = false
}

// SetSimulateLatency sets the latency simulation.
func (m *MockConnector) SetSimulateLatency(latency time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.simulateLatency = latency
}

// SetFailOnOperation sets an operation to fail.
func (m *MockConnector) SetFailOnOperation(operation string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.failOnOperation[operation] = true
}

// ClearFailOnOperation clears the fail on operation setting.
func (m *MockConnector) ClearFailOnOperation(operation string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.failOnOperation, operation)
}

// SetAllowedOperation sets an allowed operation.
func (m *MockConnector) SetAllowedOperation(operation string, allowed bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if allowed {
        m.allowedOperations[operation] = true
    } else {
        delete(m.allowedOperations, operation)
    }
}

// SetBalance sets the balance for a wallet.
func (m *MockConnector) SetBalance(walletID string, currency types.CurrencyCode, amount *types.Amount) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.balances[walletID] == nil {
        m.balances[walletID] = make(map[types.CurrencyCode]*types.Amount)
    }
    m.balances[walletID][currency] = amount
}

// GetMockTransactions returns all mock transactions.
func (m *MockConnector) GetMockTransactions() []*connector.TransactionResponse {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    result := make([]*connector.TransactionResponse, 0, len(m.txs))
    for _, tx := range m.txs {
        result = append(result, &connector.TransactionResponse{
            TransactionID: tx.ID,
            Type:          tx.Type,
            Status:        tx.Status,
            From:          tx.From,
            To:            tx.To,
            Amount:        tx.Amount,
            Currency:      tx.Currency,
            Fee:           tx.Fee,
            Timestamp:     tx.Timestamp,
            Metadata:      tx.Metadata,
        })
    }
    return result
}