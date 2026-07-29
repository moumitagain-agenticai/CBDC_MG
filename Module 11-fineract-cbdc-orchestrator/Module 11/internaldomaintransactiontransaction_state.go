package transaction

// TransactionState represents the state machine states
type TransactionState string

const (
    StateInitiated         TransactionState = "INITIATED"
    StateComplianceCheck   TransactionState = "COMPLIANCE_CHECK"
    StateFXConversion      TransactionState = "FX_CONVERSION"
    StateLockFunds         TransactionState = "LOCK_FUNDS"
    StateSettlement        TransactionState = "SETTLEMENT"
    StateReleaseLock       TransactionState = "RELEASE_LOCK"
    StateConfirmCompletion TransactionState = "CONFIRM_COMPLETION"
    StateCompleted         TransactionState = "COMPLETED"
    StateFailed            TransactionState = "FAILED"
)

// String returns the string representation
func (s TransactionState) String() string {
    return string(s)
}

// IsTerminal checks if the state is terminal
func (s TransactionState) IsTerminal() bool {
    return s == StateCompleted || s == StateFailed
}