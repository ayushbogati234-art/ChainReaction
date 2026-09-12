// Package blast calculates scan-relative structural impact evidence and the
// approved prototype Blast Radius model.
package blast

import (
	"fmt"
	"sort"
	"time"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

const (
	MaxSnapshotNodes      = 10000
	MaxSnapshotEdges      = 100000
	DefaultMaxPaths       = 20
	MaximumDisplayedPaths = 100
	DefaultMaxDepth       = 100
	MaximumPathCount      = 1000
)

func Analyze(snapshot domain.TopologySnapshot, target string, options domain.PathOptions) (domain.ImpactAnalysis, error) {
	if len(snapshot.Nodes) > MaxSnapshotNodes || len(snapshot.Edges) > MaxSnapshotEdges {
		return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: topology exceeds analysis safety limits")
	}
	if options.MaxPaths == 0 {
		options.MaxPaths = DefaultMaxPaths
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = DefaultMaxDepth
	}
	if options.CountLimit == 0 {
		options.CountLimit = MaximumPathCount
	}
	if options.MaxPaths < 1 || options.MaxPaths > MaximumDisplayedPaths {
		return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: max_paths must be between 1 and 100")
	}
	if options.MaxDepth < 1 || options.MaxDepth > DefaultMaxDepth {
		return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: max_depth must be between 1 and 100")
	}
	if options.CountLimit < options.MaxPaths || options.CountLimit > MaximumPathCount {
		return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: invalid aggregate path count limit")
	}

	nodes := make(map[string]domain.TopologyNode, len(snapshot.Nodes))
	apps := []string{}
	for _, node := range snapshot.Nodes {
		if _, duplicate := nodes[node.ID]; duplicate {
			return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: topology contains duplicate node identities")
		}
		nodes[node.ID] = node
		if node.Kind == "application" {
			apps = append(apps, node.ID)
		}
	}
	targetNode, ok := nodes[target]
	if !ok || targetNode.Kind != "package" {
		return domain.ImpactAnalysis{}, domain.ErrNotFound
	}
	sort.Strings(apps)
	for _, edge := range snapshot.Edges {
		if _, ok := nodes[edge.From]; !ok {
			return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: topology edge has an unknown source")
		}
		if _, ok := nodes[edge.To]; !ok {
			return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: topology edge has an unknown target")
		}
	}
	if options.ApplicationID != "" {
		node, exists := nodes[options.ApplicationID]
		if !exists || node.Kind != "application" {
			return domain.ImpactAnalysis{}, fmt.Errorf("INVALID_REQUEST: application_id is not an application in this scan")
		}
	}

	forward, reverse := adjacency(snapshot.Edges)
	reachable := map[string]bool{target: true}
	queue := []string{target}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, dependent := range reverse[current] {
			if !reachable[dependent] {
				reachable[dependent] = true
				queue = append(queue, dependent)
			}
		}
	}
	immediate, transitive, exposed := []string{}, []string{}, []string{}
	immediateSet := map[string]bool{}
	for _, id := range reverse[target] {
		if id != target && nodes[id].Kind == "package" {
			immediate = append(immediate, id)
			immediateSet[id] = true
		}
	}
	for id := range reachable {
		if id == target {
			continue
		}
		switch nodes[id].Kind {
		case "application":
			exposed = append(exposed, id)
		case "package":
			if !immediateSet[id] {
				transitive = append(transitive, id)
			}
		}
	}
	sort.Strings(immediate)
	sort.Strings(transitive)
	sort.Strings(exposed)
	affected := append([]string{target}, immediate...)
	affected = append(affected, transitive...)
	affected = append(affected, exposed...)
	sort.Strings(affected)
	confidence := normalize(targetNode.Confidence)
	for id := range reachable {
		confidence = lower(confidence, nodes[id].Confidence)
	}
	for _, edge := range snapshot.Edges {
		if reachable[edge.From] && reachable[edge.To] {
			confidence = lower(confidence, edge.Confidence)
		}
	}

	paths := enumeratePaths(nodes, forward, exposed, reachable, target, options)
	paths.CyclesDetected = paths.CyclesDetected || hasCycle(forward, reachable)
	if paths.CyclesDetected && !contains(paths.Limitations, "Cycles were encountered and only simple paths were counted.") {
		paths.Limitations = append(paths.Limitations, "Cycles were encountered and only simple paths were counted.")
	}
	paths.EvidenceConfidence = confidence
	paths.EvidenceComplete = true
	result := domain.ImpactAnalysis{
		Limitations: []string{},
		PackageID:   target, DownstreamPackages: len(immediate) + len(transitive),
		ImmediateDependentPackages: immediate, TransitiveDependentPackages: transitive,
		ApplicationsExposed: len(exposed), ApplicationIDs: exposed, AffectedNodeIDs: affected,
		ReachabilityExact: true, EvidenceConfidence: confidence, EvidenceComplete: true, Paths: paths,
	}
	return result, nil
}

