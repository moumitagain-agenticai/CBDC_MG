package conversion

// ConversionStatus represents the status of a conversion
type ConversionStatus string

const (
    ConversionStatusPending   ConversionStatus = "PENDING"
    ConversionStatusProcessing ConversionStatus = "PROCESSING"
    ConversionStatusCompleted  ConversionStatus = "COMPLETED"
    ConversionStatusFailed     ConversionStatus = "FAILED"
    ConversionStatusRollbacked ConversionStatus = "ROLLBACKED"
)

// String returns the string representation
func (s ConversionStatus) String() string {
    return string(s)
}

// IsTerminal checks if the status is terminal
func (s ConversionStatus) IsTerminal() bool {
    return s == ConversionStatusCompleted || s == ConversionStatusFailed || s == ConversionStatusRollbacked
}