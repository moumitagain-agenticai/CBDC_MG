package handlers

import (
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-settlement/internal/domain/settlement"
    "github.com/apache/fineract-cbdc-settlement/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// SettlementHandler handles settlement HTTP endpoints
type SettlementHandler struct {
    settlementService settlement.Service
    logger            *zap.Logger
}

// NewSettlementHandler creates a new settlement handler
func NewSettlementHandler(settlementService settlement.Service, logger *zap.Logger) *SettlementHandler {
    return &SettlementHandler{
        settlementService: settlementService,
        logger:            logger,
    }
}

// InitiateSettlementRequest represents the initiate settlement request
type InitiateSettlementRequest struct {
    TransactionID  string `json:"transactionId" validate:"required"`
    SourceNetwork  string `json:"sourceNetwork" validate:"required"`
    TargetNetwork  string `json:"targetNetwork" validate:"required"`
    SourceAccountID string `json:"sourceAccountId" validate:"required"`
    TargetAccountID string `json:"targetAccountId" validate:"required"`
    SourceCurrency string `json:"sourceCurrency" validate:"required,len=3"`
    TargetCurrency string `json:"targetCurrency" validate:"required,len=3"`
    SourceAmount   string `json:"sourceAmount" validate:"required"`
    TargetAmount   string `json:"targetAmount" validate:"required"`
    ConversionRate string `json:"conversionRate"`
    Type           string `json:"type"`
    Metadata       map[string]interface{} `json:"metadata"`
}

// InitiateSettlementResponse represents the initiate settlement response
type InitiateSettlementResponse struct {
    SettlementID string `json:"settlementId"`
    Status       string `json:"status"`
    CreatedAt    string `json:"createdAt"`
}

// InitiateSettlement handles POST /api/v1/settlements
func (h *SettlementHandler) InitiateSettlement(c *gin.Context) {
    var req InitiateSettlementRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    // Parse amounts
    sourceAmount, err := decimal.NewFromString(req.SourceAmount)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_AMOUNT", "invalid source amount"))
        return
    }

    targetAmount, err := decimal.NewFromString(req.TargetAmount)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_AMOUNT", "invalid target amount"))
        return
    }

    // Parse conversion rate
    conversionRate := decimal.Zero
    if req.ConversionRate != "" {
        conversionRate, err = decimal.NewFromString(req.ConversionRate)
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_RATE", "invalid conversion rate"))
            return
        }
    }

    // Parse settlement type
    settlementType := settlement.TypePvP
    if req.Type != "" {
        settlementType = settlement.SettlementType(req.Type)
    }

    // Create settlement request
    settlementReq := &settlement.SettlementRequest{
        TransactionID:  req.TransactionID,
        SourceNetwork:  req.SourceNetwork,
        TargetNetwork:  req.TargetNetwork,
        SourceAccountID: req.SourceAccountID,
        TargetAccountID: req.TargetAccountID,
        SourceCurrency: req.SourceCurrency,
        TargetCurrency: req.TargetCurrency,
        SourceAmount:   sourceAmount,
        TargetAmount:   targetAmount,
        ConversionRate: conversionRate,
        Type:           settlementType,
        Metadata:       req.Metadata,
    }

    // Initiate settlement
    settlement, err := h.settlementService.InitiateSettlement(c.Request.Context(), settlementReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusAccepted, InitiateSettlementResponse{
        SettlementID: settlement.ID,
        Status:       string(settlement.Status),
        CreatedAt:    settlement.CreatedAt.Format(time.RFC3339),
    })
}

// GetSettlement handles GET /api/v1/settlements/:id
func (h *SettlementHandler) GetSettlement(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", "settlement id is required"))
        return
    }

    settlement, err := h.settlementService.GetSettlement(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, settlementToResponse(settlement))
}

// ListSettlements handles GET /api/v1/settlements
func (h *SettlementHandler) ListSettlements(c *gin.Context) {
    var filter settlement.SettlementFilter
    if err := c.ShouldBindQuery(&filter); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if filter.Limit == 0 {
        filter.Limit = 20
    }

    settlements, total, err := h.settlementService.ListSettlements(c.Request.Context(), &filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    responses := make([]SettlementResponse, len(settlements))
    for i, s := range settlements {
        responses[i] = settlementToResponse(s)
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

// SettlementResponse represents a settlement response
type SettlementResponse struct {
    ID                 string    `json:"id"`
    TransactionID      string    `json:"transactionId"`
    SourceNetwork      string    `json:"sourceNetwork"`
    TargetNetwork      string    `json:"targetNetwork"`
    SourceAccountID    string    `json:"sourceAccountId"`
    TargetAccountID    string    `json:"targetAccountId"`
    SourceCurrency     string    `json:"sourceCurrency"`
    TargetCurrency     string    `json:"targetCurrency"`
    SourceAmount       string    `json:"sourceAmount"`
    TargetAmount       string    `json:"targetAmount"`
    ConversionRate     string    `json:"conversionRate"`
    SourceLockID       string    `json:"sourceLockId"`
    TargetLockID       string    `json:"targetLockId"`
    BurnTransactionID  string    `json:"burnTransactionId"`
    IssueTransactionID string    `json:"issueTransactionId"`
    Status             string    `json:"status"`
    Type               string    `json:"type"`
    ErrorMessage       string    `json:"errorMessage"`
    RetryCount         int       `json:"retryCount"`
    CreatedAt          time.Time `json:"createdAt"`
    UpdatedAt          time.Time `json:"updatedAt"`
    CompletedAt        *time.Time `json:"completedAt"`
    FailedAt           *time.Time `json:"failedAt"`
}

// settlementToResponse converts a settlement to a response
func settlementToResponse(s *settlement.Settlement) SettlementResponse {
    return SettlementResponse{
        ID:                 s.ID,
        TransactionID:      s.TransactionID,
        SourceNetwork:      s.SourceNetwork,
        TargetNetwork:      s.TargetNetwork,
        SourceAccountID:    s.SourceAccountID,
        TargetAccountID:    s.TargetAccountID,
        SourceCurrency:     s.SourceCurrency,
        TargetCurrency:     s.TargetCurrency,
        SourceAmount:       s.SourceAmount.String(),
        TargetAmount:       s.TargetAmount.String(),
        ConversionRate:     s.ConversionRate.String(),
        SourceLockID:       s.SourceLockID,
        TargetLockID:       s.TargetLockID,
        BurnTransactionID:  s.BurnTransactionID,
        IssueTransactionID: s.IssueTransactionID,
        Status:             string(s.Status),
        Type:               string(s.Type),
        ErrorMessage:       s.ErrorMessage,
        RetryCount:         s.RetryCount,
        CreatedAt:          s.CreatedAt,
        UpdatedAt:          s.UpdatedAt,
        CompletedAt:        s.CompletedAt,
        FailedAt:           s.FailedAt,
    }
}