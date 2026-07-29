package handlers

import (
    "net/http"

    "github.com/apache/fineract-cbdc-orchestrator/internal/domain/transaction"
    "github.com/apache/fineract-cbdc-orchestrator/pkg/errors"
    "github.com/apache/fineract-cbdc-orchestrator/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// OrchestratorHandler handles transaction HTTP endpoints
type OrchestratorHandler struct {
    transactionService transaction.Service
    logger             *zap.Logger
}

// NewOrchestratorHandler creates a new orchestrator handler
func NewOrchestratorHandler(
    transactionService transaction.Service,
    logger *zap.Logger,
) *OrchestratorHandler {
    return &OrchestratorHandler{
        transactionService: transactionService,
        logger:             logger,
    }
}

// InitiatePaymentRequest represents the initiate payment request
type InitiatePaymentRequest struct {
    SourceCountry   string                 `json:"sourceCountry" validate:"required"`
    TargetCountry   string                 `json:"targetCountry" validate:"required"`
    SourceAccountID string                 `json:"sourceAccountId" validate:"required"`
    TargetAccountID string                 `json:"targetAccountId" validate:"required"`
    SourceCurrency  string                 `json:"sourceCurrency" validate:"required,len=3"`
    TargetCurrency  string                 `json:"targetCurrency" validate:"required,len=3"`
    Amount          string                 `json:"amount" validate:"required"`
    Description     string                 `json:"description"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// InitiatePaymentResponse represents the initiate payment response
type InitiatePaymentResponse struct {
    TransactionID string    `json:"transactionId"`
    Status        string    `json:"status"`
    CreatedAt     time.Time `json:"createdAt"`
}

// InitiatePayment handles POST /api/v1/payments
func (h *OrchestratorHandler) InitiatePayment(c *gin.Context) {
    var req InitiatePaymentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    // Validate request
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

    // Get idempotency key from header
    idempotencyKey := c.GetHeader("Idempotency-Key")

    // Create payment request
    paymentReq := &transaction.PaymentRequest{
        SourceCountry:   req.SourceCountry,
        TargetCountry:   req.TargetCountry,
        SourceAccountID: req.SourceAccountID,
        TargetAccountID: req.TargetAccountID,
        SourceCurrency:  req.SourceCurrency,
        TargetCurrency:  req.TargetCurrency,
        Amount:          amount,
        Description:     req.Description,
        IdempotencyKey:  idempotencyKey,
        Metadata:        req.Metadata,
    }

    // Process payment
    tx, err := h.transactionService.InitiatePayment(c.Request.Context(), paymentReq)
    if err != nil {
        h.logger.Error("Failed to initiate payment",
            zap.Error(err),
            zap.String("source", req.SourceCountry),
            zap.String("target", req.TargetCountry),
        )
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusAccepted, InitiatePaymentResponse{
        TransactionID: tx.ID,
        Status:        string(tx.Status),
        CreatedAt:     tx.CreatedAt,
    })
}

// GetTransaction handles GET /api/v1/payments/:id
func (h *OrchestratorHandler) GetTransaction(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", "transaction id is required"))
        return
    }

    tx, err := h.transactionService.GetTransaction(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    if tx == nil {
        c.JSON(http.StatusNotFound, ErrorResponse("NOT_FOUND", "transaction not found"))
        return
    }

    c.JSON(http.StatusOK, transactionToResponse(tx))
}

// ListTransactions handles GET /api/v1/payments
func (h *OrchestratorHandler) ListTransactions(c *gin.Context) {
    var filter transaction.TransactionFilter
    if err := c.ShouldBindQuery(&filter); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    // Set defaults
    if filter.Limit == 0 {
        filter.Limit = 20
    }
    if filter.SortBy == "" {
        filter.SortBy = "created_at"
        filter.SortOrder = "DESC"
    }

    transactions, total, err := h.transactionService.ListTransactions(c.Request.Context(), &filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    responses := make([]TransactionResponse, len(transactions))
    for i, tx := range transactions {
        responses[i] = transactionToResponse(tx)
    }

    c.JSON(http.StatusOK, gin.H{
        "data":  responses,
        "total": total,
        "page":  filter.Offset / filter.Limit,
        "limit": filter.Limit,
    })
}

// CancelTransaction handles DELETE /api/v1/payments/:id
func (h *OrchestratorHandler) CancelTransaction(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", "transaction id is required"))
        return
    }

    var req struct {
        Reason string `json:"reason"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := h.transactionService.CancelTransaction(c.Request.Context(), id, req.Reason); err != nil {
        if err == transaction.ErrTransactionNotFound {
            c.JSON(http.StatusNotFound, ErrorResponse("NOT_FOUND", "transaction not found"))
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "transaction cancelled"})
}

// RetryTransaction handles POST /api/v1/payments/:id/retry
func (h *OrchestratorHandler) RetryTransaction(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", "transaction id is required"))
        return
    }

    if err := h.transactionService.RetryTransaction(c.Request.Context(), id); err != nil {
        if err == transaction.ErrTransactionNotFound {
            c.JSON(http.StatusNotFound, ErrorResponse("NOT_FOUND", "transaction not found"))
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "transaction retry initiated"})
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

// TransactionResponse represents a transaction response
type TransactionResponse struct {
    ID              string                 `json:"id"`
    Type            string                 `json:"type"`
    State           string                 `json:"state"`
    Status          string                 `json:"status"`
    SourceCountry   string                 `json:"sourceCountry"`
    TargetCountry   string                 `json:"targetCountry"`
    SourceAccountID string                 `json:"sourceAccountId"`
    TargetAccountID string                 `json:"targetAccountId"`
    SourceCurrency  string                 `json:"sourceCurrency"`
    TargetCurrency  string                 `json:"targetCurrency"`
    SourceAmount    string                 `json:"sourceAmount"`
    TargetAmount    string                 `json:"targetAmount"`
    ConversionRate  string                 `json:"conversionRate"`
    LockReference   string                 `json:"lockReference"`
    SettlementID    string                 `json:"settlementId"`
    ErrorMessage    string                 `json:"errorMessage"`
    CancelReason    string                 `json:"cancelReason"`
    Attempts        int                    `json:"attempts"`
    Metadata        map[string]interface{} `json:"metadata"`
    CreatedAt       time.Time              `json:"createdAt"`
    UpdatedAt       time.Time              `json:"updatedAt"`
    CompletedAt     *time.Time             `json:"completedAt"`
    FailedAt        *time.Time             `json:"failedAt"`
    CancelledAt     *time.Time             `json:"cancelledAt"`
}

// transactionToResponse converts a transaction to a response
func transactionToResponse(tx *transaction.Transaction) TransactionResponse {
    return TransactionResponse{
        ID:              tx.ID,
        Type:            string(tx.Type),
        State:           string(tx.State),
        Status:          string(tx.Status),
        SourceCountry:   tx.SourceCountry,
        TargetCountry:   tx.TargetCountry,
        SourceAccountID: tx.SourceAccountID,
        TargetAccountID: tx.TargetAccountID,
        SourceCurrency:  tx.SourceCurrency,
        TargetCurrency:  tx.TargetCurrency,
        SourceAmount:    tx.SourceAmount.String(),
        TargetAmount:    tx.TargetAmount.String(),
        ConversionRate:  tx.ConversionRate.String(),
        LockReference:   tx.LockReference,
        SettlementID:    tx.SettlementID,
        ErrorMessage:    tx.ErrorMessage,
        CancelReason:    tx.CancelReason,
        Attempts:        tx.Attempts,
        Metadata:        tx.Metadata,
        CreatedAt:       tx.CreatedAt,
        UpdatedAt:       tx.UpdatedAt,
        CompletedAt:     tx.CompletedAt,
        FailedAt:        tx.FailedAt,
        CancelledAt:     tx.CancelledAt,
    }
}