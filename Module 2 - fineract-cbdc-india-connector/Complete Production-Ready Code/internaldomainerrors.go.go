package domain

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorCode represents a domain error code
type ErrorCode string

const (
    ErrorInternal          ErrorCode = "internal_error"
    ErrorValidation        ErrorCode = "validation_error"
    ErrorNotFound          ErrorCode = "not_found"
    ErrorConflict          ErrorCode = "conflict"
    ErrorUnauthorized      ErrorCode = "unauthorized"
    ErrorInsufficientBalance ErrorCode = "insufficient_balance"
    ErrorInvalidWallet     ErrorCode = "invalid_wallet"
    ErrorTransactionFailed ErrorCode = "transaction_failed"
    ErrorTransactionTimeout ErrorCode = "transaction_timeout"
    ErrorCircuitOpen       ErrorCode = "circuit_open"
    ErrorRateLimit         ErrorCode = "rate_limit"
    ErrorRetryExhausted    ErrorCode = "retry_exhausted"
)

// Error implements the error interface
func (e *DomainError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string) *DomainError {
    return &DomainError{
        Code:    code,
        Message: message,
        Details: make(map[string]interface{}),
    }
}

// NewDomainErrorWithDetails creates a new domain error with details
func NewDomainErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *DomainError {
    return &DomainError{
        Code:    code,
        Message: message,
        Details: details,
    }
}