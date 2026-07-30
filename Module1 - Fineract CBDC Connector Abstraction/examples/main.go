// Command examples demonstrates use of the CBDC connector abstraction
// against the in-memory mock connector.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	connector "github.com/fineract/cbdc/connector-abstraction"
	"github.com/fineract/cbdc/connector-abstraction/config"
	"github.com/fineract/cbdc/connector-abstraction/test/mock"
	"github.com/fineract/cbdc/connector-abstraction/types"
)

func main() {
	ExampleMockConnectorUsage()
	ExampleConnectorErrorHandling()
	ExampleConnectorTimeout()
	ExampleConnectorHealthCheck()
	ExampleConnectorBatchOperations()
}

// ExampleMockConnectorUsage demonstrates how to use the mock connector.
func ExampleMockConnectorUsage() {
	// Create configuration
	cfg := &config.ConnectorConfig{
		ID:   "mock-connector",
		Name: "Mock CBDC Connector",
		Type: "mock",
		NetworkConfig: config.NetworkConfig{
			Endpoint: "http://mock.local:8080",
		},
	}

	// Create connector
	conn := mock.New(cfg)

	// Initialize connector
	ctx := context.Background()
	if err := conn.Initialize(ctx, cfg); err != nil {
		log.Fatalf("Failed to initialize connector: %v", err)
	}

	// Check status
	fmt.Printf("Connector status: %s\n", conn.Status())

	// Set initial balance
	walletID := "wallet-123"
	initialBalance, _ := types.NewAmount("1000.00", 2)
	conn.SetBalance(walletID, "INR", initialBalance)

	// Check balance
	balanceResp, err := conn.GetBalance(ctx, &connector.BalanceRequest{
		WalletID: walletID,
		Currency: "INR",
	})
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Balance: %s\n", balanceResp.Available.String())

	// Transfer funds
	amount, _ := types.NewAmount("250.50", 2)
	transferResp, err := conn.Transfer(ctx, &connector.TransferRequest{
		SourceWalletID:      walletID,
		DestinationWalletID: "wallet-456",
		Amount:              *amount,
		Currency:            "INR",
		ReferenceID:         "ref-001",
		Metadata: map[string]interface{}{
			"note": "Test transfer",
		},
	})
	if err != nil {
		log.Fatalf("Failed to transfer: %v", err)
	}
	fmt.Printf("Transfer completed: %s\n", transferResp.TransactionID)

	// Check updated balance
	balanceResp, err = conn.GetBalance(ctx, &connector.BalanceRequest{
		WalletID: walletID,
		Currency: "INR",
	})
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Updated balance: %s\n", balanceResp.Available.String())
}

// ExampleConnectorErrorHandling demonstrates error handling.
func ExampleConnectorErrorHandling() {
	// Create configuration
	cfg := &config.ConnectorConfig{
		ID:   "mock-connector",
		Name: "Mock CBDC Connector",
		Type: "mock",
		NetworkConfig: config.NetworkConfig{
			Endpoint: "http://mock.local:8080",
		},
	}

	// Create connector
	conn := mock.New(cfg)
	ctx := context.Background()

	// Initialize connector
	if err := conn.Initialize(ctx, cfg); err != nil {
		log.Fatalf("Failed to initialize connector: %v", err)
	}

	// Set up error simulation
	conn.SetSimulateError(connector.ErrInternal)

	// Try to perform an operation
	amount, _ := types.NewAmount("100.00", 2)
	_, err := conn.Transfer(ctx, &connector.TransferRequest{
		SourceWalletID:      "wallet-123",
		DestinationWalletID: "wallet-456",
		Amount:              *amount,
		Currency:            "INR",
	})

	// Handle error
	if err != nil {
		if ce, ok := connector.AsConnectorError(err); ok {
			fmt.Printf("Connector error: [%s] %s\n", ce.Code, ce.Message)
			fmt.Printf("HTTP Status: %d\n", ce.HTTPStatusCode())
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}
}

// ExampleConnectorTimeout demonstrates timeout handling.
func ExampleConnectorTimeout() {
	// Create configuration
	cfg := &config.ConnectorConfig{
		ID:   "mock-connector",
		Name: "Mock CBDC Connector",
		Type: "mock",
		NetworkConfig: config.NetworkConfig{
			Endpoint: "http://mock.local:8080",
		},
		TimeoutConfig: config.TimeoutConfig{
			Request: 5 * time.Second,
		},
	}

	// Create connector
	conn := mock.New(cfg)
	ctx := context.Background()

	// Initialize connector
	if err := conn.Initialize(ctx, cfg); err != nil {
		log.Fatalf("Failed to initialize connector: %v", err)
	}

	// Simulate latency
	conn.SetSimulateLatency(10 * time.Second)

	// Set up timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Try to perform operation with timeout
	amount, _ := types.NewAmount("100.00", 2)
	_, err := conn.Transfer(ctxWithTimeout, &connector.TransferRequest{
		SourceWalletID:      "wallet-123",
		DestinationWalletID: "wallet-456",
		Amount:              *amount,
		Currency:            "INR",
	})

	if err != nil {
		if ce, ok := connector.AsConnectorError(err); ok {
			fmt.Printf("Timeout error: [%s] %s\n", ce.Code, ce.Message)
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}
}

// ExampleConnectorHealthCheck demonstrates health checking.
func ExampleConnectorHealthCheck() {
	// Create configuration
	cfg := &config.ConnectorConfig{
		ID:   "mock-connector",
		Name: "Mock CBDC Connector",
		Type: "mock",
		NetworkConfig: config.NetworkConfig{
			Endpoint: "http://mock.local:8080",
		},
	}

	// Create connector
	conn := mock.New(cfg)
	ctx := context.Background()

	// Initialize connector
	if err := conn.Initialize(ctx, cfg); err != nil {
		log.Fatalf("Failed to initialize connector: %v", err)
	}

	// Perform health check
	health, err := conn.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	fmt.Printf("Health status: %s\n", health.Status)
	for name, component := range health.Components {
		fmt.Printf("Component %s: %s\n", name, component.Status)
	}
}

// ExampleConnectorBatchOperations demonstrates batch operations.
func ExampleConnectorBatchOperations() {
	// Note: This example would use a connector that supports batch operations.
	// The mock connector doesn't support batch operations, so this is just illustrative.
	//
	// In a real implementation, you would:
	// 1. Get a connector that supports batch operations
	// 2. Prepare a batch request
	// 3. Execute the batch
	// 4. Handle results

	fmt.Println("Batch operations would be performed here")
}
