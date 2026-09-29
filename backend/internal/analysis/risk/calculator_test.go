package risk

import (
	"strings"
	"testing"
	"time"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

const vector = "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"

func TestCVSSBaseVectors(t *testing.T) {
	for _, tc := range []struct {
		vector string
		want   float64
	}{
		{vector, 9.8},
		{strings.Replace(vector, "3.1", "3.0", 1), 9.8},
		{strings.Replace(vector, "S:U", "S:C", 1), 10},
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H", 7.5},
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:U/C:L/I:N/A:N", 4.3},
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N", 0},
		{"CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:C/C:H/I:H/A:H", 9.9},
		{vector + "/E:P/RL:O/RC:C/MAV:P/CR:H", 9.8}, // BASE, not temporal/environmental
	} {
		got, _, ok := cvss3(tc.vector)
		if !ok || got != tc.want {
			t.Errorf("%s: got %v, %v; want %v", tc.vector, got, ok, tc.want)
		}
	}
	for _, bad := range []string{"9.8", "NaN", "HIGH", "", strings.Replace(vector, "3.1", "4.0", 1), strings.Replace(vector, "/AV:N", "", 1), vector + "/AV:N", vector + "/E:Z", vector + "/FOO:X", strings.Replace(vector, "PR:N", "PR:X", 1), vector + "/E:", vector + "/"} {
		if _, _, ok := cvss3(bad); ok {
			t.Errorf("accepted malformed/unsupported vector %q", bad)
		}
	}
}

func riskFixture() (domain.ImpactAnalysis, domain.PackageCheck) {
	blast, criticality := 70, 80
	now := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	return domain.ImpactAnalysis{PackageID: "x", ApplicationsExposed: 2, DownstreamPackages: 3, BlastRadius: domain.BlastRadius{Score: &blast}, StructuralCriticality: domain.StructuralCriticality{Score: &criticality}}, domain.PackageCheck{PackageID: "x", Status: "CHECKED", CheckedAt: &now, Advisories: []domain.Advisory{{ID: "OSV-1", Source: "OSV", MatchConfidence: domain.ConfidenceHigh, Severity: []domain.Severity{{Type: "CVSS_V3", Score: vector + "/E:P"}}}}}
}

func TestWeightedScoreAndMissingEvidence(t *testing.T) {
	impact, check := riskFixture()
	out := Analyze(impact, check, true)
	if out.Score == nil || *out.Score != 86 || len(out.ReasonCodes) != 0 {
		t.Fatalf("%+v", out)
	}
	if *out.Components["vulnerability_severity"].Value != 98 || *out.Components["exploitability_evidence"].Contribution != 15 {
		t.Fatal(out.Components)
	}
	for _, maturity := range []string{"", "/E:X", "/E:U"} {
		check.Advisories[0].Severity[0].Score = vector + maturity
		out = Analyze(impact, check, true)
		if out.Score != nil || out.Components["exploitability_evidence"].Value != nil || out.Components["exploitability_evidence"].Contribution != nil || out.Components["vulnerability_severity"].Value == nil {
			t.Fatalf("missing evidence was zero-filled or hid severity: %+v", out)
		}
	}
}

func TestRiskAvailability(t *testing.T) {
	for _, tc := range []string{"coverage", "unknown severity", "uncertain match", "blast", "criticality", "no advisories", "generic link", "other package", "other source"} {
		t.Run(tc, func(t *testing.T) {
			impact, check := riskFixture()
			complete := true
			switch tc {
			case "coverage":
				complete = false
			case "unknown severity":
				check.Advisories = append(check.Advisories, domain.Advisory{ID: "OSV-2", MatchConfidence: domain.ConfidenceHigh})
			case "uncertain match":
				check.Advisories[0].MatchConfidence = domain.ConfidenceMedium
			case "blast":
				impact.BlastRadius.Score = nil
				impact.BlastRadius.ReasonCodes = []string{"SCAN_PATH_COUNT_CAPPED"}
			case "criticality":
				impact.StructuralCriticality.Score = nil
			case "no advisories":
				check.Advisories = nil
			case "generic link":
				check.Advisories[0].Severity[0].Score = vector
				check.Advisories[0].References = []domain.Reference{{Type: "EVIDENCE", URL: "https://example.com/exploit"}}
			case "other package":
				check.PackageID = "other"
			case "other source":
				check.Advisories[0].Source = "untrusted"
			}
			out := Analyze(impact, check, complete)
			if out.Score != nil || len(out.ReasonCodes) == 0 {
				t.Fatalf("%+v", out)
			}
		})
	}
}
