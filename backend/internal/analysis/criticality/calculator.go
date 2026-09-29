// Package criticality calculates the approved Phase 4 structural-criticality
// model from scan-scoped dependency topology only.
package criticality

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"math"
	"sync"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/analysis/blast"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

const (
	ReverseFanInWeight          = 0.30
	ApplicationReachWeight      = 0.30
	SharedDependencyWeight      = 0.20
	IndependentPathWeight       = 0.20
	maximumCachedNormalizations = 16
)

type normalization struct {
	DirectDependents map[string]int
	MaximumFanIn     int
	PathCounts       blast.ScanPathCounts
	Applications     int
	Packages         int
}

// Cache keeps bounded scan-wide normalization facts local to the API process.
// The snapshot fingerprint prevents cross-scan or stale-topology reuse.
type Cache struct {
	mu     sync.Mutex
	values map[string]normalization
}

func (c *Cache) Score(ctx context.Context, scan string, snapshot domain.TopologySnapshot, impact domain.ImpactAnalysis, incomplete bool) domain.StructuralCriticality {
	n := normalizationFor(ctx, scan, snapshot, incomplete, c)
	return score(impact, n)
}

func normalizationFor(ctx context.Context, scan string, snapshot domain.TopologySnapshot, incomplete bool, cache *Cache) normalization {
	data, _ := json.Marshal(snapshot)
	sum := sha256.Sum256(data)
	key := scan + string(sum[:])
	if incomplete {
		key += ":incomplete"
	}
	cache.mu.Lock()
	if cache.values != nil {
		if value, ok := cache.values[key]; ok {
			cache.mu.Unlock()
			return value
		}
	}
	cache.mu.Unlock()

	n := build(ctx, snapshot, incomplete)
	if n.PathCounts.Reason == "" {
		cache.mu.Lock()
		if cache.values == nil {
			cache.values = map[string]normalization{}
		}
		if len(cache.values) >= maximumCachedNormalizations {
			cache.values = map[string]normalization{}
		}
		cache.values[key] = n
		cache.mu.Unlock()
	}
	return n
}

func build(ctx context.Context, snapshot domain.TopologySnapshot, incomplete bool) normalization {
	n := normalization{DirectDependents: map[string]int{}}
	for _, node := range snapshot.Nodes {
		switch node.Kind {
		case "application":
			n.Applications++
		case "package":
			n.Packages++
			n.DirectDependents[node.ID] = 0
		}
	}
	if incomplete {
		n.PathCounts.Reason = "INCOMPLETE_DEPENDENCY_EVIDENCE"
		return n
	}
	parents := map[string]map[string]bool{}
	for _, edge := range snapshot.Edges {
		if _, packageTarget := n.DirectDependents[edge.To]; !packageTarget {
			continue
		}
		if parents[edge.To] == nil {
			parents[edge.To] = map[string]bool{}
		}
		parents[edge.To][edge.From] = true
	}
	for id := range n.DirectDependents {
		n.DirectDependents[id] = len(parents[id])
		if n.DirectDependents[id] > n.MaximumFanIn {
			n.MaximumFanIn = n.DirectDependents[id]
		}
	}
	n.PathCounts = blast.CountPaths(ctx, snapshot)
	return n
}

func component(numerator, denominator, weight float64) domain.ScoreComponent {
	value := 0.0
	if denominator > 0 {
		value = numerator / denominator
	}
	return domain.ScoreComponent{Numerator: numerator, Denominator: denominator, Value: value, Weight: weight, Contribution: 100 * weight * value}
}

func score(impact domain.ImpactAnalysis, n normalization) domain.StructuralCriticality {
	out := domain.StructuralCriticality{TotalApplications: n.Applications, TotalPackages: n.Packages, ReasonCodes: []string{}}
	if n.PathCounts.Reason != "" {
		out.ReasonCodes = append(out.ReasonCodes, n.PathCounts.Reason)
		out.BlockingPackage = n.PathCounts.Blocking
		return out
	}
	totalNodesExceptSelected := n.Applications + n.Packages - 1
	paths := n.PathCounts.Counts[impact.PackageID]
	out.Components = map[string]domain.ScoreComponent{
		"reverse_fan_in":                  component(float64(n.DirectDependents[impact.PackageID]), float64(n.MaximumFanIn), ReverseFanInWeight),
		"application_reach":               component(float64(impact.ApplicationsExposed), float64(n.Applications), ApplicationReachWeight),
		"shared_dependency_concentration": component(float64(n.DirectDependents[impact.PackageID]), float64(totalNodesExceptSelected), SharedDependencyWeight),
		"independent_path_concentration":  component(math.Log1p(float64(paths)), math.Log1p(float64(n.PathCounts.Maximum)), IndependentPathWeight),
	}
	total := 0.0
	for _, value := range out.Components {
		total += value.Contribution
	}
	value := int(math.Round(math.Min(100, math.Max(0, total))))
	out.Score = &value
	return out
}
