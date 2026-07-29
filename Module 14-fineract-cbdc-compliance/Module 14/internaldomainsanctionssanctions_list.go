package sanctions

import (
    "time"
)

// SanctionsList represents a sanctions list entry
type SanctionsList struct {
    ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name            string    `json:"name" gorm:"type:varchar(200);index"`
    Aliases         []string  `json:"aliases" gorm:"type:jsonb"`
    Type            string    `json:"type" gorm:"type:varchar(50)"`
    Source          string    `json:"source" gorm:"type:varchar(50)"` // OFAC, UN, EU, etc.
    Country         string    `json:"country" gorm:"type:varchar(3)"`
    Nationality     string    `json:"nationality" gorm:"type:varchar(3)"`
    DateOfBirth     string    `json:"dateOfBirth" gorm:"type:varchar(20)"`
    Identification  []string  `json:"identification" gorm:"type:jsonb"`
    Reasons         []string  `json:"reasons" gorm:"type:jsonb"`
    ListedDate      time.Time `json:"listedDate"`
    UnlistedDate    *time.Time `json:"unlistedDate,omitempty"`
    IsActive        bool      `json:"isActive" gorm:"default:true"`
    LastUpdated     time.Time `json:"lastUpdated"`
    CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (SanctionsList) TableName() string {
    return "cbdc_sanctions_lists"
}

// Matches checks if a name matches the sanctions list
func (s *SanctionsList) Matches(name string) bool {
    if s.Name == name {
        return true
    }
    for _, alias := range s.Aliases {
        if alias == name {
            return true
        }
    }
    return false
}