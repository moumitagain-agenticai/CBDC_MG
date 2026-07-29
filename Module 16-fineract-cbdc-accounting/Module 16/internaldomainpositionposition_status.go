package position

// PositionStatus represents the status of a currency position
type PositionStatus string

const (
    StatusOpen     PositionStatus = "OPEN"
    StatusClosed   PositionStatus = "CLOSED"
    StatusRevaluing PositionStatus = "REVALUING"
    StatusRevalued PositionStatus = "REVALUED"
)

// String returns the string representation
func (s PositionStatus) String() string {
    return string(s)
}