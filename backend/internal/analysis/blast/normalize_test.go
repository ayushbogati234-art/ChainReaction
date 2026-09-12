package blast

import (
	"context"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"testing"
)

func TestApprovedExamples(t *testing.T) {
	cases := []struct {
		name           string
		apps, packages []string
		edges          [][2]string
		target         string
		want           int
	}{
		{"single", []string{"app"}, []string{"x"}, [][2]string{{"app", "x"}}, "x", 70},
		{"chain", []string{"app"}, []string{"a", "b", "x"}, [][2]string{{"app", "a"}, {"a", "b"}, {"b", "x"}}, "x", 100},
		{"chain middle", []string{"app"}, []string{"a", "b", "x"}, [][2]string{{"app", "a"}, {"a", "b"}, {"b", "x"}}, "b", 85},
		{"diamond", []string{"app"}, []string{"a", "b", "x"}, [][2]string{{"app", "a"}, {"app", "b"}, {"a", "x"}, {"b", "x"}}, "x", 100},
		{"shared", []string{"one", "two"}, []string{"a", "x"}, [][2]string{{"one", "a"}, {"two", "a"}, {"a", "x"}}, "x", 100},
		{"isolated", []string{"app"}, []string{"a", "x"}, [][2]string{{"app", "a"}}, "x", 0},
		{"no apps", nil, []string{"x"}, nil, "x", 0},
		{"cycle versions", []string{"app"}, []string{"x@1", "x@2"}, [][2]string{{"app", "x@1"}, {"x@1", "x@2"}, {"x@2", "x@1"}}, "x@2", 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ids := map[string]string{}
			for _, id := range c.apps {
				ids[id] = "application"
			}
			for _, id := range c.packages {
				ids[id] = "package"
			}
			s := snapshot(ids, c.edges...)
			i, e := Analyze(s, c.target, domain.PathOptions{})
			if e != nil {
				t.Fatal(e)
			}
			cache := &Cache{}
			r := cache.Score(context.Background(), "scan", s, i, false)
			if r.Score == nil || *r.Score != c.want {
				t.Fatalf("%+v want %d", r, c.want)
			}
			filtered, _ := Analyze(s, c.target, domain.PathOptions{MaxDepth: 1, MaxPaths: 1, ApplicationID: func() string {
				if len(c.apps) > 0 {
					return c.apps[0]
				}
				return ""
			}()})
			r2 := cache.Score(context.Background(), "scan", s, filtered, false)
			if *r2.Score != *r.Score || len(cache.values) != 1 {
				t.Fatal("filter-dependent score or cache miss")
			}
			unavailable := cache.Score(context.Background(), "scan", s, i, true)
			if unavailable.Score != nil || unavailable.ReasonCodes[0] != "INCOMPLETE_DEPENDENCY_EVIDENCE" {
				t.Fatal(unavailable)
			}
		})
	}
}

func TestLimitsAndEmpty(t *testing.T) {
	n := normalizeScan(context.Background(), domain.TopologySnapshot{}, MaximumWork, false)
	if n.Maximum != 0 || n.Reason != "" {
		t.Fatal(n)
	}
	ids := map[string]string{"app": "application", "selected": "package"}
	edges := [][2]string{{"app", "selected"}}
	prev := []string{"app"}
	for level := 0; level < 11; level++ {
		next := []string{fmt.Sprintf("a%d", level), fmt.Sprintf("b%d", level)}
		for _, id := range next {
			ids[id] = "package"
			for _, p := range prev {
				edges = append(edges, [2]string{p, id})
			}
		}
		prev = next
	}
	s := snapshot(ids, edges...)
	n = normalizeScan(context.Background(), s, MaximumWork, false)
	if n.Reason != "SCAN_PATH_COUNT_CAPPED" || n.Blocking == "selected" {
		t.Fatal(n.Reason, n.Blocking)
	}
	r := score(domain.ImpactAnalysis{PackageID: "selected"}, n)
	if r.Score != nil || r.BlockingPackage == "" {
		t.Fatal(r)
	}
	if normalizeScan(context.Background(), s, 1, false).Reason != "WORK_LIMIT" {
		t.Fatal("unbounded work")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if normalizeScan(ctx, s, MaximumWork, false).Reason != "EXECUTION_LIMIT" {
		t.Fatal("ignored cancellation")
	}
}

func TestBandsAndRounding(t *testing.T) {
	for _, c := range []struct {
		v     int
		label string
	}{{0, "LOW"}, {19, "LOW"}, {20, "MODERATE"}, {39, "MODERATE"}, {40, "ELEVATED"}, {59, "ELEVATED"}, {60, "HIGH"}, {79, "HIGH"}, {80, "VERY HIGH"}, {100, "VERY HIGH"}} {
		if classification(c.v) != c.label {
			t.Fatal(c)
		}
	}
	// Application-only contribution 50 * 39/100 = 19.5 rounds to 20.
	r := score(domain.ImpactAnalysis{ApplicationsExposed: 39}, normalization{Applications: 100, Counts: map[string]int{}})
	if *r.Score != 20 || *r.Classification != "MODERATE" {
		t.Fatal(r)
	}
	r = score(domain.ImpactAnalysis{ApplicationsExposed: 389}, normalization{Applications: 1000, Counts: map[string]int{}})
	if *r.Score != 19 {
		t.Fatal("rounded below boundary", r)
	}
}

func TestDepthAndCacheIsolation(t *testing.T) {
	ids := map[string]string{"app": "application"}
	edges := [][2]string{}
	previous := "app"
	for i := 0; i < 101; i++ {
		id := fmt.Sprint(i)
		ids[id] = "package"
		edges = append(edges, [2]string{previous, id})
		previous = id
	}
	n := normalizeScan(context.Background(), snapshot(ids, edges...), MaximumWork, false)
	if n.Reason != "SCAN_PATH_DEPTH_LIMIT" {
		t.Fatal(n)
	}
	s := snapshot(map[string]string{"app": "application", "x": "package"}, [2]string{"app", "x"})
	impact, _ := Analyze(s, "x", domain.PathOptions{})
	cache := &Cache{}
	first := cache.Score(context.Background(), "one", s, impact, false)
	s.Edges = nil
	impact, _ = Analyze(s, "x", domain.PathOptions{})
	changed := cache.Score(context.Background(), "one", s, impact, false)
	other := cache.Score(context.Background(), "two", s, impact, false)
	if *first.Score != 70 || *changed.Score != 0 || *other.Score != 0 || len(cache.values) != 3 {
		t.Fatal("cache contaminated snapshot or scan")
	}
}
