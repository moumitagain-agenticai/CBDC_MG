package quote

// QuoteStatus represents the status of an FX quote
type QuoteStatus string

const (
    QuoteStatusActive    QuoteStatus = "ACTIVE"
    QuoteStatusLocked    QuoteStatus = "LOCKED"
    QuoteStatusUsed      QuoteStatus = "USED"
    QuoteStatusExpired   QuoteStatus = "EXPIRED"
    QuoteStatusCancelled QuoteStatus = "CANCELLED"
)

// String returns the string representation
func (s QuoteStatus) String() string {
    return string(s)
}

// IsTerminal checks if the status is terminal
func (s QuoteStatus) IsTerminal() bool {
    return s == QuoteStatusUsed || s == QuoteStatusExpired || s == QuoteStatusCancelled
}