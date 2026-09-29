package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/graph"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/service"
	transport "github.com/ayushbogati234-art/ChainReaction/backend/internal/transport/http"
)

type phase5Store struct {
	scan       domain.Scan
	enrichment domain.Enrichment
	readErr    error
	reads      int
}

func (s *phase5Store) GetScan(_ context.Context, id string) (domain.Scan, error) {
	if id != s.scan.ID {
		return domain.Scan{}, domain.ErrNotFound
	}
	return s.scan, nil
}
func (s *phase5Store) GetEnrichment(_ context.Context, id string) (domain.Enrichment, error) {
	s.reads++
	if id != s.scan.ID {
		return domain.Enrichment{}, domain.ErrNotFound
	}
	return s.enrichment, s.readErr
}

type phase5Graph struct {
	graph.Repository
	snapshot domain.TopologySnapshot
	reads    int
}

func (g *phase5Graph) Snapshot(_ context.Context, id string) (domain.TopologySnapshot, error) {
	g.reads++
	if id != "scan" {
		return domain.TopologySnapshot{}, domain.ErrNotFound
	}
	return g.snapshot, nil
}
func phase5Fixture() (*phase5Store, *phase5Graph, service.GraphService) {
	now := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	store := &phase5Store{scan: domain.Scan{ID: "scan", Status: "COMPLETED"}, enrichment: domain.Enrichment{Provider: "OSV", Status: "COMPLETE", GraphStatus: "COMPLETE"}}
	g := &phase5Graph{}
	for _, id := range []string{"app", "a", "b", "x"} {
		kind := "package"
		if id == "app" {
			kind = "application"
		} else {
			store.enrichment.Checks = append(store.enrichment.Checks, domain.PackageCheck{PackageID: id, Status: "CHECKED", CheckedAt: &now, Advisories: []domain.Advisory{}})
		}
		g.snapshot.Nodes = append(g.snapshot.Nodes, domain.TopologyNode{ID: id, Kind: kind, Confidence: domain.ConfidenceHigh})
	}
	for _, edge := range [][2]string{{"app", "a"}, {"app", "b"}, {"a", "x"}, {"b", "x"}} {
		g.snapshot.Edges = append(g.snapshot.Edges, domain.TopologyEdge{From: edge[0], To: edge[1], Confidence: domain.ConfidenceHigh})
	}
	return store, g, service.GraphService{Graph: g, Scans: store, Enrichment: store}
}
func TestPhase5HTTPAndEvidenceTransitions(t *testing.T) {
	store, g, svc := phase5Fixture()
	handler := transport.Router(service.ProjectService{}, service.ScanService{}, svc)
	readImpact := func(path string) domain.ImpactAnalysis {
		t.Helper()
		var impact domain.ImpactAnalysis
		if err := json.Unmarshal(apiRequest(t, handler, "GET", path, nil, 200), &impact); err != nil {
			t.Fatal(err)
		}
		return impact
	}
	path := "/api/v1/scans/scan/packages/x/impact"
	impact := readImpact(path)
	if impact.SystemicDependency.Status != "MATCH" || impact.Risk.Score != nil || *impact.StructuralCriticality.Score != 93 {
		t.Fatalf("%+v", impact)
	}
	var report domain.SystemicReport
	before := g.reads
	if err := json.Unmarshal(apiRequest(t, handler, "GET", "/api/v1/scans/scan/systemic-dependencies", nil, 200), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "COMPLETE" || report.EvaluatedPackages != 3 || len(report.Findings) != 1 || report.Findings[0].PackageID != "x" || g.reads != before+1 {
		t.Fatalf("%+v, graph reads %d", report, g.reads-before)
	}
	store.enrichment.Checks[2].Advisories = []domain.Advisory{{ID: "OSV-1", Source: "OSV", MatchConfidence: domain.ConfidenceHigh, Severity: []domain.Severity{{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H/E:F"}}}}
	impact = readImpact(path)
	if impact.Risk.Score == nil || *impact.Risk.Score != 98 || impact.SystemicDependency.Status != "NOT_MATCHED" {
		t.Fatalf("%+v", impact)
	}
	filtered := readImpact(path + "?application_id=app&max_paths=1&max_depth=1")
	if !reflect.DeepEqual(filtered.Risk, impact.Risk) || !reflect.DeepEqual(filtered.SystemicDependency, impact.SystemicDependency) {
		t.Fatal("path filters changed risk/systemic scores")
	}
	store.enrichment.Checks[2].Advisories[0].Severity[0].Score = "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"
	impact = readImpact(path)
	if impact.Risk.Score != nil || impact.Risk.Components["exploitability_evidence"].Contribution != nil {
		t.Fatal("unknown exploitability was scored")
	}
	store.readErr = errors.New("injected database outage")
	impact = readImpact(path)
	if impact.SystemicDependency.Status != "UNAVAILABLE" || impact.Risk.Score != nil || len(impact.Paths.Paths) != 2 {
		t.Fatal("coverage outage hid topology or produced a result")
	}
	reads := store.reads
	apiRequest(t, handler, "GET", "/api/v1/scans/scan/packages/x/paths", nil, 200)
	if store.reads != reads {
		t.Fatal("paths endpoint depends on enrichment")
	}
	store.readErr = nil
	store.scan.Incomplete = true
	impact = readImpact(path)
	if impact.SystemicDependency.Status != "UNAVAILABLE" || impact.Risk.Score != nil {
		t.Fatal("incomplete topology scored")
	}
	apiRequest(t, handler, "GET", "/api/v1/scans/other/packages/x/impact", nil, 404)
	apiRequest(t, handler, "GET", "/api/v1/scans/other/systemic-dependencies", nil, 404)
	apiRequest(t, handler, "GET", "/api/v1/scans/scan/packages/foreign/impact", nil, 404)
	store.scan.Status = "QUEUED"
	apiRequest(t, handler, "GET", path, nil, 409)
	apiRequest(t, handler, "GET", "/api/v1/scans/scan/systemic-dependencies", nil, 409)
}

func TestPhase5PartialAndLegacyReports(t *testing.T) {
	for _, mode := range []string{"missing check", "legacy", "incomplete topology", "cancelled", "conflicting finding"} {
		t.Run(mode, func(t *testing.T) {
			store, _, svc := phase5Fixture()
			ctx := context.Background()
			switch mode {
			case "missing check":
				store.enrichment.Checks = store.enrichment.Checks[:2]
			case "legacy":
				store.enrichment = domain.Enrichment{Status: "NOT_RUN"}
			case "incomplete topology":
				store.scan.Incomplete = true
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "conflicting finding":
				store.enrichment.Findings = []domain.Finding{{PackageID: "x", AdvisoryID: "OSV-1"}}
			}
			out, err := svc.GetSystemicDependencies(ctx, "scan")
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Findings) != 0 {
				t.Fatal("false systemic finding", out)
			}
			if mode != "conflicting finding" && out.Status != "PARTIAL" {
				t.Fatal("missing coverage presented as complete", out)
			}
			// A normal HTTP request still returns a JSON report for legacy scans.
			h := transport.Router(service.ProjectService{}, service.ScanService{}, svc)
			r := httptest.NewRecorder()
			h.ServeHTTP(r, httptest.NewRequest("GET", "/api/v1/scans/scan/systemic-dependencies", nil))
			if r.Code != 200 {
				t.Fatal(r.Code)
			}
		})
	}
}
