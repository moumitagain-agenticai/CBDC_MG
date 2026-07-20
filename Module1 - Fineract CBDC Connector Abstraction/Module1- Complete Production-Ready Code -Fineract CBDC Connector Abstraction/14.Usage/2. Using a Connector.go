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
connector := NewMyCBDCConnector()
if err := connector.Initialize(context.Background(), cfg); err != nil {
    log.Fatal(err)
}

// Perform operations
transfer, err := connector.Transfer(ctx, &connector.TransferRequest{
    SourceWalletID:      "wallet-123",
    DestinationWalletID: "wallet-456",
    Amount:              amount,
    Currency:            "INR",
})