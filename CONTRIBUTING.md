# Contributing

## Project Overview

`templatr-setup` is a cross-platform CLI tool written in Go that reads `.templatr.toml` manifest files from Templatr templates and automatically installs required runtimes, packages, and configuration.

## Key Technologies

- **Language**: Go 1.26+
- **CLI**: Cobra (spf13/cobra)
- **TUI**: Bubbletea + Lipgloss + Bubbles (Charmbracelet suite)
- **TOML**: pelletier/go-toml/v2
- **Semver**: Masterminds/semver/v3
- **Self-update**: creativeprojects/go-selfupdate
- **Config writer**: Regex-based .env and TypeScript editor
- **Build/Release**: GoReleaser

## Development Commands

```bash
# Build
go build -o templatr-setup .

# Run
go run . doctor
go run . setup --dry-run
go run . setup -f examples/nextjs.templatr.toml
go run . configure -f examples/nextjs.templatr.toml
go run . version
go run . logs

# Test
go test ./...
go vet ./...

# Tidy dependencies
go mod tidy
```

## CLI Commands

| Command                          | Description                                            |
| -------------------------------- | ------------------------------------------------------ |
| `templatr-setup`                 | Auto-detect: TUI if terminal, web UI if double-clicked |
| `templatr-setup setup`           | Install runtimes and packages from .templatr.toml      |
| `templatr-setup setup --dry-run` | Show what would be installed without changes           |
| `templatr-setup setup -y`        | Skip confirmation prompt                               |
| `templatr-setup configure`       | Run only the configure step (.env + site.ts)           |
| `templatr-setup doctor`          | Check system: installed runtimes, versions, PATH       |
| `templatr-setup uninstall`       | Remove runtimes installed by this tool                 |
| `templatr-setup update`          | Self-update to latest version from GitHub Releases     |
| `templatr-setup logs`            | Show recent log files                                  |
| `templatr-setup version`         | Show version and check for updates                     |

## Directory Structure

```
cmd/                    # Cobra CLI commands
internal/
├── manifest/           # TOML parser + validation
├── detect/             # OS and runtime detection
├── engine/             # Setup plan builder + display
├── install/            # Runtime installers (node, python, flutter, etc.)
├── packages/           # Package manager integration (npm, pip, pub, etc.)
├── config/             # Config file writer (.env, TypeScript)
├── state/              # State tracking + undo (reverse installation)
├── selfupdate/         # GitHub Releases self-update + version check
├── logger/             # File + stdout logging with secret masking
├── tui/                # Bubbletea terminal UI
└── server/             # Local web server for --ui mode (planned)
web/                    # React SPA embedded into binary (planned)
examples/               # Example .templatr.toml files
```

## Architecture

- Auto-detects terminal vs double-click: TUI for devs, web UI for non-devs
- Downloads runtimes from official sources, verifies SHA256 checksums
- Installs to `~/.templatr/runtimes/<name>/<version>/` (user-space, no root required)
- Tracks installations in `~/.templatr/state.json` for clean uninstall
- Previous versions are recorded so uninstall can show revert info
- Environment variables (JAVA_HOME, GOROOT) tracked and cleaned up on uninstall
- Logs to `~/.templatr/logs/` with automatic rotation (keeps last 10)
- Non-blocking update check on startup with 24h cooldown

## Adding a New Runtime Installer

1. Create `internal/install/<runtime>.go`
2. Implement the `Installer` interface:
   ```go
   type Installer interface {
       Name() string
       ResolveVersion(requirement string) (string, error)
       Install(version, targetDir string, progress ProgressFunc) error
       BinDir(installDir string) string
       EnvVars(installDir string) map[string]string
   }
   ```
3. Register it in `internal/install/installer.go` `init()`:
   ```go
   Register(&MyRuntimeInstaller{})
   ```
4. Add the runtime key to `internal/engine/plan.go` display/detect name maps
5. Add to `internal/manifest/validate.go` `validRuntimes` set
6. Add to `internal/detect/runtime.go` runtime checks

## Releasing

Releases are automated via GitHub Actions and GoReleaser.

**How it works:**

1. You create a git tag with a semver version
2. Push the tag to GitHub
3. The `release.yml` workflow triggers automatically
4. GoReleaser cross-compiles 6 binaries (Windows/Linux/macOS x amd64/arm64)
5. A GitHub Release is created with binaries, checksums, and auto-generated changelog
6. Homebrew/Scoop/winget formulas are updated automatically

**Steps:**

```bash
# Make sure you're on master with everything committed
git status

# Create a version tag (follow semver: major.minor.patch)
git tag v1.0.0

# Push the tag — this triggers the release workflow
git push origin v1.0.0
```

**Version is determined by the tag name.** GoReleaser reads the tag (e.g., `v1.0.0`) and injects it into the binary via ldflags (`-X main.version=1.0.0`). There is no version file to update manually.

**Versioning guidelines:**

- `v1.0.0` — First stable release
- `v1.1.0` — New features (backwards compatible)
- `v1.1.1` — Bug fixes
- `v2.0.0` — Breaking changes

## Conventions

- Use `internal/` for all non-exported packages
- All runtime installers implement the `Installer` interface
- Error messages must be actionable (what failed, why, how to fix)
- Secret values (.env secrets) are NEVER logged
- Tests use `t.TempDir()` for file operations
- Use `filepath.Join` (not hardcoded slashes) for cross-platform paths
