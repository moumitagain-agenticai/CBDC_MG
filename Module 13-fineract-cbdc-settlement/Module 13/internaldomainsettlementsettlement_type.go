package settlement

// SettlementType represents the type of settlement
type SettlementType string

const (
    TypePvP       SettlementType = "PVP"        // Payment vs Payment
    TypeDvP       SettlementType = "DVP"        // Delivery vs Payment
    TypeNet       SettlementType = "NET"        // Net settlement
    TypeGross     SettlementType = "GROSS"      // Gross settlement
    TypeAtomic    SettlementType = "ATOMIC"     // Atomic settlement
)

// String returns the string representation
func (t SettlementType) String() string {
    return string(t)
}