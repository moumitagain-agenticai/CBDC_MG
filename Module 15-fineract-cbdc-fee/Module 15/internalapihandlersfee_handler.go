package handlers

import (
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-fee/internal/domain/fee"
    "github.com/apache/fineract-cbdc-fee/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// FeeHandler handles fee HTTP endpoints
type FeeHandler struct {
    feeService fee.Service
    logger     *zap.Logger
}

// NewFeeHandler creates a new fee handler
func NewFeeHandler(feeService fee.Service, logger *zap.Logger) *FeeHandler {
    return &FeeHandler{
        feeService: feeService,
        logger:     logger,
    }
}

// CalculateFeeRequest represents the calculate fee request
type CalculateFeeRequest struct {
    TransactionID  string `json:"transactionId" validate:"required"`
    Amount         string `json:"amount" validate:"required"`
    Currency       string `json:"currency" validate:"required,len=3"`
    SourceCountry  string `json:"sourceCountry" validate:"required,len=3"`
    TargetCountry  string `json:"targetCountry" validate:"required,len=3"`
    SourceCurrency string `json:"sourceCurrency" validate:"required,len=3"`
    TargetCurrency string `json:"targetCurrency" validate:"required,len=3"`
    FeeCodes       []string `json:"feeCodes"`
    CorridorCode   string `json:"corridorCode"`
    Metadata       map[string]interface{} `json:"metadata"`
}

// CalculateFeeResponse represents the calculate fee response
type CalculateFeeResponse struct {
    TotalFee     string                 `json:"totalFee"`
    Currency     string                 `json:"currency"`
    FeeBreakdown []FeeBreakdownResponse `json:"feeBreakdown"`
    Corridor     *CorridorResponse      `json:"corridor,omitempty"`
    Timestamp    string                 `json:"timestamp"`
}

// FeeBreakdownResponse represents a fee breakdown in the response
type FeeBreakdownResponse struct {
    FeeID       string `json:"feeId"`
    FeeCode     string `json:"feeCode"`
    FeeName     string `json:"feeName"`
    FeeType     string `json:"feeType"`
    Amount      string `json:"amount"`
    Description string `json:"description"`
}

// CorridorResponse represents a corridor in the response
type CorridorResponse struct {
    ID      string `json:"id"`
    Code    string `json:"code"`
    Name    string `json:"name"`
    BaseFee string `json:"baseFee"`
    Markup  string `json:"markup"`
    Discount string `json:"discount"`
}

// CalculateFee handles POST /api/v1/fee/calculate
func (h *FeeHandler) CalculateFee(c *gin.Context) {
    var req CalculateFeeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    // Parse amount
    amount, err := decimal.NewFromString(req.Amount)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_AMOUNT", "invalid amount format"))
        return
    }

    // Create calculation request
    calcReq := &fee.CalculationRequest{
        TransactionID:  req.TransactionID,
        Amount:         amount,
        Currency:       req.Currency,
        SourceCountry:  req.SourceCountry,
        TargetCountry:  req.TargetCountry,
        SourceCurrency: req.SourceCurrency,
        TargetCurrency: req.TargetCurrency,
        FeeCodes:       req.FeeCodes,
        CorridorCode:   req.CorridorCode,
        Metadata:       req.Metadata,
    }

    // Calculate fee
    result, err := h.feeService.CalculateFee(c.Request.Context(), calcReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    // Build response
    breakdown := make([]FeeBreakdownResponse, len(result.FeeBreakdown))
    for i, b := range result.FeeBreakdown {
        breakdown[i] = FeeBreakdownResponse{
            FeeID:       b.FeeID,
            FeeCode:     b.FeeCode,
            FeeName:     b.FeeName,
            FeeType:     string(b.FeeType),
            Amount:      b.Amount.String(),
            Description: b.Description,
        }
    }

    response := CalculateFeeResponse{
        TotalFee:     result.TotalFee.String(),
        Currency:     result.Currency,
        FeeBreakdown: breakdown,
        Timestamp:    result.Timestamp.Format(time.RFC3339),
    }

    if result.CorridorApplied != nil {
        response.Corridor = &CorridorResponse{
            ID:       result.CorridorApplied.ID,
            Code:     result.CorridorApplied.Code,
            Name:     result.CorridorApplied.Name,
            BaseFee:  result.CorridorApplied.BaseFee.String(),
            Markup:   result.CorridorApplied.Markup.String(),
            Discount: result.CorridorApplied.Discount.String(),
        }
    }

    c.JSON(http.StatusOK, response)
}

// CreateFeeRequest represents the create fee request
type CreateFeeRequest struct {
    Name            string                 `json:"name" validate:"required"`
    Code            string                 `json:"code" validate:"required"`
    Type            string                 `json:"type" validate:"required"`
    Structure       string                 `json:"structure" validate:"required"`
    Value           string                 `json:"value" validate:"required"`
    MinAmount       string                 `json:"minAmount"`
    MaxAmount       string                 `json:"maxAmount"`
    TieredStructure []fee.Tier             `json:"tieredStructure"`
    CorridorID      string                 `json:"corridorId"`
    SourceCountry   string                 `json:"sourceCountry"`
    TargetCountry   string                 `json:"targetCountry"`
    SourceCurrency  string                 `json:"sourceCurrency"`
    TargetCurrency  string                 `json:"targetCurrency"`
    Priority        int                    `json:"priority"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// CreateFee handles POST /api/v1/fee/config
func (h *FeeHandler) CreateFee(c *gin.Context) {
    var req CreateFeeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    // Parse value
    value, err := decimal.NewFromString(req.Value)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_VALUE", "invalid value format"))
        return
    }

    // Parse min/max amounts
    minAmount := decimal.Zero
    if req.MinAmount != "" {
        minAmount, err = decimal.NewFromString(req.MinAmount)
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MIN_AMOUNT", "invalid min amount format"))
            return
        }
    }

    maxAmount := decimal.Zero
    if req.MaxAmount != "" {
        maxAmount, err = decimal.NewFromString(req.MaxAmount)
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MAX_AMOUNT", "invalid max amount format"))
            return
        }
    }

    // Create fee request
    createReq := &fee.CreateFeeRequest{
        Name:            req.Name,
        Code:            req.Code,
        Type:            fee.FeeType(req.Type),
        Structure:       fee.FeeStructure(req.Structure),
        Value:           value,
        MinAmount:       minAmount,
        MaxAmount:       maxAmount,
        TieredStructure: req.TieredStructure,
        CorridorID:      req.CorridorID,
        SourceCountry:   req.SourceCountry,
        TargetCountry:   req.TargetCountry,
        SourceCurrency:  req.SourceCurrency,
        TargetCurrency:  req.TargetCurrency,
        Priority:        req.Priority,
        Metadata:        req.Metadata,
    }

    fee, err := h.feeService.CreateFee(c.Request.Context(), createReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, fee)
}

// ErrorResponse returns a standardized error response
func ErrorResponse(code, message string) gin.H {
    return gin.H{
        "error": gin.H{
            "code":    code,
            "message": message,
        },
    }
}