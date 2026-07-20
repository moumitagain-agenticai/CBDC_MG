// Create mock connector for testing
mockConnector := mock.NewMockConnector(cfg)
mockConnector.SetBalance("wallet-123", "INR", amount)

// Use in tests
result, err := mockConnector.Transfer(ctx, &connector.TransferRequest{...})