package matching

// MatchResult represents the result of a matching operation
type MatchResult struct {
    SystemEntryID     string   `json:"systemEntryId"`
    BankEntryID       string   `json:"bankEntryId"`
    IsMatch           bool     `json:"isMatch"`
    ConfidenceScore   float64  `json:"confidenceScore"`
    MatchType         string   `json:"matchType"` // EXACT, FUZZY, PARTIAL
    MatchingRules     []string `json:"matchingRules"`
    Differences       []string `json:"differences"`
}