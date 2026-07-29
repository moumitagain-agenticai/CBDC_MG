package api

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/fineract/cbdc/india-connector/internal/domain"
    "github.com/fineract/cbdc/india-connector/internal/service"
    "github.com/fineract/cbdc/india-connector/pkg/metrics"
    "github.com/go-chi/chi/v5"
    "github.com/go-playground/validator/v10"
    "go.uber.org/zap"
)

// Handler represents the API handler
type Handler struct {
    connectorService   *service.ConnectorService
    transactionService *service.TransactionService
    healthService      *service.HealthService
    logger             *zap.Logger
    validate           *validator.Validate
}

// NewHandler creates a new API handler
func NewHandler(
    connectorService *service.ConnectorService,
    transactionService *service.TransactionService,
    healthService *service.HealthService,
    logger *zap.Logger,
) *Handler {
    return &Handler{
        connectorService:   connectorService,
        transactionService: transactionService,
        healthService:      healthService,
        logger:             logger,
        validate:           validator.New(),
    }
}

// IssueCBDC handles CBDC issuance requests
// @Summary Issue CBDC tokens
// @Description Issue new CBDC tokens to a wallet
// @Tags CBDC Operations
// @Accept json
// @Produce json
// @Param request body domain.TransactionRequest true "Issue request"
// @Success 200 {object} domain.TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cbdc/issue [post]
func (h *Handler) IssueCBDC(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := r.Context()

    var req domain.TransactionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // Set transaction type
    req.Type = domain.TransactionTypeIssue

    // Validate request
    if err := h.validate.Struct(req); err != nil {
        h.writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
        return
    }

    // Process request
    resp, err := h.connectorService.IssueCBDC(ctx, &req)
    if err != nil {
        h.handleError(w, err)
        metrics.RecordAPIRequest("issue", "error", time.Since(start))
        return
    }

    metrics.RecordAPIRequest("issue", "success", time.Since(start))
    h.writeJSON(w, http.StatusOK, resp)
}

// TransferCBDC handles CBDC transfer requests
// @Summary Transfer CBDC tokens
// @Description Transfer CBDC tokens between wallets
// @Tags CBDC Operations
// @Accept json
// @Produce json
// @Param request body domain.TransactionRequest true "Transfer request"
// @Success 200 {object} domain.TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cbdc/transfer [post]
func (h *Handler) TransferCBDC(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := r.Context()

    var req domain.TransactionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    req.Type = domain.TransactionTypeTransfer

    if err := h.validate.Struct(req); err != nil {
        h.writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
        return
    }

    resp, err := h.connectorService.TransferCBDC(ctx, &req)
    if err != nil {
        h.handleError(w, err)
        metrics.RecordAPIRequest("transfer", "error", time.Since(start))
        return
    }

    metrics.RecordAPIRequest("transfer", "success", time.Since(start))
    h.writeJSON(w, http.StatusOK, resp)
}

