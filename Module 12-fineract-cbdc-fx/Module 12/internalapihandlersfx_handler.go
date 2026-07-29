package handlers

import (
    "net/http"
    "time"

    "github.com/apache/fineract-cbdc-fx/internal/domain/fx"
    "github.com/apache/fineract-cbdc-fx/internal/domain/rate"
    "github.com/apache/fineract-cbdc-fx/internal/domain/quote"
    "github.com/apache/fineract-cbdc-fx/internal/domain/conversion"
    "github.com/apache/fineract-cbdc-fx/pkg/validator"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// FXHandler handles FX HTTP endpoints
type FXHandler struct {
    fxService fx.Service
    logger    *zap.Logger
}

// NewFXHandler creates a new FX handler
func NewFXHandler(fxService fx.Service, logger *zap.Logger) *FXHandler {
    return &FXHandler{
        fxService: fxService,
        logger:    logger,
    }
}

// GetRateRequest represents the get rate request
type GetRateRequest struct {
    BaseCurrency  string `json:"baseCurrency" form:"base" validate:"required,len=3"`
    QuoteCurrency string `json:"quoteCurrency" form:"quote" validate:"required,len=3"`
}

// GetRateResponse represents the get rate response
type GetRateResponse struct {
    BaseCurrency  string `json:"baseCurrency"`
    QuoteCurrency string `json:"quoteCurrency"`
    BidRate       string `json:"bidRate"`
    AskRate       string `json:"askRate"`
    MidRate       string `json:"midRate"`
    Spread        string `json:"spread"`
    Provider      string `json:"provider"`
    Status        string `json:"status"`
    ExpiresAt     string `json:"expiresAt"`
}

// GetRate handles GET /api/v1/fx/rate
func (h *FXHandler) GetRate(c *gin.Context) {
    var req GetRateRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    rate, err := h.fxService.GetRate(c.Request.Context(), req.BaseCurrency, req.QuoteCurrency)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, GetRateResponse{
        BaseCurrency:  rate.BaseCurrency,
        QuoteCurrency: rate.QuoteCurrency,
        BidRate:       rate.BidRate.String(),
        AskRate:       rate.AskRate.String(),
        MidRate:       rate.MidRate.String(),
        Spread:        rate.Spread.String(),
        Provider:      rate.Provider,
        Status:        string(rate.Status),
        ExpiresAt:     rate.ExpiresAt.Format(time.RFC3339),
    })
}

// GetQuoteRequest represents the get quote request
type GetQuoteRequest struct {
    TransactionID   string `json:"transactionId" validate:"required"`
    BaseCurrency    string `json:"baseCurrency" validate:"required,len=3"`
    QuoteCurrency   string `json:"quoteCurrency" validate:"required,len=3"`
    Amount          string `json:"amount" validate:"required"`
    LockDuration    string `json:"lockDuration" validate:"required"`
    SlippageTolerance string `json:"slippageTolerance"`
}

// GetQuoteResponse represents the get quote response
type GetQuoteResponse struct {
    ID              string `json:"id"`
    TransactionID   string `json:"transactionId"`
    BaseCurrency    string `json:"baseCurrency"`
    QuoteCurrency   string `json:"quoteCurrency"`
    BaseAmount      string `json:"baseAmount"`
    QuoteAmount     string `json:"quoteAmount"`
    Rate            string `json:"rate"`
    BidRate         string `json:"bidRate"`
    AskRate         string `json:"askRate"`
    Spread          string `json:"spread"`
    MarkupPercent   string `json:"markupPercent"`
    MarkupAmount    string `json:"markupAmount"`
    SlippagePercent string `json:"slippagePercent"`
    SlippageAmount  string `json:"slippageAmount"`
    FinalRate       string `json:"finalRate"`
    Status          string `json:"status"`
    ExpiresAt       string `json:"expiresAt"`
}

