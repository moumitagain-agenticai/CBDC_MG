package ledger

// EntryType represents the type of journal entry
type EntryType string

const (
    EntryTypeDebit  EntryType = "DEBIT"
    EntryTypeCredit EntryType = "CREDIT"
)

// String returns the string representation
func (t EntryType) String() string {
    return string(t)
}