// LockCBDC handles CBDC lock requests
// @Summary Lock CBDC tokens
// @Description Lock CBDC tokens for atomic settlement
// @Tags CBDC Operations
// @Accept json
// @Produce json
// @Param request body LockRequest true "Lock request"
// @Success 200 {object} domain.TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cbdc/lock [post]
func (h *Handler) LockCBDC(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := r.Context()

    var req struct {
        domain.TransactionRequest
        DurationSec int `json:"duration_sec" validate:"required,min=1,max=3600"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    req.Type = domain.TransactionTypeLock

    if err := h.validate.Struct(req); err != nil {
        h.writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
        return
    }

    resp, err := h.connectorService.LockCBDC(ctx, &req.TransactionRequest, req.DurationSec)
    if err != nil {
        h.handleError(w, err)
        metrics.RecordAPIRequest("lock", "error", time.Since(start))
        return
    }

    metrics.RecordAPIRequest("lock", "success", time.Since(start))
    h.writeJSON(w, http.StatusOK, resp)
}

// GetBalance handles balance requests
// @Summary Get CBDC balance
// @Description Get the balance of a CBDC wallet
// @Tags Query Operations
// @Produce json
// @Param wallet_id query string true "Wallet ID"
// @Param currency query string true "Currency code"
// @Success 200 {object} BalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cbdc/balance [get]
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := r.Context()

    walletID := r.URL.Query().Get("wallet_id")
    currency := r.URL.Query().Get("currency")

    if walletID == "" || currency == "" {
        h.writeError(w, http.StatusBadRequest, "wallet_id and currency are required")
        return
    }

    resp, err := h.transactionService.GetBalance(ctx, walletID, currency)
    if err != nil {
        h.handleError(w, err)
        metrics.RecordAPIRequest("balance", "error", time.Since(start))
        return
    }

    metrics.RecordAPIRequest("balance", "success", time.Since(start))
    h.writeJSON(w, http.StatusOK, resp)
}

// GetTransactionStatus handles transaction status requests
// @Summary Get transaction status
// @Description Get the status of a CBDC transaction
// @Tags Query Operations
// @Produce json
// @Param transaction_id path string true "Transaction ID"
// @Success 200 {object} TransactionStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/cbdc/transactions/{transaction_id}/status [get]
func (h *Handler) GetTransactionStatus(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    txID := chi.URLParam(r, "transaction_id")

    if txID == "" {
        h.writeError(w, http.StatusBadRequest, "transaction_id is required")
        return
    }

    resp, err := h.transactionService.GetTransactionStatus(ctx, txID)
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.writeJSON(w, http.StatusOK, resp)
}

// HealthCheck handles health check requests
// @Summary Health check
// @Description Check the health of the service and its dependencies
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 500 {object} ErrorResponse
// @Router /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    resp, err := h.healthService.Check(ctx)
    if err != nil {
        h.writeError(w, http.StatusServiceUnavailable, "service unhealthy")
        return
    }

    statusCode := http.StatusOK
    if resp.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }

    h.writeJSON(w, statusCode, resp)
}

// Readiness handles readiness check requests
// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags Health
// @Produce json
// @Success 200 {object} ReadinessResponse
// @Failure 500 {object} ErrorResponse
// @Router /ready [get]
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    ready, err := h.healthService.Ready(ctx)
    if err != nil || !ready {
        h.writeError(w, http.StatusServiceUnavailable, "service not ready")
        return
    }

    h.writeJSON(w, http.StatusOK, map[string]string{
        "status": "ready",
    })
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        h.logger.Error("failed to encode response", zap.Error(err))
    }
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
    h.writeJSON(w, status, ErrorResponse{
        Status:  "error",
        Message: message,
        Code:    http.StatusText(status),
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    })
}

// handleError handles domain errors
func (h *Handler) handleError(w http.ResponseWriter, err error) {
    if domainErr, ok := err.(*domain.DomainError); ok {
        var status int
        switch domainErr.Code {
        case domain.ErrorValidation:
            status = http.StatusBadRequest
        case domain.ErrorUnauthorized:
            status = http.StatusUnauthorized
        case domain.ErrorNotFound:
            status = http.StatusNotFound
        case domain.ErrorConflict:
            status = http.StatusConflict
        case domain.ErrorInsufficientBalance:
            status = http.StatusPaymentRequired
        case domain.ErrorRateLimit:
            status = http.StatusTooManyRequests
        case domain.ErrorTransactionTimeout:
            status = http.StatusGatewayTimeout
        case domain.ErrorCircuitOpen, domain.ErrorRetryExhausted:
            status = http.StatusServiceUnavailable
        default:
            status = http.StatusInternalServerError
        }
        h.writeError(w, status, domainErr.Message)
        return
    }

    h.writeError(w, http.StatusInternalServerError, "internal server error")
}

// ErrorResponse represents an error response
type ErrorResponse struct {
    Status    string `json:"status"`
    Message   string `json:"message"`
    Code      string `json:"code"`
    Timestamp string `json:"timestamp"`
}

// BalanceResponse represents a balance response
type BalanceResponse struct {
    WalletID  string `json:"wallet_id"`
    Balance   string `json:"balance"`
    Currency  string `json:"currency"`
    Available string `json:"available"`
    Locked    string `json:"locked"`
    UpdatedAt string `json:"updated_at"`
}

// TransactionStatusResponse represents a transaction status response
type TransactionStatusResponse struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Type          string `json:"type"`
    Amount        string `json:"amount"`
    Currency      string `json:"currency"`
    CreatedAt     string `json:"created_at"`
    CompletedAt   string `json:"completed_at,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
    Status    string                 `json:"status"`
    Version   string                 `json:"version"`
    Services  map[string]string      `json:"services"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp string                 `json:"timestamp"`
}

// ReadinessResponse represents a readiness check response
type ReadinessResponse struct {
    Status string `json:"status"`
}

// LockRequest represents a lock request
type LockRequest struct {
    domain.TransactionRequest
    DurationSec int `json:"duration_sec"`
}