package transaction

import (
    "fmt"
)

// DomainError represents a domain error
type DomainError struct {
    Code    string
    Message string
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
    return &DomainError{Code: code, Message: message}
}

// Predefined domain errors
var (
    ErrInvalidAmount       = NewDomainError("INVALID_AMOUNT", "amount must be greater than 0")
    ErrInvalidCurrency     = NewDomainError("INVALID_CURRENCY", "currency is required")
    ErrInvalidCountry      = NewDomainError("INVALID_COUNTRY", "country is required")
    ErrInvalidAccount      = NewDomainError("INVALID_ACCOUNT", "account is required")
    ErrTransactionNotFound = NewDomainError("TRANSACTION_NOT_FOUND", "transaction not found")
    ErrInvalidState        = NewDomainError("INVALID_STATE", "invalid transaction state")
)

// TransitionError is returned when an invalid transition is attempted
type TransitionError struct {
    FromState TransactionState
    ToState   TransactionState
}

func NewTransitionError(from, to TransactionState) *TransitionError {
    return &TransitionError{FromState: from, ToState: to}
}

func (e *TransitionError) Error() string {
    return fmt.Sprintf("invalid transition from %s to %s", e.FromState, e.ToState)
}

// InvalidStateError is returned when an invalid state is encountered
type InvalidStateError struct {
    State TransactionState
    Msg   string
}

func NewInvalidStateError(msg string, state TransactionState) *InvalidStateError {
    return &InvalidStateError{State: state, Msg: msg}
}

func (e *InvalidStateError) Error() string {
    return fmt.Sprintf("invalid state %s: %s", e.State, e.Msg)
}

// RetryExceededError is returned when max retries are exceeded
type RetryExceededError struct {
    TxID    string
    Attempts int
}

func NewRetryExceededError(txID string, attempts int) *RetryExceededError {
    return &RetryExceededError{TxID: txID, Attempts: attempts}
}

func (e *RetryExceededError) Error() string {
    return fmt.Sprintf("retry limit exceeded for transaction %s after %d attempts", e.TxID, e.Attempts)
}