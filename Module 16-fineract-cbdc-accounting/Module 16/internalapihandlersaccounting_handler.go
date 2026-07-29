package handlers

import (
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-accounting/internal/domain/accounting"
    "github.com/apache/fineract-cbdc-accounting/internal/domain/ledger"
    "github.com/apache/fineract-cbdc-accounting/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// AccountingHandler handles accounting HTTP endpoints
type AccountingHandler struct {
    accountingService accounting.Service
    logger            *zap.Logger
}

// NewAccountingHandler creates a new accounting handler
func NewAccountingHandler(accountingService accounting.Service, logger *zap.Logger) *AccountingHandler {
    return &AccountingHandler{
        accountingService: accountingService,
        logger:            logger,
    }
}

// PostJournalEntryRequest represents the post journal entry request
type PostJournalEntryRequest struct {
    TransactionID string                 `json:"transactionId" validate:"required"`
    Entries       []JournalEntryItem     `json:"entries" validate:"required,min=1"`
    Description   string                 `json:"description"`
    ReferenceID   string                 `json:"referenceId"`
    ReferenceType string                 `json:"referenceType"`
    TenantID      string                 `json:"tenantId" validate:"required"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// JournalEntryItem represents a journal entry item
type JournalEntryItem struct {
    AccountID   string `json:"accountId" validate:"required"`
    AccountCode string `json:"accountCode"`
    EntryType   string `json:"entryType" validate:"required,oneof=DEBIT CREDIT"`
    Amount      string `json:"amount" validate:"required"`
    Currency    string `json:"currency" validate:"required,len=3"`
    Description string `json:"description"`
}

// PostJournalEntryResponse represents the post journal entry response
type PostJournalEntryResponse struct {
    EntryIDs     []string `json:"entryIds"`
    TransactionID string  `json:"transactionId"`
    TotalDebit   string  `json:"totalDebit"`
    TotalCredit  string  `json:"totalCredit"`
    IsBalanced   bool    `json:"isBalanced"`
    CreatedAt    string  `json:"createdAt"`
}

// PostJournalEntry handles POST /api/v1/accounting/journal
func (h *AccountingHandler) PostJournalEntry(c *gin.Context) {
    var req PostJournalEntryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    // Convert to domain request
    entries := make([]accounting.JournalEntryItem, len(req.Entries))
    for i, item := range req.Entries {
        amount, err := decimal.NewFromString(item.Amount)
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_AMOUNT", "invalid amount format"))
            return
        }

        entries[i] = accounting.JournalEntryItem{
            AccountID:   item.AccountID,
            AccountCode: item.AccountCode,
            EntryType:   ledger.EntryType(item.EntryType),
            Amount:      amount,
            Currency:    item.Currency,
            Description: item.Description,
        }
    }

    journalReq := &accounting.JournalEntryRequest{
        TransactionID: req.TransactionID,
        Entries:       entries,
        Description:   req.Description,
        ReferenceID:   req.ReferenceID,
        ReferenceType: req.ReferenceType,
        TenantID:      req.TenantID,
        Metadata:      req.Metadata,
    }

    result, err := h.accountingService.PostJournalEntry(c.Request.Context(), journalReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    response := PostJournalEntryResponse{
        EntryIDs:     result.EntryIDs,
        TransactionID: result.TransactionID,
        TotalDebit:   result.TotalDebit.String(),
        TotalCredit:  result.TotalCredit.String(),
        IsBalanced:   result.IsBalanced,
        CreatedAt:    result.CreatedAt.Format(time.RFC3339),
    }

    c.JSON(http.StatusCreated, response)
}

// RevalueCurrencyRequest represents the revalue currency request
type RevalueCurrencyRequest struct {
    Currency     string `json:"currency" validate:"required,len=3"`
    TenantID     string `json:"tenantId" validate:"required"`
    NewRate      string `json:"newRate" validate:"required"`
    ReferenceID  string `json:"referenceId"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// RevalueCurrencyResponse represents the revalue currency response
type RevalueCurrencyResponse struct {
    RevaluationID  string `json:"revaluationId"`
    Currency       string `json:"currency"`
    OldRate        string `json:"oldRate"`
    NewRate        string `json:"newRate"`
    GainLoss       string `json:"gainLoss"`
    GainLossType   string `json:"gainLossType"`
    Status         string `json:"status"`
    CompletedAt    string `json:"completedAt"`
}

// RevalueCurrency handles POST /api/v1/accounting/revalue
func (h *AccountingHandler) RevalueCurrency(c *gin.Context) {
    var req RevalueCurrencyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    newRate, err := decimal.NewFromString(req.NewRate)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_RATE", "invalid rate format"))
        return
    }

    revalReq := &accounting.RevaluationRequest{
        Currency:    req.Currency,
        TenantID:    req.TenantID,
        NewRate:     newRate,
        RevaluationDate: time.Now(),
        ReferenceID: req.ReferenceID,
        Metadata:    req.Metadata,
    }

    reval, err := h.accountingService.RevalueCurrency(c.Request.Context(), revalReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    response := RevalueCurrencyResponse{
        RevaluationID: reval.ID,
        Currency:      reval.Currency,
        OldRate:       reval.OldRate.String(),
        NewRate:       reval.NewRate.String(),
        GainLoss:      reval.GainLoss.String(),
        GainLossType:  string(reval.GainLossType),
        Status:        string(reval.Status),
        CompletedAt:   reval.RevaluationDate.Format(time.RFC3339),
    }

    c.JSON(http.StatusOK, response)
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