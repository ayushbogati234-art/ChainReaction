// Package python parses metadata only; it never executes repository code.
package python

import (
	"encoding/json"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	resolvercontract "github.com/ayushbogati234-art/ChainReaction/backend/internal/resolver"
	"regexp"
	"sort"
	"strings"
)

type Input = resolvercontract.Input
type Resolver struct{}

var _ resolvercontract.Resolver = Resolver{}
var separators = regexp.MustCompile("[-_.]+")
var exactRelease = regexp.MustCompile(`^(?:==|=)?([0-9]+(?:\.[0-9]+)*)$`)

func normalize(n string) string { return separators.ReplaceAllString(strings.ToLower(n), "-") }
func id(n, v string) string     { return normalize(n) + "@" + v }
func metadata(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}
func markers(parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, "("+p+")")
		}
	}
	return strings.Join(out, " and ")
}
func (Resolver) Resolve(in Input) (domain.Graph, error) {
	if in.Path != "" && in.Path != "." {
		return domain.Graph{}, fmt.Errorf("UNSUPPORTED_INPUT: only the repository root is supported")
	}
	if len(in.Pyproject) == 0 || len(in.PoetryLock) == 0 {
		return domain.Graph{}, fmt.Errorf("UNSUPPORTED_INPUT: paired root pyproject.toml and poetry.lock are required; requirements.txt-only, lock-only and manifest-only resolution are deferred")
	}
	manifest, err := parsePyproject(in.Pyproject)
	if err != nil {
		return domain.Graph{}, err
	}
	locked, err := parsePoetryLock(in.PoetryLock)
	if err != nil {
		return domain.Graph{}, err
	}
	g := domain.Graph{InputFiles: []string{"pyproject.toml", "poetry.lock"}, Application: domain.Package{ID: "app:" + manifest.Name, Name: manifest.Name, Source: "pyproject.toml", Confidence: domain.ConfidenceHigh}, Packages: []domain.Package{}, Edges: []domain.Edge{}, Limitations: []string{}}
	limitation := func(message string) { g.Incomplete = true; g.Limitations = append(g.Limitations, message) }
	unverifiedConstraints := false
	byName := map[string][]lockedPackage{}
	accepted := map[string]bool{}
	packageIDs := map[string]bool{}
	for _, p := range locked {
		if p.Name == "" || p.Version == "" || strings.ContainsAny(p.Version, "*^<>=~ ") {
			return domain.Graph{}, fmt.Errorf("UNSUPPORTED_INPUT: lock package requires a name and exact version")
		}
		byName[normalize(p.Name)] = append(byName[normalize(p.Name)], p)
	}
	for name, candidates := range byName {
		if len(candidates) > 1 {
			limitation("Ambiguous lock entries for " + name + "; relationships involving that name are omitted")
		}
	}
	for _, p := range locked {
		sourceType, _ := p.Source["type"].(string)
		if sourceType != "" && sourceType != "legacy" && sourceType != "pypi" {
			limitation("Unsupported non-registry lock source for " + p.Name + ": " + metadata(p.Source))
			continue
		}
		key := id(p.Name, p.Version)
		if packageIDs[key] {
			continue
		}
		packageIDs[key] = true
		accepted[key] = true
		group := strings.Join(p.Groups, ",")
		if group == "" {
			group = p.Category
		}
		marker := markers(p.Marker, metadata(p.Markers))
		limits := []string{}
		if marker != "" || p.Optional || len(p.Extras) > 0 || (p.PythonVersions != "" && p.PythonVersions != "*") || (group != "" && group != "main") {
			message := "Unselected environment, extras, optional or group conditions for " + p.Name
			limitation(message)
			limits = append(limits, message)
		}
		source := "poetry.lock"
		if len(p.Source) > 0 {
			source += " " + metadata(p.Source)
		}
		g.Packages = append(g.Packages, domain.Package{ID: key, Name: p.Name, Version: p.Version, PURL: "pkg:pypi/" + normalize(p.Name) + "@" + p.Version, Source: source, Groups: group, Marker: marker, PythonVersions: p.PythonVersions, Extras: metadata(p.Extras), Optional: p.Optional, Confidence: domain.ConfidenceHigh, Limitations: limits})
	}
	var add func(string, string, any, bool, string, string, string, bool)
	add = func(from, name string, raw any, direct bool, group, source, parentMarker string, optional bool) {
		if variants, ok := raw.([]any); ok {
			limitation("Unselected dependency alternatives for " + from + " -> " + name)
			for _, variant := range variants {
				add(from, name, variant, direct, group, source, parentMarker, optional)
			}
			return
		}
		edge := domain.Edge{From: from, Direct: direct, Group: group, Source: source, Optional: optional, Confidence: string(domain.ConfidenceHigh)}
		var declaredMarker string
		switch v := raw.(type) {
		case string:
			edge.Constraint = v
		case map[string]any:
			for key, value := range v {
				valid := false
				switch key {
				case "version", "markers", "python", "platform", "source", "git", "path", "url":
					_, valid = value.(string)
				case "optional":
					_, valid = value.(bool)
				case "extras":
					if items, ok := value.([]any); ok {
						valid = true
						for _, item := range items {
							if _, ok := item.(string); !ok {
								valid = false
							}
						}
					}
				}
				if !valid {
					limitation("Unsupported dependency field or value " + key + " in " + source + " for " + from + " -> " + name + "; relationship omitted")
					return
				}
			}
			for _, key := range []string{"git", "path", "url"} {
				if _, exists := v[key]; exists {
					limitation("Unsupported " + key + " dependency " + from + " -> " + name + ": " + metadata(v))
					return
				}
			}
			edge.Constraint = metadata(v["version"])
			declaredMarker = metadata(v["markers"])
			for _, key := range []string{"python", "platform"} {
				if value := metadata(v[key]); value != "" {
					declaredMarker = markers(declaredMarker, key+"="+value)
				}
			}
			edge.Extras = metadata(v["extras"])
			if value, ok := v["optional"].(bool); ok {
				edge.Optional = edge.Optional || value
			}
			if value := metadata(v["source"]); value != "" {
				edge.Source += " source=" + value
			}
		default:
			limitation("Unsupported dependency metadata for " + from + " -> " + name)
			return
		}
		targets := byName[normalize(name)]
		if len(targets) != 1 || !accepted[id(targets[0].Name, targets[0].Version)] {
			limitation("Dependency " + from + " -> " + name + " does not map uniquely to a supported locked package")
			return
		}
		target := targets[0]
		// Check only plain release equality. Full Poetry/PEP 440 constraint
		// evaluation and lock content-hash validation are outside this parser.
		constraint := strings.TrimSpace(edge.Constraint)
		if match := exactRelease.FindStringSubmatch(constraint); match != nil && exactRelease.MatchString(target.Version) {
			canonical := func(v string) string {
				parts := strings.Split(v, ".")
				for len(parts) > 1 && parts[len(parts)-1] == "0" {
					parts = parts[:len(parts)-1]
				}
				return strings.Join(parts, ".")
			}
			if canonical(match[1]) != canonical(target.Version) {
				limitation("Inconsistent metadata: " + source + " declares " + name + " " + constraint + " but poetry.lock records " + target.Version + "; relationship omitted")
				return
			}
		} else if constraint != "*" {
			unverifiedConstraints = true
		}
		edge.To = id(target.Name, target.Version)
		edge.Marker = markers(parentMarker, declaredMarker, target.Marker, metadata(target.Markers))
		if edge.Marker != "" || edge.Extras != "" || edge.Optional || (group != "" && group != "main") {
			limitation("Unselected dependency conditions on " + from + " -> " + name)
		}
		g.Edges = append(g.Edges, edge)
	}
	for _, p := range locked {
		if !accepted[id(p.Name, p.Version)] || len(byName[normalize(p.Name)]) != 1 {
			continue
		}
		group := strings.Join(p.Groups, ",")
		if group == "" {
			group = p.Category
		}
		for dep, raw := range p.Dependencies {
			add(id(p.Name, p.Version), dep, raw, false, group, "poetry.lock", markers(p.Marker, metadata(p.Markers)), p.Optional)
		}
	}
	for name, raw := range manifest.Dependencies {
		if normalize(name) != "python" {
			add(g.Application.ID, name, raw, true, "main", "pyproject.toml + poetry.lock", "", false)
		}
	}
	for name, raw := range manifest.DevDependencies {
		add(g.Application.ID, name, raw, true, "dev", "pyproject.toml + poetry.lock", "", false)
	}
	for group, decl := range manifest.Group {
		for name, raw := range decl.Dependencies {
			add(g.Application.ID, name, raw, true, group, "pyproject.toml + poetry.lock", "", decl.Optional)
		}
	}
	for i := range g.Packages {
		for _, edge := range g.Edges {
			if edge.Direct && edge.To == g.Packages[i].ID {
				g.Packages[i].Direct = true
			}
		}
	}
	if unverifiedConstraints {
		limitation("Version constraint compatibility is unverified: relationships match declared names to unique lock entries; ranges and non-release versions are retained as evidence, not resolved")
	}
	sort.Slice(g.Packages, func(i, j int) bool { return g.Packages[i].ID < g.Packages[j].ID })
	sort.SliceStable(g.Edges, func(i, j int) bool {
		if g.Edges[i].Direct != g.Edges[j].Direct {
			return !g.Edges[i].Direct
		}
		return metadata(g.Edges[i]) < metadata(g.Edges[j])
	})
	sort.Strings(g.Limitations)
	return g, nil
}
