# Contributing

## Project Overview

`templatr-setup` is a cross-platform CLI tool written in Go that reads `.templatr.toml` manifest files from Templatr templates and automatically installs required runtimes, packages, and configuration.

## Key Technologies

- **Language**: Go 1.26+
- **CLI**: Cobra (spf13/cobra)
- **TUI**: Bubbletea + Lipgloss + Bubbles (Charmbracelet suite)
- **TOML**: pelletier/go-toml/v2
- **Semver**: Masterminds/semver/v3
- **WebSocket**: coder/websocket
- **Self-update**: creativeprojects/go-selfupdate
- **Embedded Web UI**: React 19.2, Vite 7, Tailwind CSS 4.2, Shadcn UI
- **Build/Release**: GoReleaser

## Development Commands

```bash
# Build
go build -o templatr-setup .

# Run
go run . doctor
go run . setup --dry-run
go run . --ui

# Test
go test ./...

# Web UI (in web/ directory)
cd web && npm install && npm run dev    # Development
cd web && npm run build                  # Production build

# Tidy dependencies
go mod tidy
```

## Directory Structure

```
cmd/                  # Cobra CLI commands
internal/
├── manifest/         # TOML parser + validation
├── detect/           # OS and runtime detection
├── install/          # Runtime installers (node, python, flutter, etc.)
├── packages/         # Package manager integration (npm, pip, pub, etc.)
├── config/           # Config file writer (.env, TypeScript)
├── state/            # State tracking + undo (reverse installation)
├── logger/           # Structured logging
├── tui/              # Bubbletea terminal UI
└── server/           # Local web server for --ui mode
web/                  # React SPA (embedded into binary)
examples/             # Example .templatr.toml files
docs/                 # Documentation
```

## Architecture

- Single binary with embedded web UI (Go `embed` package)
- Auto-detects terminal vs double-click: TUI for devs, web UI for non-devs
- Downloads runtimes from official sources, verifies SHA256 checksums
- Installs to `~/.templatr/runtimes/` (user-space, no root required)
- Tracks installations in `~/.templatr/state.json` for clean uninstall
- Logs to `~/.templatr/logs/`

## Conventions

- **No `any` types** in the web UI TypeScript code
- Use `internal/` for all non-exported packages
- All runtime installers implement the `Installer` interface
- Error messages must be actionable (what failed, why, how to fix)
- Secret values (.env secrets) are NEVER logged
