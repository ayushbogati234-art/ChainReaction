package neo4j

import (
	"context"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/graph"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Repository struct{ Driver neo.DriverWithContext }

var _ graph.Repository = Repository{}

func (r Repository) Write(ctx context.Context, scanID string, g domain.Graph) error {
	s := r.Driver.NewSession(ctx, neo.SessionConfig{AccessMode: neo.AccessModeWrite})
	defer s.Close(ctx)
	_, e := s.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		_, e := tx.Run(ctx, "MATCH (n {scan_id:$scan}) DETACH DELETE n", map[string]any{"scan": scanID})
		if e != nil {
			return nil, e
		}
		_, e = tx.Run(ctx, "CREATE (:Application {id:$id, scan_id:$scan, name:$name, path:'.', ecosystem:'Python', source:$source, confidence:$confidence, input_files:$input_files})", map[string]any{"id": g.Application.ID, "scan": scanID, "name": g.Application.Name, "source": g.Application.Source, "confidence": string(g.Application.Confidence), "input_files": g.InputFiles})
		if e != nil {
			return nil, e
		}
		for _, p := range g.Packages {
			_, e = tx.Run(ctx, "CREATE (:PackageVersion {id:$id, scan_id:$scan, name:$name, version:$version, purl:$purl, direct:$direct, source:$source, groups:$groups, confidence:$confidence, limitations:$limitations,marker:$marker,python_versions:$python_versions,extras:$extras,optional:$optional})", map[string]any{"id": p.ID, "scan": scanID, "name": p.Name, "version": p.Version, "purl": p.PURL, "direct": p.Direct, "source": p.Source, "groups": p.Groups, "confidence": string(p.Confidence), "limitations": p.Limitations, "marker": p.Marker, "python_versions": p.PythonVersions, "extras": p.Extras, "optional": p.Optional})
			if e != nil {
				return nil, e
			}
		}
		for _, edge := range g.Edges {
			_, e = tx.Run(ctx, "MATCH (a {scan_id:$scan,id:$from}),(b {scan_id:$scan,id:$to}) CREATE (a)-[:DEPENDS_ON {direct:$direct,marker:$marker,extras:$extras,group:$group,source:$source,confidence:$confidence,constraint:$constraint,optional:$optional}]->(b)", map[string]any{"scan": scanID, "from": edge.From, "to": edge.To, "direct": edge.Direct, "marker": edge.Marker, "extras": edge.Extras, "group": edge.Group, "source": edge.Source, "confidence": edge.Confidence, "constraint": edge.Constraint, "optional": edge.Optional})
			if e != nil {
				return nil, e
			}
		}
		return nil, nil
	})
	return e
}
func (r Repository) Graph(ctx context.Context, scan string) (map[string]any, error) {
	return r.records(ctx, "MATCH (n {scan_id:$scan}) WHERE n:Application OR n:PackageVersion OPTIONAL MATCH (n)-[e:DEPENDS_ON]->(m {scan_id:$scan}) RETURN collect(DISTINCT n{.*}) AS nodes, collect(DISTINCT CASE WHEN e IS NOT NULL THEN {source:n.id,target:m.id,direct:e.direct,marker:e.marker,extras:e.extras,group:e.group,source_evidence:e.source,confidence:e.confidence,constraint:e.constraint,optional:e.optional} END) AS edges", scan)
}
func (r Repository) Package(ctx context.Context, scan, id string) (map[string]any, error) {
	s := r.Driver.NewSession(ctx, neo.SessionConfig{AccessMode: neo.AccessModeRead})
	defer s.Close(ctx)
	res, e := s.Run(ctx, "MATCH (p:PackageVersion {scan_id:$scan,id:$id}) OPTIONAL MATCH (p)-[out:DEPENDS_ON]->(d {scan_id:$scan}) OPTIONAL MATCH (u {scan_id:$scan})-[in:DEPENDS_ON]->(p) RETURN p{.*} AS package, collect(DISTINCT d{.*}) AS dependencies, collect(DISTINCT u{.*}) AS dependents", map[string]any{"scan": scan, "id": id})
	if e != nil {
		return nil, e
	}
	if !res.Next(ctx) {
		if res.Err() != nil {
			return nil, res.Err()
		}
		return nil, domain.ErrNotFound
	}
	return res.Record().AsMap(), res.Err()
}
func (r Repository) Paths(ctx context.Context, scan, id string, limit int) (domain.PathResult, error) {
	if limit < 1 || limit > 100 {
		return domain.PathResult{}, fmt.Errorf("INVALID_REQUEST: path limit must be between 1 and 100")
	}
	s := r.Driver.NewSession(ctx, neo.SessionConfig{AccessMode: neo.AccessModeRead})
	defer s.Close(ctx)
	q := "MATCH p=(a:Application {scan_id:$scan})-[:DEPENDS_ON*1..]->(t:PackageVersion {scan_id:$scan,id:$id}) WHERE all(n IN nodes(p) WHERE n.scan_id=$scan AND single(m IN nodes(p) WHERE m=n)) RETURN [n IN nodes(p) | n.id] AS path LIMIT $limit"
	res, e := s.Run(ctx, q, map[string]any{"scan": scan, "id": id, "limit": limit + 1})
	if e != nil {
		return domain.PathResult{}, e
	}
	out := domain.PathResult{Limit: limit, Paths: [][]string{}}
	for res.Next(ctx) {
		v := res.Record().Values[0].([]any)
		p := make([]string, len(v))
		for i, x := range v {
			p[i] = x.(string)
		}
		if len(out.Paths) == limit {
			out.Truncated = true
			break
		}
		out.Paths = append(out.Paths, p)
	}
	return out, res.Err()
}
func (r Repository) records(ctx context.Context, q, scan string) (map[string]any, error) {
	s := r.Driver.NewSession(ctx, neo.SessionConfig{AccessMode: neo.AccessModeRead})
	defer s.Close(ctx)
	res, e := s.Run(ctx, q, map[string]any{"scan": scan})
	if e != nil {
		return nil, e
	}
	if !res.Next(ctx) {
		return map[string]any{"nodes": []any{}, "edges": []any{}}, res.Err()
	}
	return res.Record().AsMap(), res.Err()
}
