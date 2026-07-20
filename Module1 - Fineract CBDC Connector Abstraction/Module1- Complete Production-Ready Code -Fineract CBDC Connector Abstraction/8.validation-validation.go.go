package validation

import (
    "fmt"
    "reflect"
    "strings"
    
    "github.com/go-playground/validator/v10"
    "github.com/fineract/cbdc/connector-abstraction/types"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
    
    // Register custom validators
    validate.RegisterValidation("currency", validateCurrency)
    validate.RegisterValidation("amount", validateAmount)
    validate.RegisterValidation("transaction_type", validateTransactionType)
}

// Validate validates a struct using the validator tags.
func Validate(s interface{}) error {
    if err := validate.Struct(s); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            var errors []string
            for _, ve := range validationErrors {
                errors = append(errors, formatValidationError(ve))
            }
            return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
        }
        return err
    }
    return nil
}

// ValidateField validates a single field.
func ValidateField(field interface{}, tag string) error {
    return validate.Var(field, tag)
}

// formatValidationError formats a validation error.
func formatValidationError(ve validator.FieldError) string {
    switch ve.Tag() {
    case "required":
        return fmt.Sprintf("%s is required", ve.Field())
    case "gt":
        return fmt.Sprintf("%s must be greater than %s", ve.Field(), ve.Param())
    case "gte":
        return fmt.Sprintf("%s must be greater than or equal to %s", ve.Field(), ve.Param())
    case "lt":
        return fmt.Sprintf("%s must be less than %s", ve.Field(), ve.Param())
    case "lte":
        return fmt.Sprintf("%s must be less than or equal to %s", ve.Field(), ve.Param())
    case "len":
        return fmt.Sprintf("%s must be %s characters long", ve.Field(), ve.Param())
    case "min":
        return fmt.Sprintf("%s must be at least %s", ve.Field(), ve.Param())
    case "max":
        return fmt.Sprintf("%s must be at most %s", ve.Field(), ve.Param())
    case "currency":
        return fmt.Sprintf("%s must be a valid 3-letter currency code", ve.Field())
    case "amount":
        return fmt.Sprintf("%s must be a valid amount", ve.Field())
    case "transaction_type":
        return fmt.Sprintf("%s must be a valid transaction type", ve.Field())
    default:
        return fmt.Sprintf("%s failed validation for tag %s", ve.Field(), ve.Tag())
    }
}

// validateCurrency validates a currency code.
func validateCurrency(fl validator.FieldLevel) bool {
    currency, ok := fl.Field().Interface().(types.CurrencyCode)
    if !ok {
        return false
    }
    
    // Check if the currency code is 3 characters long
    if len(currency) != 3 {
        return false
    }
    
    // Add additional currency validation logic here
    // This could include checking against a list of supported currencies
    
    return true
}

// validateAmount validates an amount.
func validateAmount(fl validator.FieldLevel) bool {
    amount, ok := fl.Field().Interface().(types.Amount)
    if !ok {
        return false
    }
    
    // Check if amount is zero or negative
    if amount.IsZero() || amount.IsNegative() {
        return false
    }
    
    return true
}

// validateTransactionType validates a transaction type.
func validateTransactionType(fl validator.FieldLevel) bool {
    t, ok := fl.Field().Interface().(types.TransactionType)
    if !ok {
        return false
    }
    
    // Check if the transaction type is valid
    switch t {
    case types.TypeIssue, types.TypeTransfer, types.TypeLock, 
         types.TypeBurn, types.TypeRedeem, types.TypeUnlock:
        return true
    default:
        return false
    }
}

// ValidateTransferRequest validates a transfer request.
func ValidateTransferRequest(req *TransferRequest) error {
    if req == nil {
        return fmt.Errorf("transfer request cannot be nil")
    }
    
    if req.SourceWalletID == "" {
        return fmt.Errorf("source wallet ID is required")
    }
    
    if req.DestinationWalletID == "" {
        return fmt.Errorf("destination wallet ID is required")
    }
    
    if req.SourceWalletID == req.DestinationWalletID {
        return fmt.Errorf("source and destination wallet IDs cannot be the same")
    }
    
    if req.Amount.IsZero() || req.Amount.IsNegative() {
        return fmt.Errorf("amount must be positive")
    }
    
    return nil
}

// ValidateLockRequest validates a lock request.
func ValidateLockRequest(req *LockRequest) error {
    if req == nil {
        return fmt.Errorf("lock request cannot be nil")
    }
    
    if req.WalletID == "" {
        return fmt.Errorf("wallet ID is required")
    }
    
    if req.Amount.IsZero() || req.Amount.IsNegative() {
        return fmt.Errorf("amount must be positive")
    }
    
    if req.LockDuration <= 0 {
        return fmt.Errorf("lock duration must be positive")
    }
    
    return nil
}