// Package systemic detects the approved, scan-relative structural finding.
package systemic

import "github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"

const CriticalityThreshold = 80

func Detect(packageID string, score domain.StructuralCriticality, check domain.PackageCheck, coverageComplete bool, knownAdvisory bool) domain.SystemicDependency {
	out := domain.SystemicDependency{PackageID: packageID, Status: "UNAVAILABLE", Threshold: CriticalityThreshold, CriticalityScore: score.Score, Provider: "OSV", CheckedAt: check.CheckedAt, ReasonCodes: []string{}, Explanation: []string{}}
	if score.Score == nil {
		out.ReasonCodes = append(out.ReasonCodes, "CRITICALITY_UNAVAILABLE")
		out.ReasonCodes = append(out.ReasonCodes, score.ReasonCodes...)
	}
	if !coverageComplete || check.PackageID != packageID || check.Status != "CHECKED" || check.CheckedAt == nil {
		out.ReasonCodes = append(out.ReasonCodes, "INCOMPLETE_ENRICHMENT_COVERAGE")
	}
	if len(out.ReasonCodes) > 0 {
		out.Explanation = append(out.Explanation, "Systemic classification requires an available criticality score and complete timestamped OSV coverage for this scan.")
		return out
	}
	out.Status = "NOT_MATCHED"
	if len(check.Advisories) > 0 || knownAdvisory {
		out.ReasonCodes = append(out.ReasonCodes, "KNOWN_ADVISORY_PRESENT")
	} else if *score.Score < CriticalityThreshold {
		out.ReasonCodes = append(out.ReasonCodes, "BELOW_CRITICALITY_THRESHOLD")
	} else {
		out.Status = "MATCH"
		out.Explanation = append(out.Explanation, "SYSTEMIC DEPENDENCY — no known advisory returned by OSV at scan time.", "Structural criticality is at least 80/100 and all scan packages have complete vulnerability checks.")
	}
	out.Explanation = append(out.Explanation, "This prototype classification is relative to this scan. No matching advisory is not evidence of safety or absence of undiscovered vulnerabilities.")
	return out
}
