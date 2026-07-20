package connector

import (
    "fmt"
    "net/http"
)

// ConnectorError represents a connector-specific error.
type ConnectorError struct {
    // Code is the error code.
    Code ErrorCode `json:"code"`
    
    // Message is the human-readable error message.
    Message string `json:"message"`
    
    // Details contains additional error details.
    Details map[string]interface{} `json:"details,omitempty"`
    
    // Inner is the underlying error (if any).
    Inner error `json:"-"`
}

// ErrorCode represents a connector error code.
type ErrorCode string

const (
    // Common errors
    ErrInternal            ErrorCode = "internal_error"
    ErrTimeout             ErrorCode = "timeout"
    ErrNetwork             ErrorCode = "network_error"
    ErrValidation          ErrorCode = "validation_error"
    ErrUnauthorized        ErrorCode = "unauthorized"
    ErrForbidden           ErrorCode = "forbidden"
    ErrNotFound            ErrorCode = "not_found"
    ErrConflict            ErrorCode = "conflict"
    ErrRateLimit           ErrorCode = "rate_limit_exceeded"
    ErrServiceUnavailable  ErrorCode = "service_unavailable"
    
    // CBDC-specific errors
    ErrInsufficientBalance  ErrorCode = "insufficient_balance"
    ErrInvalidWallet        ErrorCode = "invalid_wallet"
    ErrInvalidAmount        ErrorCode = "invalid_amount"
    ErrCurrencyNotSupported ErrorCode = "currency_not_supported"
    ErrLockNotFound         ErrorCode = "lock_not_found"
    ErrLockExpired          ErrorCode = "lock_expired"
    ErrLockConflict         ErrorCode = "lock_conflict"
    ErrTransactionNotFound  ErrorCode = "transaction_not_found"
    ErrTransactionFailed    ErrorCode = "transaction_failed"
    ErrTransactionPending   ErrorCode = "transaction_pending"
    ErrFeatureNotSupported  ErrorCode = "feature_not_supported"
    ErrBlockchainUnreachable ErrorCode = "blockchain_unreachable"
    ErrInvalidSignature     ErrorCode = "invalid_signature"
    ErrDuplicateTransaction ErrorCode = "duplicate_transaction"
    ErrComplianceCheckFailed ErrorCode = "compliance_check_failed"
    ErrLimitExceeded        ErrorCode = "limit_exceeded"
)

// Error implements the error interface.
func (e *ConnectorError) Error() string {
    if e.Inner != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Inner)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *ConnectorError) Unwrap() error {
    return e.Inner
}

// HTTPStatusCode returns the HTTP status code for the error.
func (e *ConnectorError) HTTPStatusCode() int {
    switch e.Code {
    case ErrInternal, ErrBlockchainUnreachable:
        return http.StatusInternalServerError
    case ErrTimeout:
        return http.StatusRequestTimeout
    case ErrValidation, ErrInvalidAmount, ErrInvalidWallet:
        return http.StatusBadRequest
    case ErrUnauthorized:
        return http.StatusUnauthorized
    case ErrForbidden, ErrComplianceCheckFailed:
        return http.StatusForbidden
    case ErrNotFound, ErrLockNotFound, ErrTransactionNotFound:
        return http.StatusNotFound
    case ErrConflict, ErrLockConflict, ErrDuplicateTransaction:
        return http.StatusConflict
    case ErrRateLimit:
        return http.StatusTooManyRequests
    case ErrServiceUnavailable:
        return http.StatusServiceUnavailable
    case ErrInsufficientBalance, ErrLimitExceeded:
        return http.StatusPaymentRequired
    default:
        return http.StatusInternalServerError
    }
}

// NewError creates a new ConnectorError.
func NewError(code ErrorCode, message string, inner error) *ConnectorError {
    return &ConnectorError{
        Code:    code,
        Message: message,
        Inner:   inner,
        Details: make(map[string]interface{}),
    }
}

// NewErrorWithDetails creates a new ConnectorError with details.
func NewErrorWithDetails(code ErrorCode, message string, details map[string]interface{}, inner error) *ConnectorError {
    return &ConnectorError{
        Code:    code,
        Message: message,
        Details: details,
        Inner:   inner,
    }
}

// IsConnectorError checks if an error is a ConnectorError.
func IsConnectorError(err error) bool {
    _, ok := err.(*ConnectorError)
    return ok
}

// AsConnectorError converts an error to a ConnectorError.
func AsConnectorError(err error) (*ConnectorError, bool) {
    if err == nil {
        return nil, false
    }
    var ce *ConnectorError
    if ok := errors.As(err, &ce); ok {
        return ce, true
    }
    return nil, false
}