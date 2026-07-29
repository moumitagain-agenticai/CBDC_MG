package fee

import (
    "context"
    "time"

    "github.com/shopspring/decimal"
)

// Service defines the fee service interface
type Service interface {
    // Fee Configuration
    CreateFee(ctx context.Context, req *CreateFeeRequest) (*Fee, error)
    GetFee(ctx context.Context, id string) (*Fee, error)
    GetFeeByCode(ctx context.Context, code string) (*Fee, error)
    ListFees(ctx context.Context, filter *FeeFilter) ([]*Fee, int64, error)
    UpdateFee(ctx context.Context, id string, req *UpdateFeeRequest) (*Fee, error)
    DeleteFee(ctx context.Context, id string) error
    ActivateFee(ctx context.Context, id string) error
    DeactivateFee(ctx context.Context, id string) error

    // Fee Corridor
    CreateCorridor(ctx context.Context, req *CreateCorridorRequest) (*corridor.FeeCorridor, error)
    GetCorridor(ctx context.Context, id string) (*corridor.FeeCorridor, error)
    GetCorridorByCode(ctx context.Context, code string) (*corridor.FeeCorridor, error)
    ListCorridors(ctx context.Context, filter *CorridorFilter) ([]*corridor.FeeCorridor, int64, error)
    UpdateCorridor(ctx context.Context, id string, req *UpdateCorridorRequest) (*corridor.FeeCorridor, error)
    DeleteCorridor(ctx context.Context, id string) error

    // Fee Calculation
    CalculateFee(ctx context.Context, req *CalculationRequest) (*CalculationResult, error)
    BulkCalculateFee(ctx context.Context, reqs []*CalculationRequest) ([]*CalculationResult, error)
    GetCalculation(ctx context.Context, id string) (*calculation.Calculation, error)
    ListCalculations(ctx context.Context, filter *CalculationFilter) ([]*calculation.Calculation, int64, error)

    // Health & Monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*ServiceMetrics, error)
}

// CreateFeeRequest represents a request to create a fee
type CreateFeeRequest struct {
    Name            string                 `json:"name"`
    Code            string                 `json:"code"`
    Type            FeeType                `json:"type"`
    Structure       FeeStructure           `json:"structure"`
    Value           decimal.Decimal        `json:"value"`
    MinAmount       decimal.Decimal        `json:"minAmount"`
    MaxAmount       decimal.Decimal        `json:"maxAmount"`
    TieredStructure []Tier                 `json:"tieredStructure"`
    CorridorID      string                 `json:"corridorId"`
    SourceCountry   string                 `json:"sourceCountry"`
    TargetCountry   string                 `json:"targetCountry"`
    SourceCurrency  string                 `json:"sourceCurrency"`
    TargetCurrency  string                 `json:"targetCurrency"`
    Priority        int                    `json:"priority"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// UpdateFeeRequest represents a request to update a fee
type UpdateFeeRequest struct {
    Name            *string                `json:"name"`
    Value           *decimal.Decimal       `json:"value"`
    MinAmount       *decimal.Decimal       `json:"minAmount"`
    MaxAmount       *decimal.Decimal       `json:"maxAmount"`
    TieredStructure []Tier                 `json:"tieredStructure"`
    Priority        *int                   `json:"priority"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// FeeFilter represents filters for listing fees
type FeeFilter struct {
    Type           FeeType    `json:"type"`
    Structure      FeeStructure `json:"structure"`
    CorridorID     string     `json:"corridorId"`
    SourceCountry  string     `json:"sourceCountry"`
    TargetCountry  string     `json:"targetCountry"`
    SourceCurrency string     `json:"sourceCurrency"`
    TargetCurrency string     `json:"targetCurrency"`
    IsActive       *bool      `json:"isActive"`
    Limit          int        `json:"limit"`
    Offset         int        `json:"offset"`
}

// CreateCorridorRequest represents a request to create a fee corridor
type CreateCorridorRequest struct {
    Name            string                 `json:"name"`
    Code            string                 `json:"code"`
    SourceCountry   string                 `json:"sourceCountry"`
    TargetCountry   string                 `json:"targetCountry"`
    SourceCurrency  string                 `json:"sourceCurrency"`
    TargetCurrency  string                 `json:"targetCurrency"`
    BaseFee         decimal.Decimal        `json:"baseFee"`
    Markup          decimal.Decimal        `json:"markup"`
    Discount        decimal.Decimal        `json:"discount"`
    MinFee          decimal.Decimal        `json:"minFee"`
    MaxFee          decimal.Decimal        `json:"maxFee"`
    Priority        int                    `json:"priority"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// UpdateCorridorRequest represents a request to update a fee corridor
type UpdateCorridorRequest struct {
    Name      *string          `json:"name"`
    BaseFee   *decimal.Decimal `json:"baseFee"`
    Markup    *decimal.Decimal `json:"markup"`
    Discount  *decimal.Decimal `json:"discount"`
    MinFee    *decimal.Decimal `json:"minFee"`
    MaxFee    *decimal.Decimal `json:"maxFee"`
    Priority  *int             `json:"priority"`
    Metadata  map[string]interface{} `json:"metadata"`
}

// CorridorFilter represents filters for listing corridors
type CorridorFilter struct {
    SourceCountry  string `json:"sourceCountry"`
    TargetCountry  string `json:"targetCountry"`
    SourceCurrency string `json:"sourceCurrency"`
    TargetCurrency string `json:"targetCurrency"`
    IsActive       *bool  `json:"isActive"`
    Limit          int    `json:"limit"`
    Offset         int    `json:"offset"`
}

// CalculationRequest represents a fee calculation request
type CalculationRequest struct {
    TransactionID   string          `json:"transactionId"`
    Amount          decimal.Decimal `json:"amount"`
    Currency        string          `json:"currency"`
    SourceCountry   string          `json:"sourceCountry"`
    TargetCountry   string          `json:"targetCountry"`
    SourceCurrency  string          `json:"sourceCurrency"`
    TargetCurrency  string          `json:"targetCurrency"`
    FeeCodes        []string        `json:"feeCodes"`
    CorridorCode    string          `json:"corridorCode"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// CalculationResult represents the result of a fee calculation
type CalculationResult struct {
    TotalFee        decimal.Decimal            `json:"totalFee"`
    Currency        string                     `json:"currency"`
    FeeBreakdown    []FeeBreakdown             `json:"feeBreakdown"`
    CorridorApplied *corridor.FeeCorridor      `json:"corridorApplied"`
    Timestamp       time.Time                  `json:"timestamp"`
    Metadata        map[string]interface{}     `json:"metadata"`
}

// FeeBreakdown represents a breakdown of fee components
type FeeBreakdown struct {
    FeeID       string          `json:"feeId"`
    FeeCode     string          `json:"feeCode"`
    FeeName     string          `json:"feeName"`
    FeeType     FeeType         `json:"feeType"`
    Amount      decimal.Decimal `json:"amount"`
    Description string          `json:"description"`
}