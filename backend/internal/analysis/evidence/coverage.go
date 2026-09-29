// Package evidence validates persisted provider coverage against scan topology.
package evidence

import "github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"

// Coverage does not infer successful checks from an empty findings list.
func Coverage(snapshot domain.TopologySnapshot, report domain.Enrichment) (map[string]domain.PackageCheck, bool) {
	checks := map[string]domain.PackageCheck{}
	complete := report.Status == "COMPLETE" && report.Provider == "OSV" && report.GraphStatus == "COMPLETE"
	for _, check := range report.Checks {
		if _, duplicate := checks[check.PackageID]; duplicate {
			complete = false
		}
		checks[check.PackageID] = check
		if check.Status != "CHECKED" || check.CheckedAt == nil || check.CheckedAt.IsZero() {
			complete = false
		}
	}
	packages := 0
	for _, node := range snapshot.Nodes {
		if node.Kind != "package" {
			continue
		}
		packages++
		if _, exists := checks[node.ID]; !exists {
			complete = false
		}
	}
	return checks, complete && packages == len(checks)
}
