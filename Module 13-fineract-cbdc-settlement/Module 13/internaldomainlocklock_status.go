package lock

// LockStatus represents the status of a fund lock
type LockStatus string

const (
    LockStatusActive    LockStatus = "ACTIVE"
    LockStatusReleased  LockStatus = "RELEASED"
    LockStatusExpired   LockStatus = "EXPIRED"
    LockStatusFailed    LockStatus = "FAILED"
)

// String returns the string representation
func (s LockStatus) String() string {
    return string(s)
}