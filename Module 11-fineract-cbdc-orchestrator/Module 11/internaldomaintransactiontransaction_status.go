package transaction

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
    StatusPending         TransactionStatus = "PENDING"
    StatusComplianceCheck TransactionStatus = "COMPLIANCE_CHECK"
    StatusFXProcessing    TransactionStatus = "FX_PROCESSING"
    StatusLocking         TransactionStatus = "LOCKING"
    StatusSettling        TransactionStatus = "SETTLING"
    StatusReleasingLock   TransactionStatus = "RELEASING_LOCK"
    StatusCompleted       TransactionStatus = "COMPLETED"
    StatusFailed          TransactionStatus = "FAILED"
    StatusCancelled       TransactionStatus = "CANCELLED"
)

// String returns the string representation
func (s TransactionStatus) String() string {
    return string(s)
}

// IsTerminal checks if the status is terminal
func (s TransactionStatus) IsTerminal() bool {
    return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// IsInProgress checks if the status is in progress
func (s TransactionStatus) IsInProgress() bool {
    return !s.IsTerminal()
}