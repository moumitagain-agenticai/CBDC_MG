package exception

// ExceptionStatus represents the status of an exception
type ExceptionStatus string

const (
    StatusOpen      ExceptionStatus = "OPEN"
    StatusInProgress ExceptionStatus = "IN_PROGRESS"
    StatusResolved  ExceptionStatus = "RESOLVED"
    StatusClosed    ExceptionStatus = "CLOSED"
    StatusEscalated ExceptionStatus = "ESCALATED"
)

// String returns the string representation
func (s ExceptionStatus) String() string {
    return string(s)
}