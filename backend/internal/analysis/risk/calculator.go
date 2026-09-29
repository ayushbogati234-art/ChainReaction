// Package risk implements the transparent Phase 5 package prioritization model.
package risk

import (
	"fmt"
	"math"
	"sort"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

const (
	SeverityWeight       = .35
	BlastWeight          = .30
	CriticalityWeight    = .20
	ExploitabilityWeight = .15
)

func component(weight float64) domain.RiskComponent {
	return domain.RiskComponent{Weight: weight, ReasonCodes: []string{}, Evidence: []string{}}
}
func set(c *domain.RiskComponent, value float64) {
	c.Value = &value
	contribution := value * c.Weight
	c.Contribution = &contribution
}
func topology(c *domain.RiskComponent, score *int, reasons []string, name string) {
	if score == nil {
		c.ReasonCodes = append(c.ReasonCodes, name+"_UNAVAILABLE")
		c.ReasonCodes = append(c.ReasonCodes, reasons...)
		return
	}
	set(c, float64(*score))
	c.Evidence = append(c.Evidence, fmt.Sprintf("%s: %d/100 within this scan.", name, *score))
}

func Analyze(impact domain.ImpactAnalysis, check domain.PackageCheck, coverageComplete bool) domain.RiskScore {
	out := domain.RiskScore{Model: "phase5-v1", ReasonCodes: []string{}, Explanation: []string{fmt.Sprintf("Package-level model: %d application root(s) potentially exposed; %d downstream package(s).", impact.ApplicationsExposed, impact.DownstreamPackages), "Weights: severity 35%, Blast Radius 30%, structural criticality 20%, exploitability evidence 15%. Missing components are not zero-filled or reweighted."}}
	severity, blast, criticality, exploit := component(SeverityWeight), component(BlastWeight), component(CriticalityWeight), component(ExploitabilityWeight)
	topology(&blast, impact.BlastRadius.Score, impact.BlastRadius.ReasonCodes, "BLAST_RADIUS")
	topology(&criticality, impact.StructuralCriticality.Score, impact.StructuralCriticality.ReasonCodes, "STRUCTURAL_CRITICALITY")
	if !coverageComplete || check.PackageID != impact.PackageID || check.Status != "CHECKED" || check.CheckedAt == nil {
		out.ReasonCodes = append(out.ReasonCodes, "INCOMPLETE_ENRICHMENT_COVERAGE")
	}
	if len(check.Advisories) == 0 {
		severity.ReasonCodes = append(severity.ReasonCodes, "NO_MATCHED_ADVISORY")
	} else {
		maximum, allScored := 0.0, true
		for _, advisory := range check.Advisories {
			scored := false
			if advisory.MatchConfidence != domain.ConfidenceHigh {
				out.ReasonCodes = append(out.ReasonCodes, "UNCERTAIN_ADVISORY_MATCH")
			}
			for _, supplied := range advisory.Severity {
				if supplied.Type != "CVSS_V3" {
					continue
				}
				value, maturity, valid := cvss3(supplied.Score)
				if !valid {
					continue
				}
				scored = true
				maximum = math.Max(maximum, value*10)
				severity.Evidence = append(severity.Evidence, fmt.Sprintf("%s (%s): base CVSS %.1f/10 from %s", advisory.ID, advisory.Source, value, supplied.Score))
				if advisory.Source == "OSV" && (maturity == "P" || maturity == "F" || maturity == "H") {
					set(&exploit, 100)
					exploit.Evidence = append(exploit.Evidence, fmt.Sprintf("%s (OSV): explicit CVSS exploit-code maturity E:%s in %s", advisory.ID, maturity, supplied.Score))
				}
			}
			if !scored {
				allScored = false
				severity.Evidence = append(severity.Evidence, advisory.ID+": no supported valid CVSS v3 vector; labels and unsupported vectors are not converted to numbers.")
			}
		}
		if allScored {
			set(&severity, maximum)
		} else {
			severity.ReasonCodes = append(severity.ReasonCodes, "SEVERITY_UNAVAILABLE")
		}
	}
	if exploit.Value == nil {
		exploit.ReasonCodes = append(exploit.ReasonCodes, "EXPLOITABILITY_EVIDENCE_UNAVAILABLE")
	}
	out.Components = map[string]domain.RiskComponent{"vulnerability_severity": severity, "blast_radius": blast, "structural_criticality": criticality, "exploitability_evidence": exploit}
	total := 0.0
	for _, name := range []string{"vulnerability_severity", "blast_radius", "structural_criticality", "exploitability_evidence"} {
		c := out.Components[name]
		out.ReasonCodes = append(out.ReasonCodes, c.ReasonCodes...)
		if c.Contribution != nil {
			total += *c.Contribution
		}
	}
	sort.Strings(out.ReasonCodes)
	out.ReasonCodes = unique(out.ReasonCodes)
	if len(out.ReasonCodes) == 0 {
		value := int(math.Round(total))
		out.Score = &value
	}
	out.Explanation = append(out.Explanation, "Severity uses the maximum base CVSS across matched advisories, requiring a supported vector for every advisory. Positive exploit-code evidence contributes 15 points; absent, unproven or unspecified maturity stays unknown.", "This prototype score is not an exploit probability. Provider evidence does not establish exploitability in your application. Compare only within this scan.")
	return out
}

func unique(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}
