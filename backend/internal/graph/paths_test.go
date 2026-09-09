package graph

import (
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"testing"
)

func TestCycleSafeAndCapped(t *testing.T) {
	e := []domain.Edge{{From: "app", To: "a"}, {From: "a", To: "b"}, {From: "b", To: "a"}, {From: "a", To: "target"}, {From: "b", To: "target"}}
	r := EnumerateSimplePaths(e, "app", "target", 1)
	if len(r.Paths) != 1 || !r.Truncated {
		t.Fatalf("want one truncated path, got %#v", r)
	}
}

func TestDiamondAndSharedPaths(t *testing.T) {
	diamond := []domain.Edge{{From: "app", To: "a"}, {From: "app", To: "b"}, {From: "a", To: "shared"}, {From: "b", To: "shared"}}
	if got := EnumerateSimplePaths(diamond, "app", "shared", 10); len(got.Paths) != 2 || got.Truncated {
		t.Fatalf("diamond paths: %#v", got)
	}
	shared := []domain.Edge{{From: "app-one", To: "shared"}, {From: "app-two", To: "shared"}}
	if got := EnumerateSimplePaths(shared, "app-one", "shared", 10); len(got.Paths) != 1 {
		t.Fatalf("shared dependency path: %#v", got)
	}
}
