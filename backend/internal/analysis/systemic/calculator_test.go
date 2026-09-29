package systemic

import (
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"testing"
	"time"
)

func TestSystemicBoundaryAndEvidence(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name            string
		score           int
		complete, known bool
		status          string
	}{
		{"at boundary", 80, true, false, "MATCH"},
		{"below", 79, true, false, "NOT_MATCHED"},
		{"high", 100, true, false, "MATCH"},
		{"incomplete", 90, false, false, "UNAVAILABLE"},
		{"advisory", 90, true, true, "NOT_MATCHED"},
		{"unavailable score", -1, true, false, "UNAVAILABLE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			score := domain.StructuralCriticality{}
			if tc.score >= 0 {
				score.Score = &tc.score
			}
			check := domain.PackageCheck{PackageID: "x", Status: "CHECKED", CheckedAt: &now}
			out := Detect("x", score, check, tc.complete, tc.known)
			if out.Status != tc.status {
				t.Fatalf("%+v", out)
			}
			check.Advisories = []domain.Advisory{{ID: "OSV"}}
			if Detect("x", score, check, true, false).Status == "MATCH" {
				t.Fatal("known advisory incorrectly marked systemic")
			}
			check.Advisories = nil
			check.CheckedAt = nil
			if Detect("x", score, check, true, false).Status != "UNAVAILABLE" {
				t.Fatal("timestamp required")
			}
		})
	}
}
