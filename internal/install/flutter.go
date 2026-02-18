package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Masterminds/semver/v3"
)

// FlutterInstaller handles Flutter SDK installation from flutter.dev.
type FlutterInstaller struct{}

func (f *FlutterInstaller) Name() string { return "flutter" }

// flutterReleases is the top-level structure of the Flutter releases JSON.
type flutterReleases struct {
	BaseURL        string           `json:"base_url"`
	CurrentRelease flutterCurrent   `json:"current_release"`
	Releases       []flutterRelease `json:"releases"`
}

type flutterCurrent struct {
	Stable string `json:"stable"`
}

type flutterRelease struct {
	Hash    string `json:"hash"`
	Channel string `json:"channel"`
	Version string `json:"version"`
	Archive string `json:"archive"`
	SHA256  string `json:"sha256"`
}

func (f *FlutterInstaller) ResolveVersion(requirement string) (string, error) {
	platform := flutterPlatform()
	url := fmt.Sprintf("https://storage.googleapis.com/flutter_infra_release/releases/releases_%s.json", platform)

	data, err := FetchJSON(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Flutter releases: %w", err)
	}

	var releases flutterReleases
	if err := json.Unmarshal(data, &releases); err != nil {
		return "", fmt.Errorf("failed to parse Flutter releases: %w", err)
	}

	// Filter to stable releases
	var stable []flutterRelease
	for _, r := range releases.Releases {
		if r.Channel == "stable" {
			stable = append(stable, r)
		}
	}

	if len(stable) == 0 {
		return "", fmt.Errorf("no stable Flutter releases found")
	}

	if requirement == "latest" {
		return stable[0].Version, nil
	}

	constraint, err := semver.NewConstraint(requirement)
	if err != nil {
		return stable[0].Version, nil
	}

	for _, r := range stable {
		v, err := semver.NewVersion(r.Version)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			return r.Version, nil
		}
	}

	return stable[0].Version, nil
}

func (f *FlutterInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	platform := flutterPlatform()
	url := fmt.Sprintf("https://storage.googleapis.com/flutter_infra_release/releases/releases_%s.json", platform)

	data, err := FetchJSON(url)
	if err != nil {
		return fmt.Errorf("failed to fetch Flutter releases: %w", err)
	}

	var releases flutterReleases
	if err := json.Unmarshal(data, &releases); err != nil {
		return err
	}

	// Find the specific release
	var target *flutterRelease
	for i, r := range releases.Releases {
		if r.Version == version && r.Channel == "stable" {
			target = &releases.Releases[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("Flutter %s not found in stable releases", version)
	}

	downloadURL := releases.BaseURL + "/" + target.Archive
	filename := filepath.Base(target.Archive)

	tmpFile := filepath.Join(os.TempDir(), filename)
	defer os.Remove(tmpFile)

	if err := DownloadFile(downloadURL, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download Flutter: %w", err)
	}

	// Verify checksum
	if target.SHA256 != "" {
		if err := VerifyChecksum(tmpFile, target.SHA256); err != nil {
			return fmt.Errorf("Flutter checksum verification failed: %w", err)
		}
	}

	// Extract — Flutter archive has a "flutter/" top-level dir
	if err := ExtractAndFlatten(tmpFile, targetDir); err != nil {
		return fmt.Errorf("failed to extract Flutter: %w", err)
	}

	return nil
}

func (f *FlutterInstaller) BinDir(installDir string) string {
	return filepath.Join(installDir, "bin")
}

func (f *FlutterInstaller) EnvVars(installDir string) map[string]string { return nil }

func flutterPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

