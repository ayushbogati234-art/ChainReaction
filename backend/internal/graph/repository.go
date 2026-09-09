package graph

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

// Repository is the graph boundary used by application services. Cypher and
// Neo4j driver details stay in graph/neo4j.
type Repository interface {
	Write(context.Context, string, domain.Graph) error
	Graph(context.Context, string) (map[string]any, error)
	Package(context.Context, string, string) (map[string]any, error)
	Paths(context.Context, string, string, int) (domain.PathResult, error)
}
