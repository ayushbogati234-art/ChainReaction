package neo4j

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Repository) WriteVulnerabilities(ctx context.Context, scan string, findings []domain.Finding) error {
	s := r.Driver.NewSession(ctx, neo.SessionConfig{AccessMode: neo.AccessModeWrite})
	defer s.Close(ctx)
	_, err := s.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		if _, err := tx.Run(ctx, "MATCH (v:Vulnerability {scan_id:$scan}) DETACH DELETE v", map[string]any{"scan": scan}); err != nil {
			return nil, err
		}
		for _, f := range findings {
			severity, cvss := []string{}, []string{}
			for _, a := range f.Advisories {
				for _, value := range a.Severity {
					if value.Type == "PROVIDER_LABEL" {
						severity = append(severity, value.Score)
					} else {
						cvss = append(cvss, value.Type+":"+value.Score)
					}
				}
			}
			if _, err := tx.Run(ctx, "MATCH (p:PackageVersion {scan_id:$scan,id:$package}) CREATE (v:Vulnerability {id:$id,scan_id:$scan,advisory_id:$advisory,source:'OSV',finding_id:$id,aliases:$aliases,severity:$severity,cvss:$cvss}) CREATE (p)-[:AFFECTED_BY {scan_id:$scan,confidence:$confidence}]->(v)", map[string]any{"scan": scan, "package": f.PackageID, "id": f.ID, "advisory": f.AdvisoryID, "confidence": string(f.MatchConfidence), "aliases": f.Aliases, "severity": severity, "cvss": cvss}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}
