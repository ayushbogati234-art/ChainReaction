package graph

import "github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"

// EnumerateSimplePaths is used by tests and bounded in-memory callers. It never revisits a node.
func EnumerateSimplePaths(edges []domain.Edge, source, target string, limit int) domain.PathResult {
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	out := domain.PathResult{Limit: limit}
	var walk func(string, []string, map[string]bool)
	walk = func(n string, p []string, seen map[string]bool) {
		if out.Truncated {
			return
		}
		if n == target {
			if len(out.Paths) == limit {
				out.Truncated = true
				return
			}
			out.Paths = append(out.Paths, append([]string(nil), p...))
			return
		}
		for _, next := range adj[n] {
			if !seen[next] {
				seen[next] = true
				walk(next, append(p, next), seen)
				delete(seen, next)
			}
		}
	}
	walk(source, []string{source}, map[string]bool{source: true})
	return out
}
