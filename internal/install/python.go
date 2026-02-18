package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// PythonInstaller handles Python installation from python-build-standalone.
type PythonInstaller struct{}

func (p *PythonInstaller) Name() string { return "python" }

// githubRelease represents a GitHub release.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset represents a release asset.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (p *PythonInstaller) ResolveVersion(requirement string) (string, error) {
	// Fetch the latest release from python-build-standalone
	data, err := FetchJSON("https://api.github.com/repos/indygreg/python-build-standalone/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to fetch python-build-standalone releases: %w", err)
	}

	var release githubRelease
	if err := json.Unmarshal(data, &release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	// Extract Python version from asset names
	// Assets look like: cpython-3.13.2+20250212-x86_64-unknown-linux-gnu-install_only_stripped.tar.gz
	versions := map[string]bool{}
	for _, asset := range release.Assets {
		if v := extractPythonVersion(asset.Name); v != "" {
			versions[v] = true
		}
	}

	if requirement == "latest" {
		// Return the highest version found
		var best *semver.Version
		for v := range versions {
			sv, err := semver.NewVersion(v)
			if err != nil {
				continue
			}
			if best == nil || sv.GreaterThan(best) {
				best = sv
			}
		}
		if best != nil {
			return best.Original(), nil
		}
		return "", fmt.Errorf("no Python versions found in release")
	}

	// Find the latest version satisfying the constraint
	constraint, err := semver.NewConstraint(requirement)
	if err != nil {
		// Fall back to latest
		return p.ResolveVersion("latest")
	}

	var best *semver.Version
	for v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		if constraint.Check(sv) && (best == nil || sv.GreaterThan(best)) {
			best = sv
		}
	}

	if best != nil {
		return best.Original(), nil
	}

	return "", fmt.Errorf("no Python version satisfying %s found", requirement)
}

func (p *PythonInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	// Fetch the release to find the correct asset URL
	data, err := FetchJSON("https://api.github.com/repos/indygreg/python-build-standalone/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	var release githubRelease
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	target := pythonTarget()
	var assetURL, assetName string

	// Find the install_only asset for our platform
	for _, asset := range release.Assets {
		if !strings.Contains(asset.Name, "cpython-"+version) {
			continue
		}
		if !strings.Contains(asset.Name, target) {
			continue
		}
		if strings.Contains(asset.Name, "install_only") {
			assetURL = asset.BrowserDownloadURL
			assetName = asset.Name
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no Python %s binary found for %s", version, target)
	}

	// Download
	tmpFile := filepath.Join(os.TempDir(), assetName)
	defer os.Remove(tmpFile)

	if err := DownloadFile(assetURL, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download Python: %w", err)
	}

	// python-build-standalone archives have a "python/" top-level dir
	if err := ExtractAndFlatten(tmpFile, targetDir); err != nil {
		return fmt.Errorf("failed to extract Python: %w", err)
	}

	return nil
}

func (p *PythonInstaller) BinDir(installDir string) string {
	if runtime.GOOS == "windows" {
		return installDir // python.exe at root on Windows
	}
	return filepath.Join(installDir, "bin")
}

func (p *PythonInstaller) EnvVars(installDir string) map[string]string { return nil }

// extractPythonVersion pulls the Python version from an asset name like
// "cpython-3.13.2+20250212-x86_64-unknown-linux-gnu-install_only.tar.gz"
func extractPythonVersion(name string) string {
	if !strings.HasPrefix(name, "cpython-") {
		return ""
	}
	// Remove "cpython-" prefix
	rest := strings.TrimPrefix(name, "cpython-")
	// Version is before the "+" character
	plusIdx := strings.IndexByte(rest, '+')
	if plusIdx < 0 {
		return ""
	}
	return rest[:plusIdx]
}

// pythonTarget returns the platform target string used in python-build-standalone.
func pythonTarget() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "aarch64-unknown-linux-gnu"
		}
		return "x86_64-unknown-linux-gnu"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "aarch64-apple-darwin"
		}
		return "x86_64-apple-darwin"
	case "windows":
		return "x86_64-pc-windows-msvc"
	default:
		return "x86_64-unknown-linux-gnu"
	}
}
