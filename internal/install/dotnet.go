package install

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// DotnetInstaller handles .NET SDK installation.
type DotnetInstaller struct{}

func (d *DotnetInstaller) Name() string { return "dotnet" }

func (d *DotnetInstaller) ResolveVersion(requirement string) (string, error) {
	return "", fmt.Errorf(".NET installer not yet implemented - install .NET manually from https://dot.net/download")
}

func (d *DotnetInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	return fmt.Errorf(".NET installer not yet implemented - install .NET %s manually from https://dot.net/download", version)
}

func (d *DotnetInstaller) BinDir(installDir string) string {
	if runtime.GOOS == "windows" {
		return installDir
	}
	return filepath.Join(installDir, "bin")
}

func (d *DotnetInstaller) EnvVars(installDir string) map[string]string { return nil }
