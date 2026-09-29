package evidence

import (
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"testing"
	"time"
)

func TestCoverageMustMatchWholeSnapshot(t *testing.T) {
	now := time.Now()
	snapshot := domain.TopologySnapshot{Nodes: []domain.TopologyNode{{ID: "app", Kind: "application"}, {ID: "a", Kind: "package"}, {ID: "b", Kind: "package"}}}
	for _, mode := range []string{"complete", "missing", "duplicate", "foreign", "failed", "timestamp", "provider", "partial", "graph"} {
		t.Run(mode, func(t *testing.T) {
			report := domain.Enrichment{Status: "COMPLETE", Provider: "OSV", GraphStatus: "COMPLETE", Checks: []domain.PackageCheck{{PackageID: "a", Status: "CHECKED", CheckedAt: &now}, {PackageID: "b", Status: "CHECKED", CheckedAt: &now}}}
			switch mode {
			case "missing":
				report.Checks = report.Checks[:1]
			case "duplicate":
				report.Checks[1].PackageID = "a"
			case "foreign":
				report.Checks[1].PackageID = "other"
			case "failed":
				report.Checks[1].Status = "FAILED"
			case "timestamp":
				report.Checks[1].CheckedAt = nil
			case "provider":
				report.Provider = ""
			case "partial":
				report.Status = "PARTIAL"
			case "graph":
				report.GraphStatus = "DEGRADED"
			}
			_, complete := Coverage(snapshot, report)
			if complete != (mode == "complete") {
				t.Fatalf("unexpected completeness %v", complete)
			}
		})
	}
}
