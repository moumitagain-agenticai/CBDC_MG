package screening

// ScreeningStatus represents the status of a screening
type ScreeningStatus string

const (
    StatusPending   ScreeningStatus = "PENDING"
    StatusInProgress ScreeningStatus = "IN_PROGRESS"
    StatusCompleted  ScreeningStatus = "COMPLETED"
    StatusBlocked    ScreeningStatus = "BLOCKED"
    StatusFailed     ScreeningStatus = "FAILED"
    StatusEscalated  ScreeningStatus = "ESCALATED"
)

// String returns the string representation
func (s ScreeningStatus) String() string {
    return string(s)
}

// IsTerminal checks if the status is terminal
func (s ScreeningStatus) IsTerminal() bool {
    return s == StatusCompleted || s == StatusBlocked || s == StatusFailed
}