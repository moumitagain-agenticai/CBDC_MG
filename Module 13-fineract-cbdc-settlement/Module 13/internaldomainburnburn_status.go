package burn

// BurnStatus represents the status of a burn operation
type BurnStatus string

const (
    BurnStatusPending   BurnStatus = "PENDING"
    BurnStatusConfirmed BurnStatus = "CONFIRMED"
    BurnStatusFailed    BurnStatus = "FAILED"
    BurnStatusReversed  BurnStatus = "REVERSED"
)

// String returns the string representation
func (s BurnStatus) String() string {
    return string(s)
}