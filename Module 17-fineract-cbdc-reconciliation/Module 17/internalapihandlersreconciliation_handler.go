package handlers

import (
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-reconciliation/internal/domain/reconciliation"
    "github.com/apache/fineract-cbdc-reconciliation/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// ReconciliationHandler handles reconciliation HTTP endpoints
type ReconciliationHandler struct {
    reconciliationService reconciliation.Service
    logger                *zap.Logger
}

// NewReconciliationHandler creates a new reconciliation handler
func NewReconciliationHandler(reconciliationService reconciliation.Service, logger *zap.Logger) *ReconciliationHandler {
    return &ReconciliationHandler{
        reconciliationService: reconciliationService,
        logger:                logger,
    }
}

// CreateReconciliationRequest represents the create reconciliation request
type CreateReconciliationRequest struct {
    Name          string `json:"name" validate:"required"`
    Type          string `json:"type" validate:"required"`
    AccountID     string `json:"accountId" validate:"required"`
    AccountNumber string `json:"accountNumber" validate:"required"`
    Currency      string `json:"currency" validate:"required,len=3"`
    StartDate     string `json:"startDate" validate:"required"`
    EndDate       string `json:"endDate" validate:"required"`
    OpeningBalance string `json:"openingBalance" validate:"required"`
    TenantID      string `json:"tenantId" validate:"required"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// CreateReconciliationResponse represents the create reconciliation response
type CreateReconciliationResponse struct {
    ReconciliationID string `json:"reconciliationId"`
    Status           string `json:"status"`
    CreatedAt        string `json:"createdAt"`
}

// UploadStatementRequest represents the upload statement request
type UploadStatementRequest struct {
    ReconciliationID string `json:"reconciliationId" validate:"required"`
    StatementType    string `json:"statementType" validate:"required"`
    FileName         string `json:"fileName" validate:"required"`
    FileContent      []byte `json:"fileContent" validate:"required"`
    Metadata         map[string]interface{} `json:"metadata"`
}

// CreateReconciliation handles POST /api/v1/reconciliations
func (h *ReconciliationHandler) CreateReconciliation(c *gin.Context) {
    var req CreateReconciliationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    // Parse dates
    startDate, err := time.Parse(time.RFC3339, req.StartDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "invalid start date format"))
        return
    }

    endDate, err := time.Parse(time.RFC3339, req.EndDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "invalid end date format"))
        return
    }

    // Parse opening balance
    openingBalance, err := decimal.NewFromString(req.OpeningBalance)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_AMOUNT", "invalid opening balance format"))
        return
    }

    // Create reconciliation request
    createReq := &reconciliation.CreateReconciliationRequest{
        Name:          req.Name,
        Type:          reconciliation.ReconciliationType(req.Type),
        AccountID:     req.AccountID,
        AccountNumber: req.AccountNumber,
        Currency:      req.Currency,
        StartDate:     startDate,
        EndDate:       endDate,
        OpeningBalance: openingBalance,
        TenantID:      req.TenantID,
        Metadata:      req.Metadata,
    }

    rec, err := h.reconciliationService.CreateReconciliation(c.Request.Context(), createReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, CreateReconciliationResponse{
        ReconciliationID: rec.ID,
        Status:           string(rec.Status),
        CreatedAt:        rec.CreatedAt.Format(time.RFC3339),
    })
}

// UploadStatement handles POST /api/v1/reconciliations/:id/statements
func (h *ReconciliationHandler) UploadStatement(c *gin.Context) {
    var req UploadStatementRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    uploadReq := &reconciliation.UploadStatementRequest{
        ReconciliationID: req.ReconciliationID,
        StatementType:    req.StatementType,
        FileName:         req.FileName,
        FileContent:      req.FileContent,
        Metadata:         req.Metadata,
    }

    statement, err := h.reconciliationService.UploadStatement(c.Request.Context(), uploadReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusAccepted, gin.H{
        "statementId":  statement.ID,
        "status":       string(statement.Status),
        "entriesCount": len(statement.Entries),
    })
}

// ExecuteMatching handles POST /api/v1/reconciliations/:id/match
func (h *ReconciliationHandler) ExecuteMatching(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", "reconciliation id is required"))
        return
    }

    result, err := h.reconciliationService.ExecuteMatching(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "reconciliationId": result.ReconciliationID,
        "totalEntries":     result.TotalEntries,
        "matchedEntries":   result.MatchedEntries,
        "unmatchedEntries": result.UnmatchedEntries,
        "systemBalance":    result.SystemBalance.String(),
        "bankBalance":      result.BankBalance.String(),
        "difference":       result.Difference.String(),
        "isBalanced":       result.IsBalanced,
        "completedAt":      result.CompletedAt.Format(time.RFC3339),
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