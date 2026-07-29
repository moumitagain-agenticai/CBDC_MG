package compliance

import (
    "context"

    "github.com/apache/fineract-cbdc-compliance/internal/domain/screening"
    "github.com/apache/fineract-cbdc-compliance/internal/domain/sanctions"
)

// Service defines the compliance service interface
type Service interface {
    // Screening Operations
    ScreenTransaction(ctx context.Context, req *ScreeningRequest) (*screening.Screening, error)
    GetScreeningStatus(ctx context.Context, screeningID string) (*screening.Screening, error)
    ListScreenings(ctx context.Context, filter *ScreeningFilter) ([]*screening.Screening, int64, error)
    RejectScreening(ctx context.Context, screeningID string, reason string) error
    EscalateScreening(ctx context.Context, screeningID string, reason string) error

    // Sanctions Operations
    CheckSanctions(ctx context.Context, req *SanctionsCheckRequest) (*SanctionsCheckResponse, error)
    GetSanctionsList(ctx context.Context, filter *SanctionsFilter) ([]*sanctions.SanctionsList, int64, error)
    UpdateSanctionsList(ctx context.Context, entry *sanctions.SanctionsList) error
    ReloadSanctionsList(ctx context.Context) error

    // Compliance Operations
    CheckCompliance(ctx context.Context, req *ComplianceRequest) (*ComplianceCheck, error)
    GetComplianceStatus(ctx context.Context, complianceID string) (*ComplianceCheck, error)
    ListComplianceChecks(ctx context.Context, filter *ComplianceFilter) ([]*ComplianceCheck, int64, error)

    // Audit Operations
    GetAuditTrail(ctx context.Context, filter *AuditFilter) ([]*AuditTrail, int64, error)
    GetComplianceReport(ctx context.Context, req *ReportRequest) (*ComplianceReport, error)

    // Health & Monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*ServiceMetrics, error)
}

// ScreeningRequest represents a screening request
type ScreeningRequest struct {
    TransactionID   string          `json:"transactionId"`
    CustomerID      string          `json:"customerId"`
    CustomerName    string          `json:"customerName"`
    CustomerCountry string          `json:"customerCountry"`
    Amount          decimal.Decimal `json:"amount"`
    Currency        string          `json:"currency"`
    SourceCountry   string          `json:"sourceCountry"`
    TargetCountry   string          `json:"targetCountry"`
    Type            screening.ScreeningType `json:"type"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ScreeningFilter represents filters for listing screenings
type ScreeningFilter struct {
    Status    screening.ScreeningStatus `json:"status"`
    Type      screening.ScreeningType   `json:"type"`
    Customer  string                    `json:"customer"`
    FromDate  *time.Time                `json:"fromDate"`
    ToDate    *time.Time                `json:"toDate"`
    Limit     int                       `json:"limit"`
    Offset    int                       `json:"offset"`
}

// SanctionsCheckRequest represents a sanctions check request
type SanctionsCheckRequest struct {
    Name        string   `json:"name"`
    Aliases     []string `json:"aliases"`
    Country     string   `json:"country"`
    Nationality string   `json:"nationality"`
    DateOfBirth string   `json:"dateOfBirth"`
}

// SanctionsCheckResponse represents a sanctions check response
type SanctionsCheckResponse struct {
    IsMatch      bool     `json:"isMatch"`
    MatchedItems []sanctions.SanctionsList `json:"matchedItems"`
    Confidence   int      `json:"confidence"`
    Score        int      `json:"score"`
    Recommendation string `json:"recommendation"` // APPROVE, REVIEW, BLOCK
}

// ComplianceRequest represents a compliance check request
type ComplianceRequest struct {
    TransactionID string                 `json:"transactionId"`
    CustomerID    string                 `json:"customerId"`
    Type          string                 `json:"type"`
    Data          map[string]interface{} `json:"data"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// ComplianceFilter represents filters for listing compliance checks
type ComplianceFilter struct {
    Status   ComplianceStatus `json:"status"`
    Type     string           `json:"type"`
    Customer string           `json:"customer"`
    FromDate *time.Time       `json:"fromDate"`
    ToDate   *time.Time       `json:"toDate"`
    Limit    int              `json:"limit"`
    Offset   int              `json:"offset"`
}

// AuditFilter represents filters for audit trail
type AuditFilter struct {
    Action    string    `json:"action"`
    UserID    string    `json:"userId"`
    FromDate  *time.Time `json:"fromDate"`
    ToDate    *time.Time `json:"toDate"`
    Limit     int       `json:"limit"`
    Offset    int       `json:"offset"`
}

// ReportRequest represents a compliance report request
type ReportRequest struct {
    Type      string    `json:"type"`
    StartDate time.Time `json:"startDate"`
    EndDate   time.Time `json:"endDate"`
    Format    string    `json:"format"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
    ReportID     string    `json:"reportId"`
    Type         string    `json:"type"`
    GeneratedAt  time.Time `json:"generatedAt"`
    Summary      ReportSummary `json:"summary"`
    Details      []ReportDetail `json:"details"`
}

// ReportSummary represents a report summary
type ReportSummary struct {
    TotalScreenings     int `json:"totalScreenings"`
    BlockedScreenings   int `json:"blockedScreenings"`
    EscalatedScreenings int `json:"escalatedScreenings"`
    SanctionsHits       int `json:"sanctionsHits"`
    CompliancePasses    int `json:"compliancePasses"`
    ComplianceFails     int `json:"complianceFails"`
}

// ReportDetail represents a report detail
type ReportDetail struct {
    TransactionID string `json:"transactionId"`
    CustomerID    string `json:"customerId"`
    Type          string `json:"type"`
    Status        string `json:"status"`
    Timestamp     time.Time `json:"timestamp"`
}