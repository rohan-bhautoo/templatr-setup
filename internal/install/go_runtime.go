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

// GoInstaller handles Go installation from go.dev.
type GoInstaller struct{}

func (g *GoInstaller) Name() string { return "go" }

// goVersion represents an entry from go.dev/dl/?mode=json.
type goVersion struct {
	Version string   `json:"version"` // "go1.22.5"
	Stable  bool     `json:"stable"`
	Files   []goFile `json:"files"`
}

type goFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"` // "archive", "installer", "source"
}

func (g *GoInstaller) ResolveVersion(requirement string) (string, error) {
	data, err := FetchJSON("https://go.dev/dl/?mode=json")
	if err != nil {
		return "", fmt.Errorf("failed to fetch Go versions: %w", err)
	}

	var versions []goVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return "", fmt.Errorf("failed to parse Go versions: %w", err)
	}

	// Filter to stable versions
	var stable []goVersion
	for _, v := range versions {
		if v.Stable {
			stable = append(stable, v)
		}
	}

	if len(stable) == 0 {
		return "", fmt.Errorf("no stable Go versions found")
	}

	if requirement == "latest" {
		return goVersionClean(stable[0].Version), nil
	}

	constraint, err := semver.NewConstraint(requirement)
	if err != nil {
		return goVersionClean(stable[0].Version), nil
	}

	for _, gv := range stable {
		ver := goVersionClean(gv.Version)
		v, err := semver.NewVersion(ver)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			return ver, nil
		}
	}

	return goVersionClean(stable[0].Version), nil
}

func (g *GoInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	data, err := FetchJSON("https://go.dev/dl/?mode=json")
	if err != nil {
		return fmt.Errorf("failed to fetch Go versions: %w", err)
	}

	var versions []goVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return err
	}

	// Find the version
	goVer := "go" + version
	var target *goVersion
	for i, v := range versions {
		if v.Version == goVer {
			target = &versions[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("Go %s not found in release list", version)
	}

	// Find the archive for our platform
	var file *goFile
	for i, f := range target.Files {
		if f.OS == runtime.GOOS && f.Arch == runtime.GOARCH && f.Kind == "archive" {
			file = &target.Files[i]
			break
		}
	}

	if file == nil {
		return fmt.Errorf("no Go %s archive found for %s/%s", version, runtime.GOOS, runtime.GOARCH)
	}

	downloadURL := "https://go.dev/dl/" + file.Filename
	tmpFile := filepath.Join(os.TempDir(), file.Filename)
	defer os.Remove(tmpFile)

	if err := DownloadFile(downloadURL, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download Go: %w", err)
	}

	if file.SHA256 != "" {
		if err := VerifyChecksum(tmpFile, file.SHA256); err != nil {
			return fmt.Errorf("Go checksum verification failed: %w", err)
		}
	}

	// Go archives have a "go/" top-level directory
	if err := ExtractAndFlatten(tmpFile, targetDir); err != nil {
		return fmt.Errorf("failed to extract Go: %w", err)
	}

	return nil
}

func (g *GoInstaller) BinDir(installDir string) string {
	return filepath.Join(installDir, "bin")
}

func (g *GoInstaller) EnvVars(installDir string) map[string]string {
	return map[string]string{
		"GOROOT": installDir,
	}
}

// goVersionClean removes the "go" prefix from version strings.
func goVersionClean(v string) string {
	return strings.TrimPrefix(v, "go")
}
