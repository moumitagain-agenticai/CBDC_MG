package compliance

import (
    "time"

    "github.com/shopspring/decimal"
)

// ComplianceCheck represents a compliance check
type ComplianceCheck struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TransactionID   string          `json:"transactionId" gorm:"type:varchar(36);index"`
    ScreeningID     string          `json:"screeningId" gorm:"type:varchar(36)"`
    CustomerID      string          `json:"customerId" gorm:"type:varchar(36);index"`
    Type            string          `json:"type" gorm:"type:varchar(50)"` // LRS, TCS, FEMA, etc.
    Status          ComplianceStatus `json:"status" gorm:"type:varchar(30)"`
    Result          ComplianceResult `json:"result" gorm:"type:jsonb"`
    RulesApplied    []string        `json:"rulesApplied" gorm:"type:jsonb"`
    Score           int             `json:"score"`
    Details         map[string]interface{} `json:"details" gorm:"type:jsonb"`
    ErrorMessage    string          `json:"errorMessage" gorm:"type:text"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    CompletedAt     *time.Time      `json:"completedAt,omitempty"`
}

// TableName returns the table name for GORM
func (ComplianceCheck) TableName() string {
    return "cbdc_compliance_checks"
}