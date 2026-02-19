package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	update "github.com/creativeprojects/go-selfupdate"
)

const (
	repoOwner       = "rohan-bhautoo"
	repoName        = "templatr-setup"
	checkCooldown   = 24 * time.Hour
	checkFile       = ".templatr/last_update_check"
	latestCacheFile = ".templatr/latest_version"
)

// CheckResult contains the result of an update check.
type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	UpdateAvail    bool
}

// CheckForUpdate checks GitHub for a newer release.
// Returns nil if no update is available, the check was done recently, or on error.
func CheckForUpdate(currentVersion string) *CheckResult {
	if currentVersion == "dev" || currentVersion == "" {
		return nil
	}

	// Check cooldown
	if !shouldCheck() {
		return readCachedResult(currentVersion)
	}

	// Record check time
	recordCheckTime()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latest, err := fetchLatestVersion(ctx)
	if err != nil {
		return nil
	}

	// Cache the result
	cacheResult(latest)

	current, err := semver.NewVersion(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return nil
	}

	latestVer, err := semver.NewVersion(strings.TrimPrefix(latest, "v"))
	if err != nil {
		return nil
	}

	return &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		UpdateAvail:    latestVer.GreaterThan(current),
	}
}

// DoUpdate performs the self-update to the latest version.
func DoUpdate(currentVersion string) error {
	if currentVersion == "dev" {
		return fmt.Errorf("cannot update a development build - install a release version")
	}

	// Check if installed via package manager
	if hint := detectPackageManager(); hint != "" {
		return fmt.Errorf("it looks like you installed via %s. Update using your package manager instead", hint)
	}

	source, err := update.NewGitHubSource(update.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("failed to create update source: %w", err)
	}

	updater, err := update.NewUpdater(update.Config{
		Source:    source,
		Validator: &update.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), update.ParseSlug(repoOwner+"/"+repoName))
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	if !found {
		return fmt.Errorf("no releases found")
	}

	current, err := semver.NewVersion(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return fmt.Errorf("cannot parse current version %q: %w", currentVersion, err)
	}

	latestVer, err := semver.NewVersion(strings.TrimPrefix(latest.Version(), "v"))
	if err != nil {
		return fmt.Errorf("cannot parse latest version: %w", err)
	}

	if !latestVer.GreaterThan(current) {
		return fmt.Errorf("already up to date (v%s)", currentVersion)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// shouldCheck returns true if enough time has passed since last check.
func shouldCheck() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true
	}

	path := filepath.Join(home, checkFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return true
	}

	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		return true
	}

	return time.Since(t) > checkCooldown
}

func recordCheckTime() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	path := filepath.Join(home, checkFile)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339)), 0o644)
}

func cacheResult(version string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	path := filepath.Join(home, latestCacheFile)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(version), 0o644)
}

func readCachedResult(currentVersion string) *CheckResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(home, latestCacheFile))
	if err != nil {
		return nil
	}

	latest := strings.TrimSpace(string(data))
	if latest == "" {
		return nil
	}

	current, err := semver.NewVersion(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return nil
	}
	latestVer, err := semver.NewVersion(strings.TrimPrefix(latest, "v"))
	if err != nil {
		return nil
	}

	result := &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		UpdateAvail:    latestVer.GreaterThan(current),
	}

	if !result.UpdateAvail {
		return nil
	}
	return result
}

// fetchLatestVersion makes a lightweight API call to get the latest release tag.
func fetchLatestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// detectPackageManager checks if the binary was installed via a package manager.
func detectPackageManager() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	exe = strings.ToLower(exe)

	if strings.Contains(exe, "homebrew") || strings.Contains(exe, "linuxbrew") || strings.Contains(exe, "/cellar/") {
		return "Homebrew - run 'brew upgrade templatr-setup'"
	}
	if strings.Contains(exe, "scoop") {
		return "Scoop - run 'scoop update templatr-setup'"
	}
	if strings.Contains(exe, "winget") || strings.Contains(exe, "windowsapps") {
		return "winget - run 'winget upgrade Templatr.TemplatrSetup'"
	}

	return ""
}
