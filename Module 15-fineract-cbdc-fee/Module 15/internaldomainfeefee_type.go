package fee

// FeeType represents the type of fee
type FeeType string

const (
    TypeProcessing  FeeType = "PROCESSING"
    TypeTransaction FeeType = "TRANSACTION"
    TypeSettlement  FeeType = "SETTLEMENT"
    TypeConversion  FeeType = "CONVERSION"
    TypeCrossBorder FeeType = "CROSS_BORDER"
    TypeCompliance  FeeType = "COMPLIANCE"
    TypeService     FeeType = "SERVICE"
    TypeRegulatory  FeeType = "REGULATORY"
)

// String returns the string representation
func (t FeeType) String() string {
    return string(t)
}