// GetQuote handles POST /api/v1/fx/quote
func (h *FXHandler) GetQuote(c *gin.Context) {
    var req GetQuoteRequest
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

    // Parse lock duration
    lockDuration, err := time.ParseDuration(req.LockDuration)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DURATION", "invalid lock duration format"))
        return
    }

    // Parse slippage tolerance
    var slippageTolerance decimal.Decimal
    if req.SlippageTolerance != "" {
        slippageTolerance, err = decimal.NewFromString(req.SlippageTolerance)
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_SLIPPAGE", "invalid slippage tolerance"))
            return
        }
    }

    quoteReq := &fx.QuoteRequest{
        TransactionID:   req.TransactionID,
        BaseCurrency:    req.BaseCurrency,
        QuoteCurrency:   req.QuoteCurrency,
        Amount:          amount,
        LockDuration:    lockDuration,
        SlippageTolerance: slippageTolerance,
    }

    quote, err := h.fxService.GetQuote(c.Request.Context(), quoteReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(http.StatusOK, GetQuoteResponse{
        ID:              quote.ID,
        TransactionID:   quote.TransactionID,
        BaseCurrency:    quote.BaseCurrency,
        QuoteCurrency:   quote.QuoteCurrency,
        BaseAmount:      quote.BaseAmount.String(),
        QuoteAmount:     quote.QuoteAmount.String(),
        Rate:            quote.Rate.String(),
        BidRate:         quote.BidRate.String(),
        AskRate:         quote.AskRate.String(),
        Spread:          quote.Spread.String(),
        MarkupPercent:   quote.MarkupPercent.String(),
        MarkupAmount:    quote.MarkupAmount.String(),
        SlippagePercent: quote.SlippagePercent.String(),
        SlippageAmount:  quote.SlippageAmount.String(),
        FinalRate:       quote.FinalRate.String(),
        Status:          string(quote.Status),
        ExpiresAt:       quote.ExpiresAt.Format(time.RFC3339),
    })
}

// ConvertRequest represents the conversion request
type ConvertRequest struct {
    TransactionID string `json:"transactionId" validate:"required"`
    FromCurrency  string `json:"fromCurrency" validate:"required,len=3"`
    ToCurrency    string `json:"toCurrency" validate:"required,len=3"`
    Amount        string `json:"amount" validate:"required"`
    QuoteID       string `json:"quoteId,omitempty"`
}

// ConvertResponse represents the conversion response
type ConvertResponse struct {
    ID            string `json:"id"`
    TransactionID string `json:"transactionId"`
    QuoteID       string `json:"quoteId"`
    FromCurrency  string `json:"fromCurrency"`
    ToCurrency    string `json:"toCurrency"`
    FromAmount    string `json:"fromAmount"`
    ToAmount      string `json:"toAmount"`
    RateUsed      string `json:"rateUsed"`
    FeeAmount     string `json:"feeAmount"`
    FeeCurrency   string `json:"feeCurrency"`
    Status        string `json:"status"`
    CompletedAt   string `json:"completedAt,omitempty"`
}

// ConvertCurrency handles POST /api/v1/fx/convert
func (h *FXHandler) ConvertCurrency(c *gin.Context) {
    var req ConvertRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
        return
    }

    if err := validator.ValidateStruct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("VALIDATION_ERROR", err.Error()))
        return
    }

    amount, err := decimal.NewFromString(req.Amount)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_AMOUNT", "invalid amount format"))
        return
    }

    var conversion *conversion.Conversion

    if req.QuoteID != "" {
        // Convert using quote
        conversion, err = h.fxService.ConvertWithQuote(c.Request.Context(), req.QuoteID, amount)
    } else {
        // Convert with current rate
        conversionReq := &fx.ConversionRequest{
            TransactionID: req.TransactionID,
            FromCurrency:  req.FromCurrency,
            ToCurrency:    req.ToCurrency,
            Amount:        amount,
        }
        conversion, err = h.fxService.ConvertCurrency(c.Request.Context(), conversionReq)
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    completedAt := ""
    if conversion.CompletedAt != nil {
        completedAt = conversion.CompletedAt.Format(time.RFC3339)
    }

    c.JSON(http.StatusOK, ConvertResponse{
        ID:            conversion.ID,
        TransactionID: conversion.TransactionID,
        QuoteID:       conversion.QuoteID,
        FromCurrency:  conversion.FromCurrency,
        ToCurrency:    conversion.ToCurrency,
        FromAmount:    conversion.FromAmount.String(),
        ToAmount:      conversion.ToAmount.String(),
        RateUsed:      conversion.RateUsed.String(),
        FeeAmount:     conversion.FeeAmount.String(),
        FeeCurrency:   conversion.FeeCurrency,
        Status:        string(conversion.Status),
        CompletedAt:   completedAt,
    })
}