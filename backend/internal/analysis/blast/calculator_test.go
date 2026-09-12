package blast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

func snapshot(ids map[string]string, pairs ...[2]string) domain.TopologySnapshot {
	nodes := []domain.TopologyNode{}
	for id, kind := range ids {
		nodes = append(nodes, domain.TopologyNode{ID: id, Kind: kind, Confidence: domain.ConfidenceHigh})
	}
	edges := []domain.TopologyEdge{}
	for _, pair := range pairs {
		edges = append(edges, domain.TopologyEdge{From: pair[0], To: pair[1], Confidence: domain.ConfidenceHigh})
	}
	return domain.TopologySnapshot{Nodes: nodes, Edges: edges}
}

func TestChainReachabilityAndDepth(t *testing.T) {
	s := snapshot(map[string]string{"app": "application", "a": "package", "b": "package", "x": "package"}, [2]string{"app", "a"}, [2]string{"a", "b"}, [2]string{"b", "x"})
	got, err := Analyze(s, "x", domain.PathOptions{})
	if err != nil || got.DownstreamPackages != 2 || got.ApplicationsExposed != 1 || !reflect.DeepEqual(got.ImmediateDependentPackages, []string{"b"}) || !reflect.DeepEqual(got.TransitiveDependentPackages, []string{"a"}) || got.Paths.TotalPaths != 1 || got.Paths.MaximumDepth != 3 || !got.Paths.MaximumDepthExact {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestDiamondAndApplicationFilter(t *testing.T) {
	s := snapshot(map[string]string{"app1": "application", "app2": "application", "a": "package", "b": "package", "x": "package"}, [2]string{"app1", "a"}, [2]string{"app1", "b"}, [2]string{"app2", "a"}, [2]string{"a", "x"}, [2]string{"b", "x"})
	got, _ := Analyze(s, "x", domain.PathOptions{ApplicationID: "app1"})
	if got.ApplicationsExposed != 2 || got.Paths.TotalPaths != 2 || len(got.Paths.Paths) != 2 {
		t.Fatalf("%+v", got)
	}
	filtered, _ := Analyze(s, "x", domain.PathOptions{ApplicationID: "app2"})
	if filtered.Paths.TotalPaths != 1 || filtered.Paths.Paths[0][0] != "app2" {
		t.Fatalf("%+v", filtered)
	}
}

func TestCycleDisconnectedAndMultipleVersions(t *testing.T) {
	s := snapshot(map[string]string{"app": "application", "a@1": "package", "a@2": "package", "x": "package", "alone": "package"}, [2]string{"app", "a@1"}, [2]string{"a@1", "x"}, [2]string{"x", "a@1"}, [2]string{"app", "a@2"})
	got, _ := Analyze(s, "x", domain.PathOptions{})
	if !got.Paths.CyclesDetected || got.DownstreamPackages != 1 || got.Paths.TotalPaths != 1 {
		t.Fatalf("%+v", got)
	}
	alone, _ := Analyze(s, "alone", domain.PathOptions{})
	if alone.DownstreamPackages != 0 || alone.ApplicationsExposed != 0 || alone.Paths.TotalPaths != 0 || !alone.Paths.CountExact {
		t.Fatalf("%+v", alone)
	}
	if _, err := Analyze(s, "a@3", domain.PathOptions{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestPathCapsAndIncompleteDepth(t *testing.T) {
	s := snapshot(map[string]string{"app": "application", "a": "package", "b": "package", "x": "package"}, [2]string{"app", "a"}, [2]string{"app", "b"}, [2]string{"a", "x"}, [2]string{"b", "x"})
	capResult, _ := Analyze(s, "x", domain.PathOptions{MaxPaths: 1, CountLimit: 1})
	if !capResult.Paths.CountCapped || capResult.Paths.CountExact || !capResult.Paths.Truncated || len(capResult.Paths.Paths) != 1 {
		t.Fatalf("%+v", capResult.Paths)
	}
	displayResult, _ := Analyze(s, "x", domain.PathOptions{MaxPaths: 1})
	if !displayResult.Paths.CountExact || displayResult.Paths.CountCapped || !displayResult.Paths.Truncated || displayResult.Paths.TotalPaths != 2 || len(displayResult.Paths.Paths) != 1 {
		t.Fatalf("display and exact aggregate were conflated: %+v", displayResult.Paths)
	}
	depthResult, _ := Analyze(s, "x", domain.PathOptions{MaxDepth: 1})
	if !depthResult.Paths.DepthLimited || depthResult.Paths.CountExact || depthResult.Paths.MaximumDepthExact {
		t.Fatalf("%+v", depthResult.Paths)
	}
}

func TestConfidenceUsesWeakestImpactEvidence(t *testing.T) {
	s := snapshot(map[string]string{"app": "application", "x": "package"}, [2]string{"app", "x"})
	s.Edges[0].Confidence = domain.ConfidenceMedium
	got, _ := Analyze(s, "x", domain.PathOptions{})
	if got.EvidenceConfidence != domain.ConfidenceMedium || got.Paths.EvidenceConfidence != domain.ConfidenceMedium {
		t.Fatalf("%+v", got)
	}
}
