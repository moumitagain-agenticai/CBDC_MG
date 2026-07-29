package client

import (
    "bytes"
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/fineract/cbdc/india-connector/internal/config"
    "github.com/fineract/cbdc/india-connector/internal/domain"
    "github.com/fineract/cbdc/india-connector/internal/ports"
    "github.com/fineract/cbdc/india-connector/pkg/logger"
    "github.com/fineract/cbdc/india-connector/pkg/metrics"
    "github.com/fineract/cbdc/india-connector/pkg/utils"
    "github.com/hashicorp/go-retryablehttp"
    "github.com/sony/gobreaker"
    "go.uber.org/zap"
)

// CBDCClient implements the CBDCClient interface for Indian CBDC API
type CBDCClient struct {
    config        *config.CBDCConfig
    httpClient    *http.Client
    retryClient   *retryablehttp.Client
    circuitBreaker *gobreaker.CircuitBreaker
    logger        *zap.Logger
}

// NewCBDCClient creates a new CBDC client
func NewCBDCClient(cfg *config.CBDCConfig, log *zap.Logger) (*CBDCClient, error) {
    // Configure HTTP client
    httpClient := &http.Client{
        Timeout: cfg.Timeout,
    }

    // Configure retry client
    retryClient := retryablehttp.NewClient()
    retryClient.HTTPClient = httpClient
    retryClient.RetryMax = 3
    retryClient.RetryWaitMin = 1 * time.Second
    retryClient.RetryWaitMax = 30 * time.Second
    retryClient.Backoff = utils.ExponentialBackoff

    // Configure circuit breaker
    cbSettings := gobreaker.Settings{
        Name:          "cbdc-client",
        MaxRequests:   5,
        Interval:      30 * time.Second,
        Timeout:       60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 10 && failureRatio >= 0.5
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            log.Warn("circuit breaker state changed",
                zap.String("name", name),
                zap.String("from", from.String()),
                zap.String("to", to.String()),
            )
        },
    }
    cb := gobreaker.NewCircuitBreaker(cbSettings)

    return &CBDCClient{
        config:         cfg,
        httpClient:     httpClient,
        retryClient:    retryClient,
        circuitBreaker: cb,
        logger:         log,
    }, nil
}

