// Create mock connector
mockConnector := mock.NewMockConnector(cfg)

// Simulate errors
mockConnector.SetSimulateError(connector.ErrInternal)

// Simulate latency
mockConnector.SetSimulateLatency(100 * time.Millisecond)

// Set specific operation to fail
mockConnector.SetFailOnOperation("transfer")