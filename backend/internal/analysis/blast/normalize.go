package blast

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

const ApplicationWeight = 0.50
const DownstreamWeight = 0.30
const PathWeight = 0.20
const MaximumWork = 200000
const ExecutionLimit = 2 * time.Second

type normalization struct {
	Counts                          map[string]int
	Maximum, Applications, Packages int
	Reason, Blocking                string
}

// ScanPathCounts exposes bounded, exact-or-unavailable scan-wide simple path
// counts for other topology-only analysis packages. It never returns a lower
// bound as an exact count.
type ScanPathCounts struct {
	Counts   map[string]int
	Maximum  int
	Reason   string
	Blocking string
}

func CountPaths(ctx context.Context, snapshot domain.TopologySnapshot) ScanPathCounts {
	n := normalizeScan(ctx, snapshot, MaximumWork, false)
	counts := make(map[string]int, len(n.Counts))
	for id, count := range n.Counts {
		counts[id] = count
	}
	return ScanPathCounts{Counts: counts, Maximum: n.Maximum, Reason: n.Reason, Blocking: n.Blocking}
}

// Cache is bounded and scoped by scan identity and a fingerprint of its full
// snapshot. Completed immutable snapshots reuse normalization across clicks.
type Cache struct {
	once   sync.Once
	gate   chan struct{}
	values map[string]normalization
}

func (c *Cache) Score(ctx context.Context, scan string, snapshot domain.TopologySnapshot, impact domain.ImpactAnalysis, incomplete bool) domain.BlastRadius {
	data, _ := json.Marshal(snapshot)
	key := scan + stringHash(data)
	if incomplete {
		key += ":incomplete"
	}
	c.once.Do(func() { c.gate = make(chan struct{}, 1) })
	timer := time.NewTimer(ExecutionLimit)
	defer timer.Stop()
	select {
	case c.gate <- struct{}{}:
		defer func() { <-c.gate }()
	case <-ctx.Done():
		n := normalizeScan(ctx, snapshot, 0, true)
		n.Reason = "EXECUTION_LIMIT"
		return score(impact, n)
	case <-timer.C:
		n := normalizeScan(ctx, snapshot, 0, true)
		n.Reason = "EXECUTION_LIMIT"
		return score(impact, n)
	}
	if c.values == nil {
		c.values = map[string]normalization{}
	}
	n, ok := c.values[key]
	if !ok {
		n = normalizeScan(ctx, snapshot, MaximumWork, incomplete)
		if n.Reason != "EXECUTION_LIMIT" {
			if len(c.values) >= 16 {
				c.values = map[string]normalization{}
			}
			c.values[key] = n
		}
	}
	// Completeness is part of availability even when topology has not changed.
	if incomplete {
		n.Reason = "INCOMPLETE_DEPENDENCY_EVIDENCE"
	}
	return score(impact, n)
}

func stringHash(data []byte) string { sum := sha256.Sum256(data); return string(sum[:]) }

func normalizeScan(ctx context.Context, s domain.TopologySnapshot, budget int, incomplete bool) normalization {
	n := normalization{Counts: map[string]int{}}
	kinds := map[string]string{}
	apps := []string{}
	for _, node := range s.Nodes {
		kinds[node.ID] = node.Kind
		if node.Kind == "application" {
			n.Applications++
			apps = append(apps, node.ID)
		}
		if node.Kind == "package" {
			n.Packages++
			n.Counts[node.ID] = 0
		}
	}
	if incomplete {
		n.Reason = "INCOMPLETE_DEPENDENCY_EVIDENCE"
		return n
	}
	forward, _ := adjacency(s.Edges)
	deadline := time.Now().Add(ExecutionLimit)
	work := 0
	var walk func(string, int, map[string]bool)
	walk = func(id string, depth int, seen map[string]bool) {
		if n.Reason != "" {
			return
		}
		work++
		if work > budget {
			n.Reason = "WORK_LIMIT"
			return
		}
		if ctx.Err() != nil || time.Now().After(deadline) {
			n.Reason = "EXECUTION_LIMIT"
			return
		}
		if kinds[id] == "package" {
			n.Counts[id]++
			if n.Counts[id] > MaximumPathCount {
				n.Reason = "SCAN_PATH_COUNT_CAPPED"
				n.Blocking = id
				return
			}
			if n.Counts[id] > n.Maximum {
				n.Maximum = n.Counts[id]
			}
		}
		for _, next := range forward[id] {
			work++
			if work > budget {
				n.Reason = "WORK_LIMIT"
				return
			}
			if seen[next] {
				continue
			}
			if depth >= DefaultMaxDepth {
				n.Reason = "SCAN_PATH_DEPTH_LIMIT"
				n.Blocking = next
				return
			}
			seen[next] = true
			walk(next, depth+1, seen)
			delete(seen, next)
			if n.Reason != "" {
				return
			}
		}
	}
	for _, app := range apps {
		walk(app, 0, map[string]bool{app: true})
		if n.Reason != "" {
			break
		}
	}
	return n
}

func component(num, den, weight float64) domain.ScoreComponent {
	value := 0.0
	if den > 0 {
		value = num / den
	}
	return domain.ScoreComponent{Numerator: num, Denominator: den, Value: value, Weight: weight, Contribution: 100 * weight * value}
}

func score(i domain.ImpactAnalysis, n normalization) domain.BlastRadius {
	out := domain.BlastRadius{TotalApplications: n.Applications, TotalPackages: n.Packages, ReasonCodes: []string{}, BlockingPackage: n.Blocking}
	if n.Reason != "" {
		out.ReasonCodes = append(out.ReasonCodes, n.Reason)
		return out
	}
	out.Components = map[string]domain.ScoreComponent{
		"application_reach": component(float64(i.ApplicationsExposed), float64(n.Applications), ApplicationWeight),
		"downstream_reach":  component(float64(i.DownstreamPackages), math.Max(0, float64(n.Packages-1)), DownstreamWeight),
		"path_multiplicity": component(math.Log1p(float64(n.Counts[i.PackageID])), math.Log1p(float64(n.Maximum)), PathWeight),
	}
	total := 0.0
	for _, c := range out.Components {
		total += c.Contribution
	}
	value := int(math.Round(math.Min(100, math.Max(0, total))))
	label := classification(value)
	out.Score = &value
	out.Classification = &label
	return out
}

func classification(value int) string {
	switch {
	case value < 20:
		return "LOW"
	case value < 40:
		return "MODERATE"
	case value < 60:
		return "ELEVATED"
	case value < 80:
		return "HIGH"
	default:
		return "VERY HIGH"
	}
}
