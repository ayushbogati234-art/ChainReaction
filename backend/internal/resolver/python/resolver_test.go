package python

import (
	"strings"
	"testing"
)

func TestUnsupportedAndInconsistentDeclarations(t *testing.T) {
	lock := []byte("[[package]]\nname='a'\nversion='1.0.0'\n")
	for _, tc := range []struct{ name, declaration, message string }{
		{"exact mismatch", "'2.0.0'", "Inconsistent metadata"},
		{"unknown field", "{version='*', mystery=true}", "Unsupported dependency field"},
		{"wrong optional type", "{version='*',optional='yes'}", "Unsupported dependency field"},
		{"wrong version type", "{version=42}", "Unsupported dependency field"},
		{"path", "{path='../a'}", "Unsupported path dependency"},
		{"git", "{git='https://example.com/a.git'}", "Unsupported git dependency"},
		{"url", "{url='https://example.com/a.whl'}", "Unsupported url dependency"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, err := (Resolver{}).Resolve(Input{Pyproject: []byte("[tool.poetry]\nname='demo'\n[tool.poetry.dependencies]\na=" + tc.declaration + "\n"), PoetryLock: lock})
			if err != nil || !g.Incomplete || len(g.Packages) != 1 || len(g.Edges) != 0 || !strings.Contains(strings.Join(g.Limitations, " "), tc.message) {
				t.Fatalf("%+v %v", g, err)
			}
		})
	}
	for _, extra := range []string{"[project]\ndependencies=['a']", "[project]\noptional-dependencies={test=['a']}", "[project]\ndynamic=['dependencies']", "[project]\nname='other'", "[dependency-groups]\ntest=['a']"} {
		_, err := (Resolver{}).Resolve(Input{Pyproject: []byte("[tool.poetry]\nname='demo'\n" + extra), PoetryLock: lock})
		if err == nil || !strings.Contains(err.Error(), "UNSUPPORTED_INPUT") {
			t.Fatalf("silently ignored %s: %v", extra, err)
		}
	}
	for _, input := range []Input{{PoetryLock: lock}, {Pyproject: []byte("broken["), PoetryLock: lock}, {Pyproject: []byte("[tool.poetry]\nname='demo'"), PoetryLock: []byte("broken[")}} {
		if _, err := (Resolver{}).Resolve(input); err == nil || !strings.Contains(err.Error(), "UNSUPPORTED_INPUT") {
			t.Fatalf("unclear input error: %v", err)
		}
	}
}

func TestResolveChainAndEvidence(t *testing.T) {
	py := []byte("[tool.poetry]\nname='demo'\n[tool.poetry.dependencies]\npython='^3.11'\na='^1.0'\n")
	lock := []byte("[[package]]\nname='a'\nversion='1.0.0'\ncategory='main'\n[package.dependencies]\nb='^2'\n[[package]]\nname='b'\nversion='2.0.0'\ncategory='main'\n")
	g, e := Resolver{}.Resolve(Input{Pyproject: py, PoetryLock: lock})
	if e != nil {
		t.Fatal(e)
	}
	if len(g.Packages) != 2 || len(g.Edges) != 2 {
		t.Fatalf("packages=%d edges=%d", len(g.Packages), len(g.Edges))
	}
	if g.Edges[0].From != "a@1.0.0" || g.Edges[0].To != "b@2.0.0" {
		t.Fatalf("wrong dependent direction: %#v", g.Edges[0])
	}
	if !g.Packages[0].Direct {
		t.Fatal("direct dependency was not identified")
	}
}
func TestResolveFlagsAmbiguityAndCondition(t *testing.T) {
	py := []byte("[tool.poetry]\nname='demo'\n[tool.poetry.dependencies]\npython='^3.11'\na='^1'\n")
	lock := []byte("[[package]]\nname='a'\nversion='1.0.0'\ncategory='main'\n[package.dependencies]\nb={version='^2', markers=\"sys_platform == 'win32'\"}\n[[package]]\nname='b'\nversion='2.0.0'\n[[package]]\nname='b'\nversion='2.1.0'\n")
	g, e := Resolver{}.Resolve(Input{Pyproject: py, PoetryLock: lock})
	if e != nil {
		t.Fatal(e)
	}
	if !g.Incomplete || len(g.Limitations) == 0 {
		t.Fatal("ambiguous conditional graph must be incomplete")
	}
}
func TestRejectsMissingLock(t *testing.T) {
	_, e := Resolver{}.Resolve(Input{Pyproject: []byte("[tool.poetry]\nname='demo'\n")})
	if e == nil {
		t.Fatal("expected unsupported input")
	}
}
