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

// NodeInstaller handles Node.js installation from nodejs.org.
type NodeInstaller struct{}

func (n *NodeInstaller) Name() string { return "node" }

// nodeRelease represents an entry in the Node.js dist/index.json.
type nodeRelease struct {
	Version string      `json:"version"` // "v22.14.0"
	Date    string      `json:"date"`
	Files   []string    `json:"files"`
	LTS     interface{} `json:"lts"` // false or string like "Jod"
}

func (n *NodeInstaller) ResolveVersion(requirement string) (string, error) {
	data, err := FetchJSON("https://nodejs.org/dist/index.json")
	if err != nil {
		return "", fmt.Errorf("failed to fetch Node.js versions: %w", err)
	}

	var releases []nodeRelease
	if err := json.Unmarshal(data, &releases); err != nil {
		return "", fmt.Errorf("failed to parse Node.js versions: %w", err)
	}

	// Filter to LTS versions only
	var ltsReleases []nodeRelease
	for _, r := range releases {
		if r.LTS != nil && r.LTS != false {
			// LTS field is a string when it's an LTS version
			if _, ok := r.LTS.(string); ok {
				ltsReleases = append(ltsReleases, r)
			}
		}
	}

	if len(ltsReleases) == 0 {
		return "", fmt.Errorf("no LTS versions found")
	}

	if requirement == "latest" {
		return strings.TrimPrefix(ltsReleases[0].Version, "v"), nil
	}

	// Find the latest LTS version that satisfies the constraint
	constraint, err := semver.NewConstraint(requirement)
	if err != nil {
		// If we can't parse the constraint, just return the latest LTS
		return strings.TrimPrefix(ltsReleases[0].Version, "v"), nil
	}

	for _, r := range ltsReleases {
		ver := strings.TrimPrefix(r.Version, "v")
		v, err := semver.NewVersion(ver)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			return ver, nil
		}
	}

	// No matching version found, return the latest LTS
	return strings.TrimPrefix(ltsReleases[0].Version, "v"), nil
}

func (n *NodeInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	osName := nodeOS()
	arch := nodeArch()
	ext := PlatformExt()
	filename := fmt.Sprintf("node-v%s-%s-%s.%s", version, osName, arch, ext)
	downloadURL := fmt.Sprintf("https://nodejs.org/dist/v%s/%s", version, filename)
	checksumURL := fmt.Sprintf("https://nodejs.org/dist/v%s/SHASUMS256.txt", version)

	// Download to temp file
	tmpFile := filepath.Join(os.TempDir(), filename)
	defer os.Remove(tmpFile)

	if err := DownloadFile(downloadURL, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download Node.js: %w", err)
	}

	// Verify checksum
	expectedHash, err := FetchChecksumFromURL(checksumURL, filename)
	if err != nil {
		return fmt.Errorf("failed to fetch Node.js checksum: %w", err)
	}
	if err := VerifyChecksum(tmpFile, expectedHash); err != nil {
		return fmt.Errorf("Node.js checksum verification failed: %w", err)
	}

	// Extract and flatten (strips the top-level node-vX.Y.Z-os-arch/ dir)
	if err := ExtractAndFlatten(tmpFile, targetDir); err != nil {
		return fmt.Errorf("failed to extract Node.js: %w", err)
	}

	return nil
}

func (n *NodeInstaller) BinDir(installDir string) string {
	if runtime.GOOS == "windows" {
		return installDir // node.exe is in the root on Windows
	}
	return filepath.Join(installDir, "bin")
}

func (n *NodeInstaller) EnvVars(installDir string) map[string]string { return nil }

// nodeOS returns the OS name used in Node.js download URLs.
func nodeOS() string {
	switch runtime.GOOS {
	case "windows":
		return "win"
	default:
		return runtime.GOOS // "linux", "darwin"
	}
}

// nodeArch returns the architecture name used in Node.js download URLs.
func nodeArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "arm64"
	case "386":
		return "x86"
	default:
		return runtime.GOARCH
	}
}
