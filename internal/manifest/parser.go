package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const DefaultManifestName = ".templatr.toml"

// Load reads and parses a .templatr.toml file from the given path.
// If path is empty, it looks for .templatr.toml in the current directory.
func Load(path string) (*Manifest, error) {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		path = filepath.Join(cwd, DefaultManifestName)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest file not found: %s\n\nMake sure you're in a Templatr template directory that contains a %s file", path, DefaultManifestName)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &m, nil
}
