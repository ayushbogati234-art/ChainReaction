package python

import (
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
)

type poetryManifest struct {
	DevDependencies map[string]any `toml:"dev-dependencies"`
	Group           map[string]struct {
		Optional     bool
		Dependencies map[string]any
	} `toml:"group"`
	Name         string         `toml:"name"`
	Dependencies map[string]any `toml:"dependencies"`
}

type pyprojectDocument struct {
	Project          map[string]any `toml:"project"`
	DependencyGroups map[string]any `toml:"dependency-groups"`
	Tool             struct {
		Poetry poetryManifest `toml:"poetry"`
	} `toml:"tool"`
}

func parsePyproject(data []byte) (poetryManifest, error) {
	var document pyprojectDocument
	if err := toml.Unmarshal(data, &document); err != nil {
		return poetryManifest{}, fmt.Errorf("UNSUPPORTED_INPUT: invalid pyproject.toml: %w", err)
	}
	// Do not silently accept a Poetry name while ignoring another declaration format.
	for _, key := range []string{"dependencies", "optional-dependencies", "dynamic"} {
		if _, exists := document.Project[key]; exists {
			return poetryManifest{}, fmt.Errorf("UNSUPPORTED_INPUT: [project].%s is not supported; Phase 1 reads dependency declarations from [tool.poetry] only", key)
		}
	}
	if len(document.DependencyGroups) > 0 {
		return poetryManifest{}, fmt.Errorf("UNSUPPORTED_INPUT: [dependency-groups] is not supported; use Poetry dependency groups for Phase 1")
	}
	if name, exists := document.Project["name"]; exists && name != document.Tool.Poetry.Name {
		return poetryManifest{}, fmt.Errorf("UNSUPPORTED_INPUT: inconsistent application names in [project] and [tool.poetry]")
	}
	if document.Tool.Poetry.Name == "" {
		return poetryManifest{}, fmt.Errorf("UNSUPPORTED_INPUT: only [tool.poetry] projects are supported in Phase 1")
	}
	return document.Tool.Poetry, nil
}
