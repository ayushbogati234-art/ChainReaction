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
	Tool struct {
		Poetry poetryManifest `toml:"poetry"`
	} `toml:"tool"`
}

func parsePyproject(data []byte) (poetryManifest, error) {
	var document pyprojectDocument
	if err := toml.Unmarshal(data, &document); err != nil {
		return poetryManifest{}, fmt.Errorf("invalid pyproject.toml: %w", err)
	}
	if document.Tool.Poetry.Name == "" {
		return poetryManifest{}, fmt.Errorf("UNSUPPORTED_INPUT: only [tool.poetry] projects are supported in Phase 1")
	}
	return document.Tool.Poetry, nil
}
