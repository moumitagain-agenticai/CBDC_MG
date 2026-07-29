package lock

import (
    "time"

    "github.com/shopspring/decimal"
)

// FundLock represents a locked fund on a CBDC network
type FundLock struct {
    ID              string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    SettlementID    string          `json:"settlementId" gorm:"type:varchar(36);index"`
    Network         string          `json:"network" gorm:"type:varchar(20)"`
    AccountID       string          `json:"accountId" gorm:"type:varchar(100)"`
    LockID          string          `json:"lockId" gorm:"type:varchar(100);uniqueIndex"`
    Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(38,18)"`
    Currency        string          `json:"currency" gorm:"type:varchar(3)"`
    Status          LockStatus      `json:"status" gorm:"type:varchar(20)"`
    LockDuration    time.Duration   `json:"lockDuration"`
    ExpiresAt       time.Time       `json:"expiresAt"`
    ReleasedAt      *time.Time      `json:"releasedAt,omitempty"`
    CreatedAt       time.Time       `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
    Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (FundLock) TableName() string {
    return "cbdc_fund_locks"
}

// IsExpired checks if the lock is expired
func (l *FundLock) IsExpired() bool {
    return time.Now().After(l.ExpiresAt)
}

// IsActive checks if the lock is still active
func (l *FundLock) IsActive() bool {
    return l.Status == LockStatusActive && !l.IsExpired()
}