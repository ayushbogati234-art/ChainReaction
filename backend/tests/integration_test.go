package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	graphstore "github.com/ayushbogati234-art/ChainReaction/backend/internal/graph/neo4j"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/jobs"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/persistence/postgres"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/resolver/python"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/service"
	transport "github.com/ayushbogati234-art/ChainReaction/backend/internal/transport/http"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Real PostgreSQL and Neo4j; GitHub HTTP alone is replaced by deterministic
// archive responses. Set PHASE1_INTEGRATION=1 explicitly to enable.
func TestDatabaseWorkflow(t *testing.T) {
	if os.Getenv("PHASE1_INTEGRATION") != "1" {
		t.Skip("requires live PostgreSQL and Neo4j")
	}
	ctx := context.Background()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		dsn = "postgres://chainreaction:chainreaction@localhost:5432/chainreaction?sslmode=disable"
	}
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := "phase1_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := postgres.Store{DB: db}
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	uri := os.Getenv("TEST_NEO4J_URI")
	if uri == "" {
		uri = os.Getenv("NEO4J_URI")
	}
	if uri == "" {
		uri = "bolt://localhost:7687"
	}
	password := os.Getenv("TEST_NEO4J_PASSWORD")
	if password == "" {
		password = os.Getenv("NEO4J_PASSWORD")
	}
	if password == "" {
		password = "chainreaction-dev"
	}
	driver, err := neo.NewDriverWithContext(uri, neo.BasicAuth("neo4j", password, ""))
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close(ctx)
	if err = driver.VerifyConnectivity(ctx); err != nil {
		t.Fatal(err)
	}
	graph := graphstore.Repository{Driver: driver}
	var scans []string
	defer func() {
		session := driver.NewSession(ctx, neo.SessionConfig{})
		defer session.Close(ctx)
		result, err := session.Run(ctx, "MATCH (n) WHERE n.scan_id IN $scans DETACH DELETE n", map[string]any{"scans": scans})
		if err != nil {
			t.Error(err)
		} else if _, err = result.Consume(ctx); err != nil {
			t.Error(err)
		}
	}()
	scanService := service.ScanService{Store: store, Graph: graph, Resolver: python.Resolver{}}
	handler := transport.Router(service.ProjectService{Store: store}, scanService, service.GraphService{Graph: graph, Scans: store})
	preflight := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/v1/projects", nil)
	handler.ServeHTTP(preflight, req)
	if preflight.Code != 204 || !strings.Contains(preflight.Header().Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Fatal("browser preflight blocked")
	}
	enqueue := func(label string) domain.Scan {
		var project domain.Project
		if err := json.Unmarshal(apiRequest(t, handler, "POST", "/api/v1/projects", map[string]string{"repository_url": exampleURL(label)}, 201), &project); err != nil {
			t.Fatal(err)
		}
		scan := decodeScan(t, apiRequest(t, handler, "POST", "/api/v1/projects/"+project.ID+"/scans", nil, 202))
		scans = append(scans, scan.ID)
		var status string
		if err := db.QueryRow(ctx, "SELECT status FROM scan_jobs WHERE scan_id=$1", scan.ID).Scan(&status); err != nil || status != "QUEUED" {
			t.Fatalf("queue not persisted: %s %v", status, err)
		}
		return scan
	}
	process := func(files map[string][]byte) {
		scanService.Ingest = fixtureClient(t, files)
		job := jobs.ScanJob{Queue: store, Runner: scanService}
		worked, err := job.ProcessOne(ctx)
		if err != nil || !worked {
			t.Fatalf("worker: %t %v", worked, err)
		}
	}
	first := enqueue("chain")
	// A separate connection observes the queued job before the worker starts.
	observer, err := pgxpool.NewWithConfig(ctx, cfg.Copy())
	if err != nil {
		t.Fatal(err)
	}
	var persisted string
	err = observer.QueryRow(ctx, "SELECT status FROM scan_jobs WHERE scan_id=$1", first.ID).Scan(&persisted)
	observer.Close()
	if err != nil || persisted != "QUEUED" {
		t.Fatalf("queue persistence: %s %v", persisted, err)
	}
	apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/graph", nil, 409)
	process(map[string][]byte{"pyproject.toml": fixture(t, "chain/pyproject.toml"), "poetry.lock": fixture(t, "chain/poetry.lock")})
	completed := decodeScan(t, apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID, nil, 200))
	if completed.Status != "COMPLETED" || completed.CommitSHA != fixtureSHA || !completed.Incomplete || !strings.Contains(strings.Join(completed.Limitations, " "), "Version constraint compatibility is unverified") {
		t.Fatalf("completion: %+v", completed)
	}
	statusScan := decodeScan(t, apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/status", nil, 200))
	if statusScan.CommitSHA != fixtureSHA {
		t.Fatal("polling response lost repository provenance")
	}
	var topology struct {
		Nodes []map[string]any
		Edges []map[string]any
	}
	if err = json.Unmarshal(apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/graph", nil, 200), &topology); err != nil {
		t.Fatal(err)
	}
	if len(topology.Nodes) != 3 || len(topology.Edges) != 2 {
		t.Fatalf("graph counts: %+v", topology)
	}
	apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/packages/a@1.0.0", nil, 200)
	apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/packages/a%401.0.0", nil, 200)
	apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/packages/b%402.0.0/paths", nil, 200)
	var paths domain.PathResult
	if err = json.Unmarshal(apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/packages/b@2.0.0/paths", nil, 200), &paths); err != nil {
		t.Fatal(err)
	}
	if len(paths.Paths) != 1 || strings.Join(paths.Paths[0], " -> ") != "app:chain -> a@1.0.0 -> b@2.0.0" {
		t.Fatalf("wrong paths: %+v", paths)
	}
	changed := enqueue("provenance")
	process(map[string][]byte{"pyproject.toml": fixture(t, "provenance/pyproject.toml"), "poetry.lock": fixture(t, "chain/poetry.lock")})
	provenance := string(apiRequest(t, handler, "GET", "/api/v1/scans/"+changed.ID+"/graph", nil, 200))
	for _, evidence := range []string{"input_files", "pyproject.toml", "poetry.lock", "app:changed-application", "a@1.0.0", "b@2.0.0"} {
		if !strings.Contains(provenance, evidence) {
			t.Fatalf("missing persisted evidence %s: %s", evidence, provenance)
		}
	}
	if err = json.Unmarshal(apiRequest(t, handler, "GET", "/api/v1/scans/"+changed.ID+"/packages/b@2.0.0/paths", nil, 200), &paths); err != nil {
		t.Fatal(err)
	}
	if len(paths.Paths) != 1 || strings.Join(paths.Paths[0], " -> ") != "app:changed-application -> b@2.0.0" {
		t.Fatalf("manifest change not persisted: %+v", paths)
	}
	if err = json.Unmarshal(apiRequest(t, handler, "GET", "/api/v1/scans/"+changed.ID+"/packages/a@1.0.0/paths", nil, 200), &paths); err != nil || len(paths.Paths) != 0 {
		t.Fatalf("removed direct declaration still reachable: %+v %v", paths, err)
	}
	second := enqueue("other")
	process(map[string][]byte{"pyproject.toml": fixture(t, "chain/pyproject.toml"), "poetry.lock": bytesReplace(fixture(t, "chain/poetry.lock"), "2.0.0", "3.0.0")})
	apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/packages/b@3.0.0", nil, 404)
	apiRequest(t, handler, "GET", "/api/v1/scans/"+second.ID+"/packages/b@2.0.0", nil, 404)
	apiRequest(t, handler, "GET", "/api/v1/scans/"+second.ID+"/packages/b@2.0.0/paths", nil, 404)
	// Introduce a cross-scan edge deliberately: queries must still exclude it.
	session := driver.NewSession(ctx, neo.SessionConfig{})
	result, err := session.Run(ctx, "MATCH (a {scan_id:$first,id:'a@1.0.0'}),(b {scan_id:$second,id:'b@3.0.0'}) CREATE (a)-[:DEPENDS_ON]->(b)", map[string]any{"first": first.ID, "second": second.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = result.Consume(ctx); err != nil {
		t.Fatal(err)
	}
	session.Close(ctx)
	body := apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/packages/a@1.0.0", nil, 200)
	if strings.Contains(string(body), "b@3.0.0") {
		t.Fatal("cross-scan package leakage")
	}
	body = apiRequest(t, handler, "GET", "/api/v1/scans/"+first.ID+"/graph", nil, 200)
	if strings.Contains(string(body), second.ID) || strings.Contains(string(body), "b@3.0.0") {
		t.Fatal("cross-scan graph leakage")
	}
	unsupported := enqueue("unsupported")
	process(map[string][]byte{"requirements.txt": []byte("a==1.0.0")})
	failed := decodeScan(t, apiRequest(t, handler, "GET", "/api/v1/scans/"+unsupported.ID, nil, 200))
	if failed.Status != "FAILED" || failed.ErrorCode == nil || *failed.ErrorCode != "UNSUPPORTED_INPUT" {
		t.Fatalf("unsupported error: %+v", failed)
	}
	conditional := enqueue("conditional")
	process(map[string][]byte{"pyproject.toml": []byte("[tool.poetry]\nname='conditioned'\n[tool.poetry.dependencies]\nconditioned='^1'\n"), "poetry.lock": fixture(t, "conditional/poetry.lock")})
	incomplete := decodeScan(t, apiRequest(t, handler, "GET", "/api/v1/scans/"+conditional.ID, nil, 200))
	if incomplete.Status != "COMPLETED" || !incomplete.Incomplete || len(incomplete.Limitations) == 0 {
		t.Fatalf("incomplete reporting: %+v", incomplete)
	}
	body = apiRequest(t, handler, "GET", "/api/v1/scans/"+conditional.ID+"/graph", nil, 200)
	if !strings.Contains(string(body), "win32") {
		t.Fatal("persisted marker lost")
	}
	abandoned := enqueue("abandoned")
	if _, _, err = store.Claim(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "UPDATE scan_jobs SET locked_at=now()-interval '4 minutes' WHERE scan_id=$1", abandoned.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = store.Claim(ctx); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	abandoned, err = store.GetScan(ctx, abandoned.ID)
	if err != nil || abandoned.Status != "FAILED" || abandoned.ErrorCode == nil || *abandoned.ErrorCode != "WORKER_INTERRUPTED" {
		t.Fatalf("abandoned job: %+v %v", abandoned, err)
	}
	t.Log("PASS: persisted queue, completion, package inspection, paths, isolation including cross-scan edge, unsupported input, incomplete evidence, abandoned-worker recovery")
}
func bytesReplace(b []byte, old, new string) []byte {
	return []byte(strings.ReplaceAll(string(b), old, new))
}
