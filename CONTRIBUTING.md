# Contributing to templatr-setup

Contributions are welcome! Please open an issue first to discuss what you'd like to change, then submit a pull request.

## Prerequisites

| Tool        | Version | Install                          | Purpose                    |
| ----------- | ------- | -------------------------------- | -------------------------- |
| **Go**      | 1.25+   | [go.dev/dl](https://go.dev/dl)   | Compiles the CLI tool      |
| **Node.js** | 20+     | [nodejs.org](https://nodejs.org) | Builds the embedded web UI |
| Git         | any     | -                                | Version control            |

GoReleaser is only needed in CI for release builds. You do not need it for local development.

## Development Commands

```bash
# Build the web UI (required before building the Go binary)
cd web && npm install && npm run build && cd ..

# Build the binary
go build -o templatr-setup .

# Run commands directly (without building)
go run . doctor
go run . setup --dry-run
go run . setup -f examples/nextjs.templatr.toml
go run . setup -y -f examples/nextjs.templatr.toml
go run . configure -f examples/nextjs.templatr.toml
go run . --ui -f examples/nextjs.templatr.toml
go run . version
go run . logs
go run . uninstall

# Run tests
go test ./...

# Static analysis
go vet ./...

# Tidy dependencies
go mod tidy

# Web UI development (hot reload)
cd web && npm run dev
```

> **Note**: The web UI must be built before building the Go binary because the Go binary embeds the `web/dist/` directory at compile time via `//go:embed all:web/dist` in `embed.go`.

## Project Structure

```
templatr-setup/
├── main.go                     # Entry point - sets version info, web assets, calls cmd.Execute()
├── embed.go                    # //go:embed all:web/dist - embeds built React app into binary
├── .goreleaser.yaml            # GoReleaser config for cross-platform builds + package managers
├── go.mod / go.sum             # Go module (github.com/templatr/templatr-setup)
│
├── cmd/                        # Cobra CLI commands
│   ├── root.go                 # Root command, --ui/--file flags, auto-detect logic, update check
│   ├── setup.go                # setup command - full flow (parse → detect → install → configure)
│   ├── configure.go            # configure command - standalone .env + site.ts config
│   ├── doctor.go               # doctor command - system info + all detected runtimes
│   ├── uninstall.go            # uninstall command - reverse installations from state.json
│   ├── update.go               # update command - self-update via GitHub Releases
│   ├── version.go              # version command - show version + check for updates
│   ├── logs.go                 # logs command - list recent log files
│   ├── console_windows.go      # Windows: AttachConsole for CLI mode when built with -H windowsgui
│   └── console_other.go        # Unix: no-op stub (build tag !windows)
│
├── internal/
│   ├── manifest/               # TOML parser + validation
│   │   ├── schema.go           # Go structs: Manifest, TemplateInfo, PackageConfig, EnvVar, ConfigFile, etc.
│   │   ├── parser.go           # Load(path) and Parse(bytes) - uses pelletier/go-toml/v2
│   │   └── validate.go         # Validate(m) - checks all fields, returns []error
│   │
│   ├── detect/                 # System and runtime detection
│   │   ├── os.go               # GetSystemInfo() - OS, arch, home dir
│   │   └── runtime.go          # ScanRuntimes() - checks 17 runtimes, handles Windows Store stubs
│   │
│   ├── engine/                 # Setup plan builder + display
│   │   ├── plan.go             # BuildPlan(m) - compares manifest requirements vs installed runtimes
│   │   └── display.go          # PrintSummary(plan) - formatted ASCII table output
│   │
│   ├── install/                # Runtime installers + download engine
│   │   ├── installer.go        # Installer interface, registry, ExecutePlan(), InstallSingleRuntime()
│   │   ├── download.go         # DownloadFile, VerifyChecksum, ExtractTarGz/Zip/AndFlatten, RuntimesDir
│   │   ├── path.go             # AddToPath, RemoveFromPath, SetEnvVar, RemoveEnvVar (Unix + Windows)
│   │   ├── node.go             # Node.js installer - nodejs.org dist API, SHASUMS256 verification
│   │   ├── python.go           # Python installer - python-build-standalone from GitHub releases
│   │   ├── flutter.go          # Flutter installer - flutter.dev releases JSON, SHA256 verification
│   │   ├── java.go             # Java installer - Adoptium API v3, sets JAVA_HOME
│   │   ├── go_runtime.go       # Go installer - go.dev API, sets GOROOT
│   │   ├── rust.go             # Rust installer - rustup-init with custom CARGO_HOME/RUSTUP_HOME
│   │   ├── ruby.go             # Ruby installer - stub, returns manual install instructions
│   │   ├── php.go              # PHP installer - stub, returns manual install instructions
│   │   └── dotnet.go           # .NET installer - stub, returns manual install instructions
│   │
│   ├── packages/               # Package manager integration
│   │   └── manager.go          # RunInstall, RunGlobalInstalls, RunPostSetup
│   │
│   ├── config/                 # Config file writers
│   │   ├── env.go              # WriteEnvFile, ReadEnvFile - .env with comments, quoting, secret masking
│   │   └── typescript.go       # UpdateConfigFile - regex-based key:value replacement, preserves quote style
│   │
│   ├── state/                  # Installation state tracking
│   │   ├── state.go            # State struct, Load/Save, Add/Remove for installations, path mods, env vars
│   │   └── undo.go             # UndoInstallation, UndoAll - reverse installs, clean up PATH + env vars
│   │
│   ├── selfupdate/             # Self-update + version checking
│   │   └── update.go           # CheckForUpdate (24h cooldown, cached), DoUpdate (go-selfupdate library)
│   │
│   ├── logger/                 # Logging system
│   │   └── logger.go           # File + stdout, secret masking, log rotation (keeps 10), RecentLogFiles
│   │
│   ├── tui/                    # Terminal UI (Bubbletea)
│   │   ├── app.go              # Main model with phase state machine (summary → confirm → install → packages → configure → complete)
│   │   ├── styles.go           # Lipgloss color palette (purple primary, green/yellow/red status)
│   │   ├── summary.go          # Plan summary table view
│   │   ├── progress.go         # Per-runtime install progress with spinner and download bar
│   │   └── configure.go        # Text input form for .env and config fields (password mode for secrets)
│   │
│   └── server/                 # Local web server for --ui mode
│       ├── server.go           # HTTP server with embedded SPA, auto-port (19532-19631), auto-shutdown
│       ├── api.go              # /api/status endpoint
│       └── ws.go               # WebSocket hub + handler - real-time progress, manifest upload, config save
│
├── web/                        # Embedded React web UI (Vite + React 19 + Tailwind CSS 4 + Shadcn UI)
│   ├── src/
│   │   ├── App.tsx             # Main app with step navigation
│   │   ├── components/
│   │   │   ├── steps/          # 5-step wizard: Welcome, Summary, Install, Configure, Complete
│   │   │   └── ui/             # Shadcn UI components (button, card, input, badge, progress)
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts # WebSocket connection to Go backend
│   │   │   └── useSetupState.ts# Global state management
│   │   └── lib/
│   │       └── utils.ts        # Tailwind cn() helper
│   └── dist/                   # Built assets (embedded into Go binary, gitignored)
│
├── examples/                   # Example .templatr.toml manifests
│   ├── nextjs.templatr.toml    # Next.js SaaS template (node, npm, env vars, site.ts config)
│   ├── django.templatr.toml    # Django CRM template (python + node, pip)
│   ├── flutter.templatr.toml   # Flutter mobile app (flutter + java, pub)
│   └── java-spring.templatr.toml # Java Spring Boot CMS (java, maven)
│
├── docs/
│   ├── RELEASE.md              # Release process, distribution channels, CI/CD
│   └── MANIFEST_SPEC.md        # Full .templatr.toml specification
│
├── .github/
│   ├── workflows/
│   │   ├── ci.yml              # CI: build web UI + Go build/test/vet on Windows/macOS/Linux
│   │   └── release.yml         # Release: GoReleaser on v* tags
│   └── pull_request_template.md
│
└── LICENSE                     # MIT
```

## Architecture

### Launch Detection

The binary auto-detects how it was launched:

- **Terminal with manifest**: Runs the interactive TUI (Bubbletea) in the terminal
- **Terminal without manifest**: Prints usage help (available commands and flags)
- **Terminal with `--ui` flag**: Opens the web dashboard in the default browser
- **Double-click (no terminal)**: Opens the web dashboard in the default browser

Detection logic in `cmd/root.go`: `shouldLaunchWebUI()` returns `true` when `len(os.Args) == 1` AND `launchedFromConsole()` returns `false` (i.e., no terminal detected). The manifest presence is only checked inside cobra's `rootCmd.Run` after the terminal check passes.

On Windows, terminal detection uses two methods in `cmd/console_windows.go`:

1. `AttachConsole(ATTACH_PARENT_PROCESS)` — succeeds when run from cmd/PowerShell (for `-H windowsgui` release builds where the process starts without a console)
2. `GetConsoleProcessList` fallback — if `AttachConsole` fails (console subsystem dev builds already have a console), checks if multiple processes share the console. Count > 1 means a parent shell exists (terminal); count of 1 means Windows created a fresh console (double-click)

On Unix, `launchedFromConsole()` checks if stdin is a terminal via `term.IsTerminal()`.

Release builds use `-H windowsgui` linker flag for Windows to suppress the console window on double-click. `AttachConsole` re-attaches stdout/stderr when run from cmd/PowerShell.

### Runtime Installation

1. Each installer implements the `Installer` interface (see below)
2. `ResolveVersion()` queries the official release API to find the exact version matching the semver constraint
3. `Install()` downloads the binary, verifies SHA256, extracts to `~/.templatr/runtimes/<name>/<version>/`
4. `BinDir()` returns the path to add to PATH
5. `EnvVars()` returns any environment variables to set (e.g., `JAVA_HOME`)
6. Everything is tracked in `~/.templatr/state.json` for clean uninstall

### State Tracking

`~/.templatr/state.json` records:

- Each installation: runtime, version, path, timestamp, previous version (for revert info)
- PATH modifications: method (shell_rc or windows_env), file modified, line added
- Environment variable modifications: name, value, method, file

On uninstall, the tool reads this state to cleanly reverse everything.

### Web UI Communication

The Go backend serves the React SPA and communicates via WebSocket on the same port:

**Server to client messages**: `step`, `runtime`, `download` (progress), `install`, `log`, `complete`, `plan`, `error`

**Client to server messages**: `load_manifest` (with optional file content for drag-and-drop), `confirm`, `configure` (with env/config values), `cancel`

### Logging

- All operations logged to `~/.templatr/logs/setup-{timestamp}.log`
- Log rotation keeps the 10 most recent files
- Secret values (`.env` secrets) are masked in logs as `****`
- Terminal output: INFO to stdout, ERROR to stderr, WARN prefixed

## Adding a New Runtime Installer

1. Create `internal/install/<runtime>.go`
2. Implement the `Installer` interface:

```go
type Installer interface {
    // Name returns the runtime key as used in .templatr.toml (e.g., "node", "python")
    Name() string

    // ResolveVersion takes a semver constraint (e.g., ">=20.0.0", "latest")
    // and returns the exact version to install (e.g., "22.14.0")
    ResolveVersion(requirement string) (string, error)

    // Install downloads and extracts the runtime to targetDir.
    // The progress function is called with (bytesDownloaded, totalBytes) during download.
    Install(version, targetDir string, progress ProgressFunc) error

    // BinDir returns the path to the directory containing executables within installDir.
    // This is added to the user's PATH.
    BinDir(installDir string) string

    // EnvVars returns environment variables to set (e.g., {"JAVA_HOME": installDir}).
    // Return nil or empty map if no env vars are needed.
    EnvVars(installDir string) map[string]string
}
```

3. Register it in `internal/install/installer.go` `init()`:

```go
Register(&MyRuntimeInstaller{})
```

4. Add the runtime key to:
   - `internal/manifest/validate.go` - `validRuntimes` set
   - `internal/detect/runtime.go` - `runtimeChecks` slice (binary name and version flag)
   - `internal/engine/plan.go` - `runtimeDetectNames` and `runtimeDisplayNames` maps

5. Add an example manifest in `examples/` demonstrating the new runtime

6. Write tests - at minimum, test `ResolveVersion` with mock HTTP responses using `httptest`

### Stub Installers

Ruby, PHP, and .NET are currently stub installers that return error messages with manual install instructions. To fully implement them, replace the `ResolveVersion` and `Install` methods with actual download logic following the pattern in `node.go` or `python.go`.

## CLI Commands Reference

| Command                          | Description                                                               |
| -------------------------------- | ------------------------------------------------------------------------- |
| `templatr-setup`                 | TUI if manifest found, help text if no manifest, web UI if double-clicked |
| `templatr-setup setup`           | Install runtimes and packages from .templatr.toml                         |
| `templatr-setup setup --dry-run` | Show what would be installed without changes                              |
| `templatr-setup setup -y`        | Skip confirmation prompt                                                  |
| `templatr-setup configure`       | Run only the configure step (.env + site.ts)                              |
| `templatr-setup doctor`          | Check system: installed runtimes, versions, PATH                          |
| `templatr-setup uninstall`       | Remove runtimes installed by this tool                                    |
| `templatr-setup uninstall --all` | Remove all without prompting                                              |
| `templatr-setup update`          | Self-update to latest version from GitHub Releases                        |
| `templatr-setup logs`            | Show recent log files                                                     |
| `templatr-setup version`         | Show version and check for updates                                        |

## Key Technologies

| Component       | Technology                                                                                                                                                            | Purpose                                               |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| Language        | Go 1.25+                                                                                                                                                              | Cross-compilation, single binary, fast startup        |
| CLI framework   | [Cobra](https://github.com/spf13/cobra)                                                                                                                               | Command/flag parsing (used by Docker CLI, GitHub CLI) |
| TUI framework   | [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles) | Interactive terminal UI with progress bars            |
| TOML parser     | [pelletier/go-toml/v2](https://github.com/pelletier/go-toml)                                                                                                          | Full TOML v1.0 compliance                             |
| Semver          | [Masterminds/semver/v3](https://github.com/Masterminds/semver)                                                                                                        | Version range parsing and comparison                  |
| WebSocket       | [coder/websocket](https://github.com/coder/websocket)                                                                                                                 | Real-time web UI communication                        |
| Self-update     | [go-selfupdate](https://github.com/creativeprojects/go-selfupdate)                                                                                                    | Binary self-update from GitHub Releases               |
| Embedded Web UI | Go `embed` + React 19 + Vite 7 + Tailwind CSS 4 + Shadcn UI                                                                                                           | Single binary with web dashboard                      |
| Build/Release   | [GoReleaser](https://goreleaser.com)                                                                                                                                  | Cross-platform builds + Homebrew/Scoop/winget         |

## Releasing

See [docs/RELEASE.md](docs/RELEASE.md) for the full release process.

Quick summary:

```bash
git tag -a v1.0.0 -m "v1.0.0"
git push origin v1.0.0
```

The GitHub Actions release workflow handles everything: builds the web UI, cross-compiles 6 binaries, creates a GitHub Release, and updates Homebrew/Scoop/winget.

## Conventions

- Use `internal/` for all non-exported packages
- All runtime installers implement the `Installer` interface
- Error messages must be actionable: what failed, why, and how to fix it
- Secret values (`.env` secrets) are NEVER logged - use `logger.AddSecret()`
- Tests use `t.TempDir()` for file operations
- Use `filepath.Join` (not hardcoded slashes) for cross-platform paths
- Use conventional commit prefixes: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- The `--file`/`-f` flag is a persistent root flag shared by all subcommands
- Archives with a single top-level directory are flattened by `ExtractAndFlatten`
