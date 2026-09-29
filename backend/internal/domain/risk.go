package domain

import "time"

// RiskScore is a package-level prototype prioritization score, not a probability.
type RiskScore struct {
	Score       *int                     `json:"score"`
	Model       string                   `json:"model"`
	ReasonCodes []string                 `json:"reason_codes"`
	Components  map[string]RiskComponent `json:"components"`
	Explanation []string                 `json:"explanation"`
}

type RiskComponent struct {
	Value        *float64 `json:"value"` // normalized to 0–100
	Weight       float64  `json:"weight"`
	Contribution *float64 `json:"contribution"`
	ReasonCodes  []string `json:"reason_codes"`
	Evidence     []string `json:"evidence"`
}

type SystemicDependency struct {
	PackageID        string     `json:"package_id"`
	Status           string     `json:"status"` // MATCH, NOT_MATCHED, UNAVAILABLE
	Threshold        int        `json:"threshold"`
	CriticalityScore *int       `json:"criticality_score"`
	Provider         string     `json:"provider"`
	CheckedAt        *time.Time `json:"checked_at"`
	ReasonCodes      []string   `json:"reason_codes"`
	Explanation      []string   `json:"explanation"`
}

type SystemicReport struct {
	Status            string               `json:"status"` // COMPLETE or PARTIAL
	TotalPackages     int                  `json:"total_packages"`
	EvaluatedPackages int                  `json:"evaluated_packages"`
	Findings          []SystemicDependency `json:"findings"`
	ReasonCodes       []string             `json:"reason_codes"`
}
