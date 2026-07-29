package compliance

// ComplianceStatus represents the status of a compliance check
type ComplianceStatus string

const (
    StatusPending   ComplianceStatus = "PENDING"
    StatusPassed    ComplianceStatus = "PASSED"
    StatusFailed    ComplianceStatus = "FAILED"
    StatusError     ComplianceStatus = "ERROR"
    StatusManual    ComplianceStatus = "MANUAL_REQUIRED"
)

// String returns the string representation
func (s ComplianceStatus) String() string {
    return string(s)
}