// Check for connector errors
if err != nil {
    if ce, ok := connector.AsConnectorError(err); ok {
        // Handle specific error
        fmt.Printf("Error: [%s] %s\n", ce.Code, ce.Message)
        fmt.Printf("HTTP Status: %d\n", ce.HTTPStatusCode())
    }
}