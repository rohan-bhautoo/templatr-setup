# Release Process

This document describes how releases are built, published, and distributed for `templatr-setup`.

## Overview

Releases are fully automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. Pushing a semver tag triggers the pipeline which:

1. Builds the embedded web UI (`cd web && npm ci && npm run build`)
2. Cross-compiles Go binaries for 6 targets (3 OS x 2 architectures)
3. Generates SHA256 checksums for all archives
4. Creates a GitHub Release with binaries, archives, and auto-generated changelog
5. Updates the Homebrew tap formula (`rohan-bhautoo/homebrew-tap`)
6. Updates the Scoop bucket manifest (`rohan-bhautoo/scoop-bucket`)
7. Submits a winget manifest (`rohan-bhautoo/winget-pkgs`)

## Creating a Release

```bash
# 1. Ensure you're on master with a clean working tree
git checkout master
git pull origin master
git status

# 2. Tag the release (follows semver)
git tag -a v1.0.0 -m "v1.0.0"

# 3. Push the tag - this triggers the release workflow
git push origin v1.0.0
```

The GitHub Actions workflow (`.github/workflows/release.yml`) handles everything from here. You do not need GoReleaser installed locally.

### Versioning Guidelines

| Tag      | When to use                         | Example                                          |
| -------- | ----------------------------------- | ------------------------------------------------ |
| `v1.0.0` | First stable release                | Initial public release                           |
| `v1.1.0` | New features (backwards compatible) | Added Ruby installer, new TUI theme              |
| `v1.1.1` | Bug fixes                           | Fixed Python detection on Windows                |
| `v2.0.0` | Breaking changes                    | Changed manifest format, removed deprecated flag |

**Version is determined by the tag name.** GoReleaser reads the tag (e.g., `v1.0.0`) and injects it into the binary via ldflags. There is no version file to update manually. Development builds show `dev` as the version.

## Build Targets

GoReleaser produces 6 binaries, all statically linked (`CGO_ENABLED=0`) and stripped (`-s -w` ldflags):

| OS      | Architecture | Archive Format | Archive Name                               |
| ------- | ------------ | -------------- | ------------------------------------------ |
| Windows | amd64        | .zip           | `templatr-setup_X.Y.Z_windows_amd64.zip`   |
| Windows | arm64        | .zip           | `templatr-setup_X.Y.Z_windows_arm64.zip`   |
| Linux   | amd64        | .tar.gz        | `templatr-setup_X.Y.Z_linux_amd64.tar.gz`  |
| Linux   | arm64        | .tar.gz        | `templatr-setup_X.Y.Z_linux_arm64.tar.gz`  |
| macOS   | amd64        | .tar.gz        | `templatr-setup_X.Y.Z_darwin_amd64.tar.gz` |
| macOS   | arm64        | .tar.gz        | `templatr-setup_X.Y.Z_darwin_arm64.tar.gz` |

The web UI is built first by GoReleaser's `before.hooks` (`npm --prefix web ci` + `npm --prefix web run build`), then embedded into all 6 binaries via Go's `embed` package.

## Checksums

Every release includes a `checksums.txt` file with SHA256 hashes for all archives.

### Verifying Downloads

**Linux / macOS:**

```bash
# Download the checksums file
curl -LO https://github.com/rohan-bhautoo/templatr-setup/releases/download/v1.0.0/checksums.txt

# Verify your download
sha256sum -c checksums.txt --ignore-missing
```

**Windows (PowerShell):**

```powershell
# Get the hash of the downloaded file
Get-FileHash templatr-setup_1.0.0_windows_amd64.zip -Algorithm SHA256

# Compare against the value in checksums.txt
```

## Version Injection

GoReleaser injects version info into the binary via ldflags at build time:

```go
// main.go
var (
    version = "dev"     // Set to tag version (e.g., "1.0.0") by GoReleaser
    commit  = "none"    // Set to short commit hash
    date    = "unknown" // Set to build date (RFC3339)
)
```

