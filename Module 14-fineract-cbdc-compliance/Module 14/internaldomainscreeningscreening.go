package screening

import (
    "time"

    "github.com/shopspring/decimal"
)

// Screening represents a compliance screening operation
type Screening struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TransactionID   string          `json:"transactionId" gorm:"type:varchar(36);index"`
    CustomerID      string          `json:"customerId" gorm:"type:varchar(36);index"`
    CustomerName    string          `json:"customerName" gorm:"type:varchar(200)"`
    CustomerCountry string          `json:"customerCountry" gorm:"type:varchar(3)"`
    Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(38,18)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    SourceCountry   string          `json:"sourceCountry" gorm:"type:varchar(3)"`
    TargetCountry   string          `json:"targetCountry" gorm:"type:varchar(3)"`
    Status          ScreeningStatus `json:"status" gorm:"type:varchar(30)"`
    Type            ScreeningType   `json:"type" gorm:"type:varchar(20)"`
    Result          string          `json:"result" gorm:"type:text"`
    Score           int             `json:"score"`
    Details         map[string]interface{} `json:"details" gorm:"type:jsonb"`
    MatchedSanctions []string        `json:"matchedSanctions" gorm:"type:jsonb"`
    ErrorMessage    string          `json:"errorMessage" gorm:"type:text"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    CompletedAt     *time.Time      `json:"completedAt,omitempty"`
    ExpiresAt       time.Time       `json:"expiresAt"`
}

// TableName returns the table name for GORM
func (Screening) TableName() string {
    return "cbdc_screenings"
}

// IsTerminal checks if the screening is in a terminal state
func (s *Screening) IsTerminal() bool {
    return s.Status == StatusCompleted || s.Status == StatusFailed || s.Status == StatusBlocked
}

// IsExpired checks if the screening is expired
func (s *Screening) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the screening is still valid
func (s *Screening) IsValid() bool {
    return s.Status == StatusCompleted && !s.IsExpired() && s.Result == ResultApproved
}