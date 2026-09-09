package tests

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/ingest/github"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/resolver/python"
)

const fixtureSHA = "1234567890123456789012345678901234567890"

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../fixtures/python/" + path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func archive(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tarball := tar.NewWriter(gz)
	for name, data := range files {
		if err := tarball.WriteHeader(&tar.Header{Name: "root/" + name, Size: int64(len(data)), Mode: 0644}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarball.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarball.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
func fixtureClient(t *testing.T, files map[string][]byte) github.Client {
	t.Helper()
	payload := archive(t, files)
	return github.Client{HTTP: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		body := payload
		if r.URL.Host == "api.github.com" {
			body = []byte(`{"sha":"` + fixtureSHA + `"}`)
		} else if r.URL.Host != "codeload.github.com" || !strings.HasSuffix(r.URL.Path, fixtureSHA) {
			t.Errorf("unpinned/unexpected retrieval: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}}, nil
	})}}
}
func TestControlledIngestionAndResolution(t *testing.T) {
	client := fixtureClient(t, map[string][]byte{"pyproject.toml": fixture(t, "chain/pyproject.toml"), "poetry.lock": fixture(t, "chain/poetry.lock"), "do-not-run.sh": []byte("exit 99")})
	files, err := client.Fetch(context.Background(), "https://github.com/controlled/chain")
	if err != nil {
		t.Fatal(err)
	}
	if files.CommitSHA != fixtureSHA {
		t.Fatal("missing pinned SHA")
	}
	g, err := (python.Resolver{}).Resolve(python.Input{Pyproject: files.Pyproject, PoetryLock: files.PoetryLock})
	if err != nil {
		t.Fatal(err)
	}
	if !g.Incomplete || !strings.Contains(strings.Join(g.Limitations, " "), "Version constraint compatibility is unverified") || len(g.Packages) != 2 || len(g.Edges) != 2 {
		t.Fatalf("unexpected controlled graph: %+v", g)
	}
}
func TestManifestChangesWithIdenticalLock(t *testing.T) {
	lock := fixture(t, "chain/poetry.lock")
	resolve := func(manifest string) domain.Graph {
		t.Helper()
		client := fixtureClient(t, map[string][]byte{"pyproject.toml": fixture(t, manifest), "poetry.lock": lock, "tox.ini": []byte("[testenv]\ndeps = ignored-tox-package==9\n")})
		files, err := client.Fetch(context.Background(), "https://github.com/controlled/provenance")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(files.PoetryLock, lock) {
			t.Fatal("lock changed during retrieval")
		}
		g, err := (python.Resolver{}).Resolve(python.Input{Pyproject: files.Pyproject, PoetryLock: files.PoetryLock})
		if err != nil {
			t.Fatal(err)
		}
		return g
	}
	before, after := resolve("chain/pyproject.toml"), resolve("provenance/pyproject.toml")
	if before.Application.ID != "app:chain" || after.Application.ID != "app:changed-application" || after.Application.Source != "pyproject.toml" {
		t.Fatal("manifest identity not used")
	}
	if !reflect.DeepEqual(after.InputFiles, []string{"pyproject.toml", "poetry.lock"}) {
		t.Fatal("incorrect input provenance")
	}
	for i := range before.Packages {
		if before.Packages[i].Direct == after.Packages[i].Direct {
			t.Fatal("directness did not change")
		}
		before.Packages[i].Direct, after.Packages[i].Direct = false, false
	}
	if !reflect.DeepEqual(before.Packages, after.Packages) {
		t.Fatal("unchanged lock produced different package evidence")
	}
	if !reflect.DeepEqual(before.Edges[0], after.Edges[0]) {
		t.Fatal("lock dependency evidence changed")
	}
	if len(after.Edges) != 2 || before.Edges[1].To != "a@1.0.0" || after.Edges[1].From != "app:changed-application" || after.Edges[1].To != "b@2.0.0" || after.Edges[1].Constraint != "2.0.0" || after.Edges[1].Source != "pyproject.toml + poetry.lock" {
		t.Fatalf("declaration change not reflected: %+v", after.Edges)
	}
	t.Log("same lock: a@1.0.0 and b@2.0.0 retained; a -> b retained; application identity changed; direct app -> a replaced by app -> b; tox.ini ignored")
}
func TestUnsupportedAndRetrievalFailures(t *testing.T) {
	for _, url := range []string{"http://github.com/a/b", "https://example.com/a/b", "https://user@github.com/a/b", "https://github.com/a/b?ref=main", "https://github.com/a/..", "https://github.com/a/b/tree/main"} {
		if _, _, err := github.ParseRepositoryURL(url); err == nil {
			t.Errorf("accepted invalid URL %s", url)
		}
	}
	for _, tc := range []struct {
		name, code string
		files      map[string][]byte
		limit      int64
	}{
		{"requirements", "UNSUPPORTED_INPUT", map[string][]byte{"requirements.txt": []byte("a==1")}, 0},
		{"manifest-only", "UNSUPPORTED_INPUT", map[string][]byte{"pyproject.toml": fixture(t, "chain/pyproject.toml")}, 0},
		{"nested", "UNSUPPORTED_INPUT", map[string][]byte{"app/pyproject.toml": fixture(t, "chain/pyproject.toml"), "app/poetry.lock": fixture(t, "chain/poetry.lock")}, 0},
		{"file-limit", "REPOSITORY_FILE_TOO_LARGE", map[string][]byte{"pyproject.toml": bytes.Repeat([]byte("x"), (1<<20)+1)}, 0},
		{"archive-limit", "REPOSITORY_FILE_TOO_LARGE", map[string][]byte{"pyproject.toml": []byte("test")}, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := fixtureClient(t, tc.files)
			client.MaxBytes = tc.limit
			_, err := client.Fetch(context.Background(), "https://github.com/controlled/test")
			if err == nil || !strings.Contains(err.Error(), tc.code) {
				t.Fatalf("got %v, want %s", err, tc.code)
			}
		})
	}
}
func TestConditionalMetadataAndMissingEdges(t *testing.T) {
	py := []byte("[tool.poetry]\nname='conditioned'\n[tool.poetry.dependencies]\nconditioned={version='^1',markers=\"sys_platform == 'win32'\",extras=['speed'],optional=true}\n")
	g, err := (python.Resolver{}).Resolve(python.Input{Pyproject: py, PoetryLock: fixture(t, "conditional/poetry.lock")})
	if err != nil {
		t.Fatal(err)
	}
	if !g.Incomplete || len(g.Limitations) == 0 || len(g.Edges) != 2 {
		t.Fatalf("conditional evidence lost: %+v", g)
	}
	for _, edge := range g.Edges {
		if edge.Marker == "" {
			t.Fatal("marker lost")
		}
		if edge.Direct && (edge.Extras == "" || !edge.Optional) {
			t.Fatal("direct metadata lost")
		}
	}
	lock := []byte("[[package]]\nname='conditioned'\nversion='1.0.0'\n[package.dependencies]\nmissing='*'\n")
	g, err = (python.Resolver{}).Resolve(python.Input{Pyproject: py, PoetryLock: lock})
	if err != nil || !g.Incomplete || !strings.Contains(strings.Join(g.Limitations, " "), "missing") {
		t.Fatalf("unresolved relation not explained: %+v %v", g, err)
	}
}

func apiRequest(t *testing.T, handler http.Handler, method, path string, body any, status int) []byte {
	t.Helper()
	raw, _ := json.Marshal(body)
	request := httptest.NewRequest(method, path, bytes.NewReader(raw))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != status {
		t.Fatalf("%s %s: got %d %s; expected %d", method, path, response.Code, response.Body.String(), status)
	}
	return response.Body.Bytes()
}
func decodeScan(t *testing.T, raw []byte) domain.Scan {
	t.Helper()
	var s domain.Scan
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	return s
}
func exampleURL(label string) string { return fmt.Sprintf("https://github.com/controlled/%s", label) }
