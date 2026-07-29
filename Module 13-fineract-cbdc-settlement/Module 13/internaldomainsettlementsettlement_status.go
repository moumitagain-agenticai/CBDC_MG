package settlement

// SettlementStatus represents the status of a settlement
type SettlementStatus string

const (
    StatusPending     SettlementStatus = "PENDING"
    StatusLocking     SettlementStatus = "LOCKING"
    StatusBurning     SettlementStatus = "BURNING"
    StatusIssuing     SettlementStatus = "ISSUING"
    StatusReleasing   SettlementStatus = "RELEASING"
    StatusCompleted   SettlementStatus = "COMPLETED"
    StatusFailed      SettlementStatus = "FAILED"
    StatusRollingBack SettlementStatus = "ROLLING_BACK"
    StatusRolledBack  SettlementStatus = "ROLLED_BACK"
)

// String returns the string representation
func (s SettlementStatus) String() string {
    return string(s)
}

// IsTerminal checks if the status is terminal
func (s SettlementStatus) IsTerminal() bool {
    return s == StatusCompleted || s == StatusFailed || s == StatusRolledBack
}