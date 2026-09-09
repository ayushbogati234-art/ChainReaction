// Package resolver defines the ecosystem-neutral resolver contract.
package resolver

import "github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"

// Input contains the metadata files needed by a resolver. Resolvers must parse
// bytes as data only; they must never execute files from an analyzed project.
type Input struct {
	Pyproject  []byte
	PoetryLock []byte
	Path       string
}

// ResolvedApplication is the neutral graph representation emitted by an
// ecosystem resolver. Graph edges point from a dependent to its dependency.
type ResolvedApplication = domain.Graph

// Resolver converts supported dependency metadata into a neutral graph.
type Resolver interface {
	Resolve(Input) (ResolvedApplication, error)
}
