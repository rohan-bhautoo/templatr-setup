package install

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// PHPInstaller handles PHP installation.
type PHPInstaller struct{}

func (p *PHPInstaller) Name() string { return "php" }

func (p *PHPInstaller) ResolveVersion(requirement string) (string, error) {
	return "", fmt.Errorf("PHP installer not yet implemented — install PHP manually from https://www.php.net/downloads")
}

func (p *PHPInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	return fmt.Errorf("PHP installer not yet implemented — install PHP %s manually from https://www.php.net/downloads", version)
}

func (p *PHPInstaller) BinDir(installDir string) string {
	if runtime.GOOS == "windows" {
		return installDir
	}
	return filepath.Join(installDir, "bin")
}

func (p *PHPInstaller) EnvVars(installDir string) map[string]string { return nil }
