package handlers

import (
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-compliance/internal/domain/compliance"
    "github.com/apache/fineract-cbdc-compliance/internal/domain/screening"
    "github.com/apache/fineract-cbdc-compliance/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// ComplianceHandler handles compliance HTTP endpoints
type ComplianceHandler struct {
    complianceService compliance.Service
    logger            *zap.Logger
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(complianceService compliance.Service, logger *zap.Logger) *ComplianceHandler {
    return &ComplianceHandler{
        complianceService: complianceService,
        logger:            logger,
    }
}

// ScreenTransactionRequest represents the screen transaction request
type ScreenTransactionRequest struct {
    TransactionID   string `json:"transactionId" validate:"required"`
    CustomerID      string `json:"customerId" validate:"required"`
    CustomerName    string `json:"customerName" validate:"required"`
    CustomerCountry string `json:"customerCountry" validate:"required,len=3"`
    Amount          string `json:"amount" validate:"required"`
    Currency        string `json:"currency" validate:"required,len=3"`
    SourceCountry   string `json:"sourceCountry" validate:"required,len=3"`
    TargetCountry   string `json:"targetCountry" validate:"required,len=3"`
    Type            string `json:"type"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ScreenTransactionResponse represents the screen transaction response
type ScreenTransactionResponse struct {
    ScreeningID string `json:"screeningId"`
    Status      string `json:"status"`
    Result      string `json:"result"`
    Score       int    `json:"score"`
    CreatedAt   string `json:"createdAt"`
}

// ScreenTransaction handles POST /api/v1/compliance/screen
func (h *ComplianceHandler) ScreenTransaction(c *gin.Context) {
    var req ScreenTransactionRequest
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

    // Parse screening type
    screeningType := screening.TypeSanctions
    if req.Type != "" {
        screeningType = screening.ScreeningType(req.Type)
    }

    // Create screening request
    screeningReq := &compliance.ScreeningRequest{
        TransactionID:   req.TransactionID,
        CustomerID:      req.CustomerID,
        CustomerName:    req.CustomerName,
        CustomerCountry: req.CustomerCountry,
        Amount:          amount,
        Currency:        req.Currency,
        SourceCountry:   req.SourceCountry,
        TargetCountry:   req.TargetCountry,
        Type:            screeningType,
        Metadata:        req.Metadata,
    }

    // Execute screening
    screening, err := h.complianceService.ScreenTransaction(c.Request.Context(), screeningReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusAccepted, ScreenTransactionResponse{
        ScreeningID: screening.ID,
        Status:      string(screening.Status),
        Result:      screening.Result,
        Score:       screening.Score,
        CreatedAt:   screening.CreatedAt.Format(time.RFC3339),
    })
}

// GetScreening handles GET /api/v1/compliance/screenings/:id
func (h *ComplianceHandler) GetScreening(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", "screening id is required"))
        return
    }

    screening, err := h.complianceService.GetScreeningStatus(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, screeningToResponse(screening))
}

// ListScreenings handles GET /api/v1/compliance/screenings
func (h *ComplianceHandler) ListScreenings(c *gin.Context) {
    var filter compliance.ScreeningFilter
    if err := c.ShouldBindQuery(&filter); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if filter.Limit == 0 {
        filter.Limit = 20
    }

    screenings, total, err := h.complianceService.ListScreenings(c.Request.Context(), &filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    responses := make([]ScreeningResponse, len(screenings))
    for i, s := range screenings {
        responses[i] = screeningToResponse(s)
    }

    c.JSON(http.StatusOK, gin.H{
        "data":  responses,
        "total": total,
        "page":  filter.Offset / filter.Limit,
        "limit": filter.Limit,
    })
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

// ScreeningResponse represents a screening response
type ScreeningResponse struct {
    ID              string    `json:"id"`
    TransactionID   string    `json:"transactionId"`
    CustomerID      string    `json:"customerId"`
    CustomerName    string    `json:"customerName"`
    CustomerCountry string    `json:"customerCountry"`
    Amount          string    `json:"amount"`
    Currency        string    `json:"currency"`
    SourceCountry   string    `json:"sourceCountry"`
    TargetCountry   string    `json:"targetCountry"`
    Status          string    `json:"status"`
    Type            string    `json:"type"`
    Result          string    `json:"result"`
    Score           int       `json:"score"`
    MatchedSanctions []string `json:"matchedSanctions"`
    ErrorMessage    string    `json:"errorMessage"`
    CreatedAt       time.Time `json:"createdAt"`
    CompletedAt     *time.Time `json:"completedAt"`
}

// screeningToResponse converts a screening to a response
func screeningToResponse(s *screening.Screening) ScreeningResponse {
    return ScreeningResponse{
        ID:              s.ID,
        TransactionID:   s.TransactionID,
        CustomerID:      s.CustomerID,
        CustomerName:    s.CustomerName,
        CustomerCountry: s.CustomerCountry,
        Amount:          s.Amount.String(),
        Currency:        s.Currency,
        SourceCountry:   s.SourceCountry,
        TargetCountry:   s.TargetCountry,
        Status:          string(s.Status),
        Type:            string(s.Type),
        Result:          s.Result,
        Score:           s.Score,
        MatchedSanctions: s.MatchedSanctions,
        ErrorMessage:    s.ErrorMessage,
        CreatedAt:       s.CreatedAt,
        CompletedAt:     s.CompletedAt,
    }
}