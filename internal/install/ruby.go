package install

import (
	"fmt"
	"path/filepath"
)

// RubyInstaller handles Ruby installation.
type RubyInstaller struct{}

func (r *RubyInstaller) Name() string { return "ruby" }

func (r *RubyInstaller) ResolveVersion(requirement string) (string, error) {
	return "", fmt.Errorf("Ruby installer not yet implemented - install Ruby manually from https://www.ruby-lang.org/en/downloads/")
}

func (r *RubyInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	return fmt.Errorf("Ruby installer not yet implemented - install Ruby %s manually from https://www.ruby-lang.org/en/downloads/", version)
}

func (r *RubyInstaller) BinDir(installDir string) string {
	return filepath.Join(installDir, "bin")
}

func (r *RubyInstaller) EnvVars(installDir string) map[string]string { return nil }
