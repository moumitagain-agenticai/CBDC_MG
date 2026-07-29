package screening

// ScreeningType represents the type of screening
type ScreeningType string

const (
    TypeSanctions   ScreeningType = "SANCTIONS"
    TypePEP         ScreeningType = "PEP"         // Politically Exposed Person
    TypeAML         ScreeningType = "AML"         // Anti-Money Laundering
    TypeCTF         ScreeningType = "CTF"         // Counter-Terrorism Financing
    TypeRisk        ScreeningType = "RISK"        // Risk Assessment
    TypeFraud       ScreeningType = "FRAUD"       // Fraud Detection
    TypeCompliance  ScreeningType = "COMPLIANCE"  // General Compliance
)

// String returns the string representation
func (t ScreeningType) String() string {
    return string(t)
}