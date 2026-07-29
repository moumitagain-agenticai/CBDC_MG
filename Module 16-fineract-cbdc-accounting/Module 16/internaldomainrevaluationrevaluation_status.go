package revaluation

// RevaluationStatus represents the status of a revaluation
type RevaluationStatus string

const (
    StatusPending   RevaluationStatus = "PENDING"
    StatusProcessing RevaluationStatus = "PROCESSING"
    StatusCompleted  RevaluationStatus = "COMPLETED"
    StatusFailed     RevaluationStatus = "FAILED"
)

// String returns the string representation
func (s RevaluationStatus) String() string {
    return string(s)
}

// GainLossType represents the type of gain/loss
type GainLossType string

const (
    GainType    GainLossType = "GAIN"
    LossType    GainLossType = "LOSS"
)

// String returns the string representation
func (g GainLossType) String() string {
    return string(g)
}