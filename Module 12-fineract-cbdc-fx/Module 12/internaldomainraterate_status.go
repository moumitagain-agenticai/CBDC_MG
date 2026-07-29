package rate

// RateStatus represents the status of an exchange rate
type RateStatus string

const (
    RateStatusActive    RateStatus = "ACTIVE"
    RateStatusExpired   RateStatus = "EXPIRED"
    RateStatusStale     RateStatus = "STALE"
    RateStatusFailed    RateStatus = "FAILED"
)

// String returns the string representation
func (s RateStatus) String() string {
    return string(s)
}

// IsValid checks if the status is valid
func (s RateStatus) IsValid() bool {
    switch s {
    case RateStatusActive, RateStatusExpired, RateStatusStale, RateStatusFailed:
        return true
    default:
        return false
    }
}