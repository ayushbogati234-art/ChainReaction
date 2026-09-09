package python

import (
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
)

type lockedPackage struct {
	Groups         []string
	Markers        any
	Optional       bool
	PythonVersions string `toml:"python-versions"`
	Extras         map[string]any
	Name           string
	Version        string
	Category       string
	Marker         string
	Source         map[string]any
	Dependencies   map[string]any `toml:"dependencies"`
}

type poetryLockDocument struct {
	Packages []lockedPackage `toml:"package"`
}

func parsePoetryLock(data []byte) ([]lockedPackage, error) {
	var document poetryLockDocument
	if err := toml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("invalid poetry.lock: %w", err)
	}
	if len(document.Packages) == 0 {
		return nil, fmt.Errorf("UNSUPPORTED_INPUT: poetry.lock contains no package records")
	}
	return document.Packages, nil
}
