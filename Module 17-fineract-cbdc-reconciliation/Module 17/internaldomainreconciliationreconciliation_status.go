package reconciliation

// ReconciliationStatus represents the status of a reconciliation
type ReconciliationStatus string

const (
    StatusPending   ReconciliationStatus = "PENDING"
    StatusProcessing ReconciliationStatus = "PROCESSING"
    StatusCompleted  ReconciliationStatus = "COMPLETED"
    StatusFailed     ReconciliationStatus = "FAILED"
    StatusCancelled  ReconciliationStatus = "CANCELLED"
    StatusReview     ReconciliationStatus = "REVIEW_REQUIRED"
)

// String returns the string representation
func (s ReconciliationStatus) String() string {
    return string(s)
}