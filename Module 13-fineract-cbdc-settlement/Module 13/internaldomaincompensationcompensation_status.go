package compensation

// CompensationStatus represents the status of a compensation
type CompensationStatus string

const (
    CompStatusPending    CompensationStatus = "PENDING"
    CompStatusProcessing CompensationStatus = "PROCESSING"
    CompStatusCompleted  CompensationStatus = "COMPLETED"
    CompStatusFailed     CompensationStatus = "FAILED"
    CompStatusManual     CompensationStatus = "MANUAL_REQUIRED"
)

// String returns the string representation
func (s CompensationStatus) String() string {
    return string(s)
}