func adjacency(edges []domain.TopologyEdge) (map[string][]string, map[string][]string) {
	forward, reverse := map[string][]string{}, map[string][]string{}
	seen := map[string]bool{}
	for _, edge := range edges {
		key := edge.From + "\x00" + edge.To
		if seen[key] {
			continue
		}
		seen[key] = true
		forward[edge.From] = append(forward[edge.From], edge.To)
		reverse[edge.To] = append(reverse[edge.To], edge.From)
	}
	for key := range forward {
		sort.Strings(forward[key])
	}
	for key := range reverse {
		sort.Strings(reverse[key])
	}
	return forward, reverse
}

func enumeratePaths(nodes map[string]domain.TopologyNode, forward map[string][]string, apps []string, eligible map[string]bool, target string, options domain.PathOptions) domain.PathResult {
	selectedApps := apps
	if options.ApplicationID != "" {
		selectedApps = []string{options.ApplicationID}
	}
	out := domain.PathResult{Limitations: []string{}, Paths: [][]string{}, PathLengths: []int{}, Applications: selectedApps, Limit: options.MaxPaths, CountLimit: options.CountLimit, MaxDepthLimit: options.MaxDepth, EvidenceComplete: true}
	stopped := false
	work := 0
	workLimited := false
	deadline := time.Now().Add(ExecutionLimit)
	var walk func(string, []string, map[string]bool)
	walk = func(current string, path []string, seen map[string]bool) {
		if stopped {
			return
		}
		work++
		if work > MaximumWork || time.Now().After(deadline) {
			stopped = true
			workLimited = true
			return
		}
		if current == target {
			if out.TotalPaths == options.CountLimit {
				out.CountCapped = true
				stopped = true
				return
			}
			out.TotalPaths++
			depth := len(path) - 1
			if depth > out.MaximumDepth {
				out.MaximumDepth = depth
			}
			if len(out.Paths) < options.MaxPaths {
				out.Paths = append(out.Paths, append([]string(nil), path...))
				out.PathLengths = append(out.PathLengths, depth)
			}
			return
		}
		if len(path)-1 >= options.MaxDepth {
			for _, next := range forward[current] {
				if eligible[next] && !seen[next] {
					out.DepthLimited = true
					break
				}
			}
			return
		}
		for _, next := range forward[current] {
			work++
			if work > MaximumWork || time.Now().After(deadline) {
				stopped = true
				workLimited = true
				return
			}
			if !eligible[next] {
				continue
			}
			if _, valid := nodes[next]; !valid {
				continue
			}
			if seen[next] {
				out.CyclesDetected = true
				continue
			}
			seen[next] = true
			walk(next, append(path, next), seen)
			delete(seen, next)
			if stopped {
				return
			}
		}
	}
	for _, app := range selectedApps {
		walk(app, []string{app}, map[string]bool{app: true})
		if stopped {
			break
		}
	}
	out.CountExact = !out.CountCapped && !out.DepthLimited && !workLimited
	out.MaximumDepthExact = out.CountExact
	out.Truncated = workLimited || out.CountCapped || out.DepthLimited || out.TotalPaths > len(out.Paths)
	if workLimited {
		out.Limitations = append(out.Limitations, "Path traversal reached its execution/work limit; counts and depth are bounded observations.")
	}
	if out.CountCapped {
		out.Limitations = append(out.Limitations, fmt.Sprintf("Aggregate path count reached the %d safety cap; total_paths is a lower bound.", options.CountLimit))
	}
	if out.DepthLimited {
		out.Limitations = append(out.Limitations, fmt.Sprintf("Traversal reached max_depth=%d; paths and maximum depth are bounded observations.", options.MaxDepth))
	}
	if out.CyclesDetected {
		out.Limitations = append(out.Limitations, "Cycles were encountered and only simple paths were counted.")
	}
	return out
}

func lower(a, b domain.Confidence) domain.Confidence {
	a, b = normalize(a), normalize(b)
	rank := map[domain.Confidence]int{domain.ConfidenceLow: 0, domain.ConfidenceMedium: 1, domain.ConfidenceHigh: 2}
	if rank[b] < rank[a] {
		return b
	}
	return a
}

func normalize(value domain.Confidence) domain.Confidence {
	switch value {
	case domain.ConfidenceHigh, domain.ConfidenceMedium, domain.ConfidenceLow:
		return value
	}
	return domain.ConfidenceLow
}

func hasCycle(forward map[string][]string, included map[string]bool) bool {
	state := map[string]uint8{}
	var visit func(string) bool
	visit = func(id string) bool {
		if state[id] == 1 {
			return true
		}
		if state[id] == 2 {
			return false
		}
		state[id] = 1
		for _, next := range forward[id] {
			if included[next] && visit(next) {
				return true
			}
		}
		state[id] = 2
		return false
	}
	for id := range included {
		if visit(id) {
			return true
		}
	}
	return false
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