// Issue implements the CBDCClient interface
func (c *CBDCClient) Issue(ctx context.Context, req *ports.IssueRequest) (*ports.IssueResponse, error) {
    var resp ports.IssueResponse
    err := c.executeWithCircuitBreaker(ctx, "issue", c.config.Endpoints.Issue, req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// Transfer implements the CBDCClient interface
func (c *CBDCClient) Transfer(ctx context.Context, req *ports.TransferRequest) (*ports.TransferResponse, error) {
    var resp ports.TransferResponse
    err := c.executeWithCircuitBreaker(ctx, "transfer", c.config.Endpoints.Transfer, req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// Lock implements the CBDCClient interface
func (c *CBDCClient) Lock(ctx context.Context, req *ports.LockRequest) (*ports.LockResponse, error) {
    var resp ports.LockResponse
    err := c.executeWithCircuitBreaker(ctx, "lock", c.config.Endpoints.Lock, req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// Burn implements the CBDCClient interface
func (c *CBDCClient) Burn(ctx context.Context, req *ports.BurnRequest) (*ports.BurnResponse, error) {
    var resp ports.BurnResponse
    err := c.executeWithCircuitBreaker(ctx, "burn", c.config.Endpoints.Burn, req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// Redeem implements the CBDCClient interface
func (c *CBDCClient) Redeem(ctx context.Context, req *ports.RedeemRequest) (*ports.RedeemResponse, error) {
    var resp ports.RedeemResponse
    err := c.executeWithCircuitBreaker(ctx, "redeem", c.config.Endpoints.Redeem, req, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetBalance implements the CBDCClient interface
func (c *CBDCClient) GetBalance(ctx context.Context, req *ports.BalanceRequest) (*ports.BalanceResponse, error) {
    var resp ports.BalanceResponse
    endpoint := fmt.Sprintf("%s?wallet_id=%s&currency=%s", c.config.Endpoints.Balance, req.WalletID, req.Currency)
    err := c.executeWithCircuitBreaker(ctx, "balance", endpoint, nil, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetTransactionStatus implements the CBDCClient interface
func (c *CBDCClient) GetTransactionStatus(ctx context.Context, txID string) (*ports.TransactionStatusResponse, error) {
    var resp ports.TransactionStatusResponse
    endpoint := fmt.Sprintf("%s/%s", c.config.Endpoints.Status, txID)
    err := c.executeWithCircuitBreaker(ctx, "status", endpoint, nil, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// HealthCheck implements the CBDCClient interface
func (c *CBDCClient) HealthCheck(ctx context.Context) (*ports.HealthResponse, error) {
    var resp ports.HealthResponse
    err := c.execute(ctx, "GET", c.config.Endpoints.Health, nil, &resp)
    if err != nil {
        return nil, err
    }
    return &resp, nil
}

// executeWithCircuitBreaker executes a request with circuit breaker protection
func (c *CBDCClient) executeWithCircuitBreaker(ctx context.Context, operation, endpoint string, reqBody interface{}, respBody interface{}) error {
    return c.circuitBreaker.Execute(func() (interface{}, error) {
        return nil, c.execute(ctx, "POST", endpoint, reqBody, respBody)
    })
}

// execute executes an HTTP request to the CBDC API
func (c *CBDCClient) execute(ctx context.Context, method, endpoint string, reqBody interface{}, respBody interface{}) error {
    start := time.Now()
    operation := endpoint

    // Prepare request body
    var body io.Reader
    if reqBody != nil {
        data, err := json.Marshal(reqBody)
        if err != nil {
            return domain.NewDomainError(domain.ErrorInternal, "failed to marshal request")
        }
        body = bytes.NewReader(data)
    }

    // Build full URL
    url := c.config.BaseURL + endpoint

    // Create request
    req, err := http.NewRequestWithContext(ctx, method, url, body)
    if err != nil {
        return domain.NewDomainError(domain.ErrorInternal, "failed to create request")
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", c.config.APIKey)
    req.Header.Set("X-Sponsor-Bank-ID", c.config.SponsorBankID)
    req.Header.Set("X-Merchant-ID", c.config.MerchantID)
    req.Header.Set("X-Terminal-ID", c.config.TerminalID)

    // Generate HMAC signature
    signature, err := c.generateSignature(reqBody)
    if err == nil {
        req.Header.Set("X-Signature", signature)
    }

    // Execute request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        c.logger.Error("CBDC API request failed",
            zap.Error(err),
            zap.String("endpoint", endpoint),
        )
        return domain.NewDomainError(domain.ErrorTransactionFailed, "CBDC API request failed")
    }
    defer resp.Body.Close()

    // Record metrics
    metrics.RecordRequest("cbdc_client", operation, resp.StatusCode, time.Since(start))

    // Read response body
    respData, err := io.ReadAll(resp.Body)
    if err != nil {
        return domain.NewDomainError(domain.ErrorInternal, "failed to read response")
    }

    // Check for error status
    if resp.StatusCode >= 400 {
        var errResp struct {
            Code    string `json:"code"`
            Message string `json:"message"`
        }
        if err := json.Unmarshal(respData, &errResp); err == nil {
            return c.mapHTTPError(resp.StatusCode, errResp.Code, errResp.Message)
        }
        return c.mapHTTPError(resp.StatusCode, "", string(respData))
    }

    // Parse response
    if respBody != nil && len(respData) > 0 {
        if err := json.Unmarshal(respData, respBody); err != nil {
            return domain.NewDomainError(domain.ErrorInternal, "failed to parse response")
        }
    }

    return nil
}

// generateSignature generates an HMAC-SHA256 signature for the request
func (c *CBDCClient) generateSignature(data interface{}) (string, error) {
    if data == nil {
        return "", nil
    }

    jsonData, err := json.Marshal(data)
    if err != nil {
        return "", err
    }

    h := hmac.New(sha256.New, []byte(c.config.APISecret))
    h.Write(jsonData)
    return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// mapHTTPError maps HTTP status codes to domain errors
func (c *CBDCClient) mapHTTPError(statusCode int, code, message string) error {
    switch statusCode {
    case http.StatusBadRequest:
        return domain.NewDomainErrorWithDetails(domain.ErrorValidation, message, map[string]interface{}{
            "status_code": statusCode,
            "code": code,
        })
    case http.StatusUnauthorized:
        return domain.NewDomainError(domain.ErrorUnauthorized, "unauthorized access to CBDC API")
    case http.StatusForbidden:
        return domain.NewDomainError(domain.ErrorUnauthorized, "forbidden access to CBDC API")
    case http.StatusNotFound:
        return domain.NewDomainError(domain.ErrorNotFound, "resource not found")
    case http.StatusConflict:
        return domain.NewDomainError(domain.ErrorConflict, "conflict with existing transaction")
    case http.StatusTooManyRequests:
        return domain.NewDomainError(domain.ErrorRateLimit, "rate limit exceeded")
    case http.StatusServiceUnavailable:
        return domain.NewDomainError(domain.ErrorTransactionTimeout, "CBDC service unavailable")
    default:
        return domain.NewDomainError(domain.ErrorTransactionFailed, fmt.Sprintf("CBDC API error: %s", message))
    }
}