These are passed to `cmd.SetVersionInfo(version, commit, date)` which stores them for the `version` command and self-update comparisons. Development builds show `dev` as the version and skip update checks.

## Required GitHub Secrets

Configure these in the repository **Settings > Secrets and variables > Actions**:

| Secret             | Purpose                                                     | Notes                        |
| ------------------ | ----------------------------------------------------------- | ---------------------------- |
| `GITHUB_TOKEN`     | Creates the GitHub Release, uploads assets                  | Built-in, no setup needed    |
| `TAP_GITHUB_TOKEN` | Pushes to homebrew-tap, scoop-bucket, and winget-pkgs repos | Fine-grained PAT (see below) |

### Creating the TAP_GITHUB_TOKEN

1. Go to [github.com/settings/tokens](https://github.com/settings/tokens) > **Fine-grained tokens** > **Generate new token**
2. Set:
   - **Token name**: `templatr-setup-release`
   - **Expiration**: 1 year (set a reminder to rotate)
   - **Repository access**: Only select repositories:
     - `rohan-bhautoo/homebrew-tap`
     - `rohan-bhautoo/scoop-bucket`
     - `rohan-bhautoo/winget-pkgs`
   - **Permissions**: Repository permissions > **Contents** > Read and write
3. Copy the token
4. In the `rohan-bhautoo/templatr-setup` repo, go to **Settings > Secrets and variables > Actions**
5. Add a new repository secret named `TAP_GITHUB_TOKEN` with the token value

Alternatively, use a **Classic token** with `repo` scope (simpler but broader access).

## Distribution Channels

### GitHub Releases (automatic)

Every tagged release creates a GitHub Release at:

```
https://github.com/rohan-bhautoo/templatr-setup/releases/latest
```

Direct download links for specific versions:

```
https://github.com/rohan-bhautoo/templatr-setup/releases/download/v1.0.0/templatr-setup_1.0.0_windows_amd64.zip
https://github.com/rohan-bhautoo/templatr-setup/releases/download/v1.0.0/templatr-setup_1.0.0_darwin_arm64.tar.gz
https://github.com/rohan-bhautoo/templatr-setup/releases/download/v1.0.0/templatr-setup_1.0.0_linux_amd64.tar.gz
```

The `/releases/latest` URL always redirects to the newest version. The Templatr website download page uses `/releases/latest/download/<filename>` URLs for direct binary downloads.

### Homebrew (automatic)

GoReleaser pushes an updated formula to `rohan-bhautoo/homebrew-tap` on every release.

**One-time setup** (already done):

```bash
gh repo create rohan-bhautoo/homebrew-tap --public --description "Homebrew tap for Templatr tools"
```

**User install command:**

```bash
brew install rohan-bhautoo/tap/templatr-setup
```

**User update command:**

```bash
brew upgrade templatr-setup
```

### Scoop (automatic)

GoReleaser pushes an updated manifest to `rohan-bhautoo/scoop-bucket` on every release.

**One-time setup** (already done):

```bash
gh repo create rohan-bhautoo/scoop-bucket --public --description "Scoop bucket for Templatr tools"
```

**User install commands:**

```bash
scoop bucket add templatr https://github.com/rohan-bhautoo/scoop-bucket
scoop install templatr-setup
```

**User update command:**

```bash
scoop update templatr-setup
```

### winget (semi-automatic)

GoReleaser pushes a manifest to `rohan-bhautoo/winget-pkgs`. For the official Microsoft winget repository, a PR needs to be submitted manually after the first release.

**One-time setup** (already done):

```bash
gh repo create rohan-bhautoo/winget-pkgs --public --description "winget manifests for Templatr tools"
```

**Submitting to official winget repository:**

After the first release, submit a PR from `rohan-bhautoo/winget-pkgs` to `microsoft/winget-pkgs`. Subsequent releases can use [wingetcreate](https://github.com/microsoft/winget-create) to automate updates.

**User install command:**

```bash
winget install Templatr.TemplatrSetup
```

## Changelog

Release notes are auto-generated from commit messages between tags. Commits are sorted ascending and filtered:

| Prefix      | Included in changelog |
| ----------- | --------------------- |
| `feat:`     | Yes                   |
| `fix:`      | Yes                   |
| `refactor:` | Yes                   |
| `docs:`     | No (excluded)         |
| `test:`     | No (excluded)         |
| `chore:`    | No (excluded)         |

Use conventional commit prefixes for meaningful release notes.

## CI Workflows

### CI Pipeline (`.github/workflows/ci.yml`)

Runs on every push and pull request:

1. **Build web UI**: Node.js 25, `npm --prefix web ci`, `npm --prefix web run build`
2. **Build Go**: `go build -v .` on Windows, macOS, and Linux
3. **Test**: `go test -v ./...` on all 3 OSes
4. **Lint**: `go vet ./...` on all 3 OSes

The master branch has protection rules requiring all 4 CI checks to pass plus 1 PR review.

### Release Pipeline (`.github/workflows/release.yml`)

Runs only on version tags (`v*`):

1. Checks out code with full history (`fetch-depth: 0` for changelog generation)
2. Sets up Node.js 25 and Go
3. Runs GoReleaser with `TAP_GITHUB_TOKEN` in environment
4. GoReleaser's before hooks build the web UI, then compiles, archives, and publishes

## Testing a Release Locally

Use GoReleaser's snapshot mode to validate the build without publishing:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Run a snapshot build
goreleaser release --snapshot --clean
```

This runs the full pipeline including web UI build, Go cross-compilation, archive creation, and checksum generation. All output goes to `dist/`. Nothing is published.

Verify the output:

```bash
ls dist/
# templatr-setup_X.Y.Z-SNAPSHOT-xxx_windows_amd64.zip
# templatr-setup_X.Y.Z-SNAPSHOT-xxx_darwin_arm64.tar.gz
# templatr-setup_X.Y.Z-SNAPSHOT-xxx_linux_amd64.tar.gz
# checksums.txt
# ...
```

## Code Signing (Deferred)

Code signing is not included for v1. Users installing via Homebrew, Scoop, or winget bypass OS security warnings entirely. Direct downloads require a one-time bypass:

- **Windows**: SmartScreen - click **"More info"** > **"Run anyway"**
- **macOS**: Gatekeeper - run `xattr -d com.apple.quarantine templatr-setup`
- **Linux**: No OS gatekeeper. No action needed.

The Templatr website download page includes step-by-step instructions with screenshots for these bypasses.

**Future signing options** (when revenue justifies the cost):

| Platform | Method                                               | Cost                |
| -------- | ---------------------------------------------------- | ------------------- |
| Windows  | Azure Artifact Signing or SignPath.io (free for OSS) | $9.99/month or free |
| macOS    | Apple Developer ID                                   | $99/year            |
| Linux    | GPG signing of checksums.txt                         | Free                |

## Self-Update Mechanism

The binary includes a self-update feature (`templatr-setup update`):

1. Uses `creativeprojects/go-selfupdate` to download the correct binary for the current OS/arch
2. Verifies SHA256 checksum against `checksums.txt` in the release
3. Replaces the running binary in-place

If the tool detects it was installed via a package manager (by checking the executable path for `homebrew`, `scoop`, or `winget` directories), it returns an error directing the user to their package manager instead (e.g., `brew upgrade templatr-setup`).

### Background Update Check

On every invocation, a non-blocking goroutine checks for updates:

1. Reads `~/.templatr/last_update_check` - skips if checked within the last 24 hours
2. Queries `https://api.github.com/repos/rohan-bhautoo/templatr-setup/releases/latest` (5-second timeout)
3. Caches the result to `~/.templatr/latest_version`
4. After the main command finishes, prints a notice if an update is available

This adds < 200ms overhead and never delays the main operation. If the API is unreachable (offline, rate-limited), it silently continues.
