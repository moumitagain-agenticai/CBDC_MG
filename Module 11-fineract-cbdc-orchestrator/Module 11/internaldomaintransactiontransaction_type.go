package transaction

// TransactionType represents the type of transaction
type TransactionType string

const (
    TypeCrossBorder TransactionType = "CROSS_BORDER"
    TypeDomestic    TransactionType = "DOMESTIC"
    TypeSettlement  TransactionType = "SETTLEMENT"
)

// String returns the string representation
func (t TransactionType) String() string {
    return string(t)
}