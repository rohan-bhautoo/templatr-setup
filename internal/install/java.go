package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// JavaInstaller handles Java (Adoptium Temurin) installation.
type JavaInstaller struct{}

func (j *JavaInstaller) Name() string { return "java" }

// adoptiumAsset represents an Adoptium API response entry.
type adoptiumAsset struct {
	Binary  adoptiumBinary `json:"binary"`
	Version adoptiumVer    `json:"version"`
}

type adoptiumBinary struct {
	Architecture string      `json:"architecture"`
	ImageType    string      `json:"image_type"`
	OS           string      `json:"os"`
	Package      adoptiumPkg `json:"package"`
}

type adoptiumPkg struct {
	Checksum string `json:"checksum"`
	Link     string `json:"link"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
}

type adoptiumVer struct {
	Major    int    `json:"major"`
	Minor    int    `json:"minor"`
	Security int    `json:"security"`
	Semver   string `json:"semver"`
}

func (j *JavaInstaller) ResolveVersion(requirement string) (string, error) {
	// Determine the major version to fetch
	major := 21 // default to latest LTS

	if requirement != "latest" {
		constraint, err := semver.NewConstraint(requirement)
		if err == nil {
			// Try to determine the major version from the constraint
			// Common patterns: ">=21", ">=17", ">=11"
			for _, m := range []int{25, 24, 23, 22, 21, 17, 11, 8} {
				v, _ := semver.NewVersion(fmt.Sprintf("%d.0.0", m))
				if constraint.Check(v) {
					major = m
					break
				}
			}
		} else {
			// Try parsing as a plain major version
			parts := strings.Split(requirement, ".")
			clean := strings.TrimLeft(parts[0], "><=~^")
			if m, err := strconv.Atoi(clean); err == nil {
				major = m
			}
		}
	}

	// Fetch from Adoptium API
	apiURL := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%d/hotspot?architecture=%s&image_type=jdk&os=%s&vendor=eclipse",
		major, javaArch(), javaOS())

	data, err := FetchJSON(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Adoptium releases: %w", err)
	}

	var assets []adoptiumAsset
	if err := json.Unmarshal(data, &assets); err != nil {
		return "", fmt.Errorf("failed to parse Adoptium response: %w", err)
	}

	if len(assets) == 0 {
		return "", fmt.Errorf("no Adoptium JDK %d found for %s/%s", major, javaOS(), javaArch())
	}

	return assets[0].Version.Semver, nil
}

func (j *JavaInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	// Determine major version from the version string
	parts := strings.Split(version, ".")
	major := parts[0]

	apiURL := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&image_type=jdk&os=%s&vendor=eclipse",
		major, javaArch(), javaOS())

	data, err := FetchJSON(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Adoptium releases: %w", err)
	}

	var assets []adoptiumAsset
	if err := json.Unmarshal(data, &assets); err != nil {
		return err
	}

	if len(assets) == 0 {
		return fmt.Errorf("no Adoptium JDK found")
	}

	asset := assets[0]
	tmpFile := filepath.Join(os.TempDir(), asset.Binary.Package.Name)
	defer os.Remove(tmpFile)

	if err := DownloadFile(asset.Binary.Package.Link, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download Java: %w", err)
	}

	// Verify checksum
	if asset.Binary.Package.Checksum != "" {
		if err := VerifyChecksum(tmpFile, asset.Binary.Package.Checksum); err != nil {
			return fmt.Errorf("Java checksum verification failed: %w", err)
		}
	}

	// Extract - Adoptium archives have a top-level jdk-* dir
	if err := ExtractAndFlatten(tmpFile, targetDir); err != nil {
		return fmt.Errorf("failed to extract Java: %w", err)
	}

	return nil
}

func (j *JavaInstaller) BinDir(installDir string) string {
	return filepath.Join(installDir, "bin")
}

func (j *JavaInstaller) EnvVars(installDir string) map[string]string {
	return map[string]string{
		"JAVA_HOME": installDir,
	}
}

func javaOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "mac"
	default:
		return runtime.GOOS // "linux", "windows"
	}
}

func javaArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}
