package graph

import "github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"

// Snapshot is deliberately storage-neutral. Analysis packages never receive
// Neo4j records, labels or driver types.
type Snapshot = domain.TopologySnapshot
