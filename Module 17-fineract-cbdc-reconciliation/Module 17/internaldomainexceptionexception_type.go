package exception

// ExceptionType represents the type of exception
type ExceptionType string

const (
    TypeUnmatched       ExceptionType = "UNMATCHED"
    TypePartialMatch    ExceptionType = "PARTIAL_MATCH"
    TypeAmountMismatch  ExceptionType = "AMOUNT_MISMATCH"
    TypeDateMismatch    ExceptionType = "DATE_MISMATCH"
    TypeCurrencyMismatch ExceptionType = "CURRENCY_MISMATCH"
    TypeMissingEntry    ExceptionType = "MISSING_ENTRY"
    TypeDuplicateEntry  ExceptionType = "DUPLICATE_ENTRY"
    TypeSuspicious      ExceptionType = "SUSPICIOUS"
    TypeCutOff          ExceptionType = "CUT_OFF"
    TypeBankError       ExceptionType = "BANK_ERROR"
    TypeSystemError     ExceptionType = "SYSTEM_ERROR"
)

// String returns the string representation
func (t ExceptionType) String() string {
    return string(t)
}