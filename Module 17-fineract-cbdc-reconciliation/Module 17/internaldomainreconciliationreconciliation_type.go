package reconciliation

// ReconciliationType represents the type of reconciliation
type ReconciliationType string

const (
    TypeNostro   ReconciliationType = "NOSTRO"
    TypeVostro   ReconciliationType = "VOSTRO"
    TypeInternal ReconciliationType = "INTERNAL"
    TypeInterBank ReconciliationType = "INTER_BANK"
    TypeClearing ReconciliationType = "CLEARING"
)

// String returns the string representation
func (t ReconciliationType) String() string {
    return string(t)
}