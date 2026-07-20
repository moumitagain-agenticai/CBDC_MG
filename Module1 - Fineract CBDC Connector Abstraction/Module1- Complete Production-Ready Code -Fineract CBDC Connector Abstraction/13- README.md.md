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