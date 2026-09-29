package criticality

import (
	"context"
	"testing"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/analysis/blast"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

func snapshot(ids map[string]string, edges ...[2]string) domain.TopologySnapshot {
	out := domain.TopologySnapshot{}
	for id, kind := range ids {
		out.Nodes = append(out.Nodes, domain.TopologyNode{ID: id, Kind: kind, Confidence: domain.ConfidenceHigh})
	}
	for _, edge := range edges {
		out.Edges = append(out.Edges, domain.TopologyEdge{From: edge[0], To: edge[1], Confidence: domain.ConfidenceHigh})
	}
	return out
}

func impact(t *testing.T, s domain.TopologySnapshot, id string) domain.ImpactAnalysis {
	t.Helper()
	value, err := blast.Analyze(s, id, domain.PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestApprovedFormulaAndEvidence(t *testing.T) {
	s := snapshot(map[string]string{"app": "application", "a": "package", "b": "package", "x": "package"}, [2]string{"app", "a"}, [2]string{"app", "b"}, [2]string{"a", "x"}, [2]string{"b", "x"})
	cache := &Cache{}
	got := cache.Score(context.Background(), "scan", s, impact(t, s, "x"), false)
	if got.Score == nil || *got.Score != 93 {
		t.Fatalf("%+v", got)
	}
	if got.Components["reverse_fan_in"].Contribution != 30 || got.Components["application_reach"].Contribution != 30 || got.Components["independent_path_concentration"].Contribution != 20 {
		t.Fatalf("missing formula evidence: %+v", got.Components)
	}
	parent := cache.Score(context.Background(), "scan", s, impact(t, s, "a"), false)
	if parent.Score == nil || *parent.Score != 64 {
		t.Fatalf("%+v", parent)
	}
}

func TestUnavailableEvidenceAndCacheIsolation(t *testing.T) {
	s := snapshot(map[string]string{"app": "application", "x": "package"}, [2]string{"app", "x"})
	cache := &Cache{}
	got := cache.Score(context.Background(), "scan", s, impact(t, s, "x"), true)
	if got.Score != nil || len(got.ReasonCodes) != 1 || got.ReasonCodes[0] != "INCOMPLETE_DEPENDENCY_EVIDENCE" {
		t.Fatalf("%+v", got)
	}
	complete := cache.Score(context.Background(), "scan", s, impact(t, s, "x"), false)
	if complete.Score == nil || *complete.Score != 100 {
		t.Fatalf("%+v", complete)
	}
	s.Edges = nil
	changed := cache.Score(context.Background(), "scan", s, impact(t, s, "x"), false)
	if changed.Score == nil || *changed.Score != 0 {
		t.Fatalf("stale cache: %+v", changed)
	}
}
