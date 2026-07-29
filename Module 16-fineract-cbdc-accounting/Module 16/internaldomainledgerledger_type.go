package ledger

// LedgerType represents the type of ledger account
type LedgerType string

const (
    TypeAsset     LedgerType = "ASSET"
    TypeLiability LedgerType = "LIABILITY"
    TypeEquity    LedgerType = "EQUITY"
    TypeRevenue   LedgerType = "REVENUE"
    TypeExpense   LedgerType = "EXPENSE"
    TypeSuspense  LedgerType = "SUSPENSE"
    TypeContra    LedgerType = "CONTRA"
)

// String returns the string representation
func (t LedgerType) String() string {
    return string(t)
}