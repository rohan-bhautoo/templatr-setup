# Release Process

This document describes how releases are built, published, and distributed.

## Overview

Releases are fully automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. Pushing a version tag triggers the pipeline which:

1. Builds the embedded web UI (React/Vite)
2. Cross-compiles Go binaries for 6 targets (3 OS x 2 arch)
3. Generates SHA256 checksums
4. Creates a GitHub Release with binaries and changelog
5. Updates the Homebrew tap formula
6. Updates the Scoop bucket manifest
7. Submits a winget manifest PR

## Creating a Release

```bash
# 1. Ensure you're on master with a clean working tree
git checkout master
git pull origin master

# 2. Tag the release (follows semver)
git tag -a v1.0.0 -m "v1.0.0"

# 3. Push the tag — this triggers the release workflow
git push origin v1.0.0
```

The GitHub Actions workflow handles everything from here.

## Build Targets

| OS      | Architecture | Archive Format |
| ------- | ------------ | -------------- |
| Windows | amd64        | .zip           |
| Windows | arm64        | .zip           |
| Linux   | amd64        | .tar.gz        |
| Linux   | arm64        | .tar.gz        |
| macOS   | amd64        | .tar.gz        |
| macOS   | arm64        | .tar.gz        |

All binaries are statically linked (`CGO_ENABLED=0`) and stripped (`-s -w`).

## Checksums

Every release includes a `checksums.txt` file with SHA256 hashes for all archives. Users can verify downloads:

```bash
# Linux/macOS
sha256sum -c checksums.txt --ignore-missing

# Windows (PowerShell)
Get-FileHash templatr-setup_1.0.0_windows_amd64.zip -Algorithm SHA256
```

## Required GitHub Secrets

Configure these in the repository Settings > Secrets and variables > Actions:

| Secret             | Purpose                                               | Scope                          |
| ------------------ | ----------------------------------------------------- | ------------------------------ |
| `GITHUB_TOKEN`     | Automatic — creates the GitHub Release                | Built-in (no setup)            |
| `TAP_GITHUB_TOKEN` | PAT for pushing to homebrew-tap, scoop-bucket, winget | `repo` scope on `templatr` org |

### Creating the TAP_GITHUB_TOKEN

1. Go to https://github.com/settings/tokens (or use Fine-grained tokens)
2. Create a new Personal Access Token with:
   - **Classic token**: `repo` scope (full control of private repositories)
   - **Fine-grained token**: Read/Write access to `Contents` on:
     - `templatr/homebrew-tap`
     - `templatr/scoop-bucket`
     - `templatr/winget-pkgs`
3. Add it as a repository secret named `TAP_GITHUB_TOKEN` in `templatr/templatr-setup`

## Distribution Channels

### GitHub Releases (automatic)

Direct download links follow this pattern:

```
https://github.com/templatr/templatr-setup/releases/latest
https://github.com/templatr/templatr-setup/releases/download/v1.0.0/templatr-setup_1.0.0_windows_amd64.zip
https://github.com/templatr/templatr-setup/releases/download/v1.0.0/templatr-setup_1.0.0_darwin_arm64.tar.gz
```

The `/releases/latest` URL always redirects to the newest version.

### Homebrew (automatic)

GoReleaser pushes an updated formula to `templatr/homebrew-tap` on every release.

**One-time setup** — create the `templatr/homebrew-tap` repository:

```bash
# Create the repo on GitHub (public, with a README)
gh repo create templatr/homebrew-tap --public --description "Homebrew tap for Templatr tools"
```

Users install via:

```bash
brew install templatr/tap/templatr-setup
```

### Scoop (automatic)

GoReleaser pushes an updated manifest to `templatr/scoop-bucket` on every release.

**One-time setup** — create the `templatr/scoop-bucket` repository:

```bash
# Create the repo on GitHub (public, with a README)
gh repo create templatr/scoop-bucket --public --description "Scoop bucket for Templatr tools"
```

Users install via:

```bash
scoop bucket add templatr https://github.com/templatr/scoop-bucket
scoop install templatr-setup
```

### winget (automatic)

GoReleaser pushes a manifest to `templatr/winget-pkgs` which can then be submitted to the official `microsoft/winget-pkgs` repository.

**One-time setup** — create the `templatr/winget-pkgs` repository:

```bash
gh repo create templatr/winget-pkgs --public --description "winget manifests for Templatr tools"
```

**Submitting to official winget repository:**
After the first release, manually submit a PR from `templatr/winget-pkgs` to `microsoft/winget-pkgs`. Subsequent releases can use [wingetcreate](https://github.com/microsoft/winget-create) to automate updates.

Users install via:

```bash
winget install Templatr.TemplatrSetup
```

## Version Injection

GoReleaser injects version info into the binary via ldflags:

```go
// main.go
var (
    version = "dev"     // Set to tag version (e.g., "1.0.0")
    commit  = "none"    // Set to short commit hash
    date    = "unknown" // Set to build date
)
```

Development builds show `dev` as the version. Release builds show the actual version from the git tag.

## Changelog

Release notes are auto-generated from commit messages. Commits prefixed with `docs:`, `test:`, or `chore:` are excluded. Use conventional commit prefixes:

- `feat:` — New features (included in changelog)
- `fix:` — Bug fixes (included in changelog)
- `refactor:` — Code changes (included in changelog)
- `docs:` — Documentation only (excluded)
- `test:` — Test changes (excluded)
- `chore:` — Maintenance (excluded)

## Code Signing (Deferred)

Code signing is not included in v1. Users installing via Homebrew, Scoop, or winget bypass OS security warnings. Direct downloads require a one-time bypass:

- **Windows**: SmartScreen — click "More info" > "Run anyway"
- **macOS**: Gatekeeper — run `xattr -d com.apple.quarantine templatr-setup`

Future signing options when revenue justifies the cost:

- Windows: Azure Artifact Signing ($9.99/month) or SignPath.io (free for OSS)
- macOS: Apple Developer ID ($99/year)
- Linux: GPG signing of checksums.txt (free)

## Testing a Release Locally

Use GoReleaser's snapshot mode to test the build without publishing:

```bash
# Build web UI first
cd web && npm ci && npm run build && cd ..

# Dry-run release (no publish)
goreleaser release --snapshot --clean
```

This creates all archives in the `dist/` directory without pushing anything.

## CI Workflow

The CI pipeline (`.github/workflows/ci.yml`) runs on every push and PR:

1. **Test** — Builds web UI, compiles Go, runs tests on all 3 OSes
2. **Lint** — Runs `go vet` for static analysis

The release pipeline (`.github/workflows/release.yml`) runs only on version tags (`v*`).
