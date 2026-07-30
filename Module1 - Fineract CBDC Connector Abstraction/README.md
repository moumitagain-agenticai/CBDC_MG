# Fineract CBDC Connector Abstraction

This module defines the standard interface for all CBDC (Central Bank Digital Currency) connectors in the E1 Cross-Border CBDC Payment Platform.

## Overview

The connector abstraction module provides:

- **Standard Interface**: Defines the contract for all CBDC connectors
- **Data Models**: Common data structures for CBDC operations
- **Error Handling**: Standardized error codes and handling
- **Metrics**: Observability and monitoring
- **Tracing**: Distributed tracing support
- **Validation**: Input validation for all operations
- **Testing**: Mock implementations for testing

## Layout

```
go.mod
connector.go          package connector — CBDCConnector interface, request/response types
models.go             package connector — metadata, capabilities, events
errors.go             package connector — ConnectorError and error codes
types/                shared value types (Amount, CurrencyCode, statuses)
config/               connector configuration structs
reqctx/               request-scoped context keys and accessors
validation/           request validators
metrics/              Prometheus collectors
tracing/              OpenTelemetry span helpers
test/mock/            in-memory MockConnector for tests
examples/             runnable usage example
testdata/config.yaml  sample connector configuration
```

## Key Components

### CBDCConnector Interface

The main interface that all connectors must implement:

```go
type CBDCConnector interface {
    // Core operations
    Issue(ctx context.Context, req *IssueRequest) (*IssueResponse, error)
    Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)
    Lock(ctx context.Context, req *LockRequest) (*LockResponse, error)
    Burn(ctx context.Context, req *BurnRequest) (*BurnResponse, error)
    Redeem(ctx context.Context, req *RedeemRequest) (*RedeemResponse, error)

    // Query operations
    GetBalance(ctx context.Context, req *BalanceRequest) (*BalanceResponse, error)
    GetTransaction(ctx context.Context, req *TransactionQueryRequest) (*TransactionResponse, error)
    GetTransactionStatus(ctx context.Context, req *TransactionStatusRequest) (*TransactionStatusResponse, error)

    // Health & monitoring
    HealthCheck(ctx context.Context) (*HealthResponse, error)
    GetNetworkInfo(ctx context.Context) (*NetworkInfoResponse, error)

    // Lifecycle
    Initialize(ctx context.Context, cfg *config.ConnectorConfig) error
    Shutdown(ctx context.Context) error
    Status() ConnectorStatus
}
```

## Usage

### 1. Implementing a Connector

```go
type MyCBDCConnector struct {
    // Connector-specific fields
}

func (c *MyCBDCConnector) Initialize(ctx context.Context, cfg *config.ConnectorConfig) error {
    // Initialize connection to CBDC network
    return nil
}

func (c *MyCBDCConnector) Issue(ctx context.Context, req *connector.IssueRequest) (*connector.IssueResponse, error) {
    // Implementation specific to the CBDC network
    // Validate request
    if err := validation.Validate(req); err != nil {
        return nil, err
    }

    // Perform operation
    // ...

    return &connector.IssueResponse{
        TransactionID: "tx-123",
        Status:        types.StatusConfirmed,
    }, nil
}

// Implement all other required methods...
```

### 2. Using a Connector

```go
// Create configuration
cfg := &config.ConnectorConfig{
    ID:   "my-connector",
    Name: "My CBDC Connector",
    Type: "my-type",
    NetworkConfig: config.NetworkConfig{
        Endpoint: "https://api.my-cbdc.com",
    },
}

// Create and initialize connector
conn := NewMyCBDCConnector()
if err := conn.Initialize(context.Background(), cfg); err != nil {
    log.Fatal(err)
}

// Perform operations
transfer, err := conn.Transfer(ctx, &connector.TransferRequest{
    SourceWalletID:      "wallet-123",
    DestinationWalletID: "wallet-456",
    Amount:              amount,
    Currency:            "INR",
})
```

Note: do not name the local variable `connector` — it shadows the imported
package of the same name and makes `connector.TransferRequest` fail to compile.

### 3. Testing with the Mock Connector

```go
// Create mock connector for testing
mockConnector := mock.New(cfg)
mockConnector.SetBalance("wallet-123", "INR", amount)

// Use in tests
result, err := mockConnector.Transfer(ctx, &connector.TransferRequest{ /* ... */ })
```

## Error Handling

```go
// Check for connector errors
if err != nil {
    if ce, ok := connector.AsConnectorError(err); ok {
        // Handle specific error
        fmt.Printf("Error: [%s] %s\n", ce.Code, ce.Message)
        fmt.Printf("HTTP Status: %d\n", ce.HTTPStatusCode())
    }
}
```

## Validation

```go
// Validate a request
if err := validation.Validate(transferRequest); err != nil {
    return nil, err
}

// Or use specific validators
if err := validation.ValidateTransferRequest(transferRequest); err != nil {
    return nil, err
}
```

## Testing

```go
// Create mock connector
mockConnector := mock.New(cfg)

// Simulate errors
mockConnector.SetSimulateError(connector.ErrInternal)

// Simulate latency
mockConnector.SetSimulateLatency(100 * time.Millisecond)

// Set specific operation to fail
mockConnector.SetFailOnOperation("transfer")
```

## Integration with Fineract Core

When integrating this module into the Fineract core banking software:

1. **Add as a dependency**: Include this module in your Fineract project's `go.mod`
2. **Implement connectors**: Create specific connector implementations for each CBDC network (India, UAE, Gift City)
3. **Use in orchestrator**: The orchestrator module should use the `CBDCConnector` interface to interact with various CBDC networks
4. **Configuration**: Provide configuration for each connector in the Fineract configuration
5. **Health checks**: The monitoring system can use `HealthCheck()` to monitor connector health
6. **Error handling**: Standardized errors ensure consistent error handling across the system

This abstraction layer ensures that the core Fineract system can work with multiple CBDC networks without being tightly coupled to any specific implementation.

## Building

```
go mod tidy
go build ./...
go test ./...
```
