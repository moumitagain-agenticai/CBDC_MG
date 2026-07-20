// Validate a request
if err := validation.Validate(transferRequest); err != nil {
    return nil, err
}

// Or use specific validators
if err := validation.ValidateTransferRequest(transferRequest); err != nil {
    return nil, err
}