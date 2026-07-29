package statement

// StatementStatus represents the status of a bank statement
type StatementStatus string

const (
    StatusReceived  StatementStatus = "RECEIVED"
    StatusProcessed StatementStatus = "PROCESSED"
    StatusMatched   StatementStatus = "MATCHED"
    StatusFailed    StatementStatus = "FAILED"
)

// String returns the string representation
func (s StatementStatus) String() string {
    return string(s)
}