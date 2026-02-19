# templatr-setup

A cross-platform setup tool that reads a `.templatr.toml` manifest file bundled with every [Templatr](https://templatr.co) template and automatically installs all required runtimes, packages, and configuration.

- **Non-developers**: Double-click the downloaded file - a visual setup wizard opens in your browser. No terminal knowledge required.
- **Developers**: Run `templatr-setup` in the terminal for an interactive TUI with progress bars and colored output.

## Installation

### macOS / Linux (Homebrew)

```bash
brew install rohan-bhautoo/tap/templatr-setup
```

### Windows (Scoop)

```bash
scoop bucket add templatr https://github.com/rohan-bhautoo/scoop-bucket
scoop install templatr-setup
```

### Windows (winget)

```bash
winget install Templatr.TemplatrSetup
```

### Direct Download

Download the latest release for your platform from the [Releases](https://github.com/rohan-bhautoo/templatr-setup/releases/latest) page.

| Platform | Architecture  | File                                       |
| -------- | ------------- | ------------------------------------------ |
| Windows  | x64           | `templatr-setup_X.Y.Z_windows_amd64.zip`   |
| Windows  | ARM64         | `templatr-setup_X.Y.Z_windows_arm64.zip`   |
| macOS    | Apple Silicon | `templatr-setup_X.Y.Z_darwin_arm64.tar.gz` |
| macOS    | Intel         | `templatr-setup_X.Y.Z_darwin_amd64.tar.gz` |
| Linux    | x64           | `templatr-setup_X.Y.Z_linux_amd64.tar.gz`  |
| Linux    | ARM64         | `templatr-setup_X.Y.Z_linux_arm64.tar.gz`  |

Every release includes a `checksums.txt` file with SHA256 hashes for verification.

> **Note**: If you download directly (not via a package manager), your OS may show a security warning because the binary is not yet code-signed. See the [FAQ](#faq) below for how to bypass this.

## Quick Start

1. Purchase and download a template from [templatr.co](https://templatr.co)
2. Extract the template ZIP into a directory
3. Run the setup tool inside that directory:

```bash
templatr-setup
```

The tool reads the `.templatr.toml` manifest included in the template, scans your system for installed runtimes, and installs only what's missing. After installation, it guides you through configuring `.env` files and site settings.

### Visual Dashboard (no terminal needed)

If you double-click the downloaded binary (or use the `--ui` flag), a local web dashboard opens in your browser at `http://localhost:19532` with a step-by-step wizard:

1. **Welcome** - Detects or lets you upload the `.templatr.toml` manifest
2. **Summary** - Shows what runtimes are needed and what actions will be taken
3. **Install** - Downloads and installs missing runtimes with real-time progress
4. **Configure** - Visual forms for `.env` variables and site configuration files
5. **Complete** - Success summary with next steps

The dashboard communicates with the Go backend over WebSocket for real-time progress updates. When you close the browser tab, the tool shuts down automatically.

## Commands

| Command                          | Description                                                                                 |
| -------------------------------- | ------------------------------------------------------------------------------------------- |
| `templatr-setup`                 | Auto-detect mode: TUI if run from terminal with a manifest present, web dashboard otherwise |
| `templatr-setup --ui`            | Force the web dashboard to open in your browser                                             |
| `templatr-setup setup`           | Run the full setup flow (detect, install, configure)                                        |
| `templatr-setup setup --dry-run` | Preview what would be installed without making changes                                      |
| `templatr-setup setup -y`        | Skip the confirmation prompt and install immediately                                        |
| `templatr-setup setup -f <path>` | Use a specific `.templatr.toml` file instead of auto-detecting                              |
| `templatr-setup configure`       | Run only the configure step (`.env` and site config files)                                  |
| `templatr-setup doctor`          | Show system info and all detected runtimes with versions                                    |
| `templatr-setup uninstall`       | Remove all runtimes installed by this tool                                                  |
| `templatr-setup uninstall --all` | Remove all without prompting for confirmation                                               |
| `templatr-setup update`          | Self-update to the latest version from GitHub Releases                                      |
| `templatr-setup version`         | Show version, build commit, build date, and check for updates                               |
| `templatr-setup logs`            | List the 10 most recent log files                                                           |
| `templatr-setup help`            | Show help text                                                                              |

### Global Flags

| Flag     | Short | Description                                 |
| -------- | ----- | ------------------------------------------- |
| `--ui`   |       | Launch the web dashboard instead of the TUI |
| `--file` | `-f`  | Path to a `.templatr.toml` manifest file    |

### Dry Run Example

```bash
templatr-setup setup --dry-run
```

```
Template: SaaS Landing Page Template (business)

  Runtime     Required    Installed   Action
  ──────────  ──────────  ──────────  ────────
✓ Node.js     >=20.0.0    25.2.1      OK
✗ Python      >=3.12.0    -           Install

Actions needed: 1 to install

Package manager: npm (available)
Install command: npm install

Environment variables: 3 to configure
Config files: 1 to configure
```

### Doctor Example

```bash
templatr-setup doctor
```

```
System Information
  OS:           windows
  Architecture: amd64

Installed Runtimes
  Runtime     Version     Path
  ──────────  ──────────  ──────────────────────────────
✓ Node.js     25.2.1      C:\Program Files\nodejs\node.exe
✓ npm         11.4.2      C:\Program Files\nodejs\npm.cmd
✗ Python      -           -
✓ Go          1.26.0      C:\Program Files\Go\bin\go.exe
✓ Git         2.52.0      C:\Program Files\Git\cmd\git.exe
  ...
```

## How It Works

```
1. PARSE       Read .templatr.toml, validate against schema, check tool version compatibility
2. DETECT      Scan PATH for installed runtimes, get versions (handles Windows Store stubs)
3. COMPARE     Check installed versions against manifest requirements using semver ranges
4. SUMMARIZE   Show exactly what will be installed/upgraded, ask for confirmation
5. INSTALL     Download official binaries, verify SHA256, extract to ~/.templatr/runtimes/
6. PACKAGES    Run package manager install (npm install, pip install, etc.)
7. CONFIGURE   Interactive forms for .env variables and site config files (site.ts etc.)
8. POST-SETUP  Run post-setup commands (npm run build, etc.), show success message
```

All operations are logged to `~/.templatr/logs/` and installations are tracked in `~/.templatr/state.json` for clean uninstall.

### Where Runtimes Are Installed

Runtimes are installed to user-space directories - no root or admin required:

```
~/.templatr/
├── runtimes/
│   ├── node/22.14.0/       # Each runtime gets its own versioned directory
│   ├── python/3.12.8/
│   └── java/21.0.2/
├── state.json               # Tracks what was installed (for uninstall)
├── logs/                    # Log files (keeps last 10, auto-rotated)
│   └── setup-2026-02-19_143000.log
├── last_update_check        # Timestamp for 24h update check cooldown
└── latest_version           # Cached latest version from GitHub
```

The tool prepends the installed runtime's `bin/` directory to your PATH by modifying your shell config file (`~/.bashrc`, `~/.zshrc`) on Unix, or the user PATH environment variable on Windows. Some runtimes also set environment variables (e.g., `JAVA_HOME`, `GOROOT`).

### Uninstall

The `uninstall` command reads `state.json` and cleanly reverses everything:

1. Removes runtime directories from `~/.templatr/runtimes/`
2. Removes PATH entries from shell config files or Windows user PATH
3. Removes environment variables (JAVA_HOME, GOROOT, etc.)
4. Shows revert info if a previous version was detected before the tool ran

It never touches runtimes that were installed by other means.

## Supported Runtimes

| Runtime | Official Source                                                                | Detection Command   | Notes                                                    |
| ------- | ------------------------------------------------------------------------------ | ------------------- | -------------------------------------------------------- |
| Node.js | [nodejs.org](https://nodejs.org/dist/)                                         | `node --version`    | Downloads LTS releases, verified via SHASUMS256          |
| Python  | [python-build-standalone](https://github.com/indygreg/python-build-standalone) | `python3 --version` | Standalone builds, no system Python conflicts            |
| Flutter | [flutter.dev](https://flutter.dev)                                             | `flutter --version` | Stable channel only, SHA256 verified                     |
| Java    | [Adoptium Temurin](https://adoptium.net)                                       | `java --version`    | Sets `JAVA_HOME`, supports major version ranges (`>=21`) |
| Go      | [go.dev](https://go.dev/dl/)                                                   | `go version`        | Sets `GOROOT`, stable releases only                      |
| Rust    | [rustup.rs](https://rustup.rs)                                                 | `rustc --version`   | Installs via rustup-init with custom paths               |
| Ruby    | Manual install instructions provided                                           | `ruby --version`    | Stub - links to official install guides                  |
| PHP     | Manual install instructions provided                                           | `php --version`     | Stub - links to official install guides                  |
| .NET    | Manual install instructions provided                                           | `dotnet --version`  | Stub - links to official install guides                  |

All fully-implemented installers download from official sources and verify SHA256 checksums before installation. Ruby, PHP, and .NET provide installation guidance with links to official sources (these are less commonly needed for Templatr templates).

## The `.templatr.toml` Manifest

Every Templatr template includes a `.templatr.toml` file at the root that describes what the template needs. Here's a complete example:

```toml
# =============================================================================
# .templatr.toml - Template Setup Manifest
# =============================================================================
# Docs: https://templatr.co/tools/setup
# =============================================================================

[template]
name = "SaaS Landing Page Template"
version = "1.0.0"
tier = "business"
category = "website"
slug = "saas-landing-template"

[runtimes]
node = ">=20.0.0"

[packages]
manager = "npm"
install_command = "npm install"
# global = ["typescript", "tsx"]   # Optional: global packages installed first

[[env]]
key = "NEXT_PUBLIC_SITE_URL"
label = "Site URL"
description = "Your production website URL"
default = "http://localhost:3000"
required = true
type = "url"

[[env]]
key = "RESEND_API_KEY"
label = "Resend API Key"
description = "Get your API key at https://resend.com/api-keys"
default = ""
required = false
type = "secret"
docs_url = "https://resend.com/docs"

[[config]]
file = "src/config/site.ts"
label = "Site Configuration"
description = "Core identity and branding for your website"

  [[config.fields]]
  path = "siteConfig.name"
  label = "Site Name"
  type = "text"
  default = "My SaaS"

  [[config.fields]]
  path = "siteConfig.url"
  label = "Production URL"
  type = "url"
  default = "https://your-domain.com"

[post_setup]
commands = ["npm run build"]
message = """
Your template is ready! Run 'npm run dev' to start the development server.
Visit http://localhost:3000 to see your site.
"""

[meta]
min_tool_version = "1.0.0"
docs = "https://templatr.co/saas-landing-template"
```

### Supported Field Types

| Type      | Rendered As             | Validation          |
| --------- | ----------------------- | ------------------- |
| `text`    | Text input              | Required check only |
| `url`     | URL input               | Valid URL format    |
| `email`   | Email input             | Valid email format  |
| `secret`  | Password input (masked) | Never logged        |
| `number`  | Number input            | Numeric validation  |
| `boolean` | Toggle/checkbox         | true/false          |

### Supported Package Managers

`npm`, `pnpm`, `yarn`, `bun`, `pip`, `pub` (Flutter/Dart), `composer` (PHP), `cargo` (Rust), `go`

### Version Ranges

Runtime versions use [semver](https://semver.org/) constraints:

- `">=20.0.0"` - version 20.0.0 or higher
- `"^20.0.0"` - compatible with 20.x.x
- `"~20.0.0"` - patch-level changes only (20.0.x)
- `">=21"` - major version 21 or higher (useful for Java)
- `"latest"` - always satisfied, installs the latest stable version if missing

See the [examples/](examples/) directory for manifests for Next.js, Django, Flutter, and Java Spring templates. For the full specification, see [docs/MANIFEST_SPEC.md](docs/MANIFEST_SPEC.md).

## Auto-Update

The tool checks for newer versions automatically:

- On every run, a background check queries the [latest GitHub release](https://github.com/rohan-bhautoo/templatr-setup/releases/latest) (non-blocking, < 200ms, 24-hour cooldown)
- If a newer version is available, a notice is printed after the main command finishes
- Run `templatr-setup update` to update in-place
- If you installed via Homebrew, Scoop, or winget, the tool detects this and suggests using your package manager instead

## Security

- **Open source** - Inspect every line of code on GitHub before running
- **SHA256 checksums** - Every runtime download is verified against official checksums
- **Official sources only** - Downloads from nodejs.org, flutter.dev, go.dev, adoptium.net, etc.
- **No telemetry** - Zero data leaves your machine. No analytics, no crash reporting
- **User-space installation** - Installs to `~/.templatr/runtimes/`, no root or admin required
- **Reversible** - `templatr-setup uninstall` cleanly removes everything the tool installed
- **Secret masking** - `.env` secret values are never written to log files
- **No arbitrary code execution** - The tool does not run template-provided scripts. Only known package manager commands are executed

## FAQ

### Is this safe to run on my computer?

Yes. The tool is fully open source. It only downloads runtimes from official sources (nodejs.org, flutter.dev, etc.) and every download is verified with SHA256 checksums. No data leaves your machine.

### Windows shows "Windows protected your PC"

This is normal for open-source tools that are not yet code-signed. Click **"More info"** then **"Run anyway"**. This is a one-time action. Installing via **Scoop** or **winget** avoids this warning entirely.

### macOS blocks the app from opening

Run this once in Terminal:

```bash
xattr -d com.apple.quarantine templatr-setup
```

Or install via **Homebrew** to skip this entirely.

### What if I already have Node.js installed?

The tool detects your existing installations. If your version meets the template's requirements, it skips installation entirely. It only installs what's actually missing.

### Can I uninstall everything it installed?

Yes. Run `templatr-setup uninstall` to cleanly remove all runtimes the tool installed. It tracks everything in `~/.templatr/state.json` and never touches software you installed yourself.

### Does it work without internet?

No. The tool needs internet access to download runtimes. However, if everything is already installed, the `configure` command works offline.

### Do I need a Templatr account?

No. The tool works completely standalone. All it needs is the `.templatr.toml` file that comes with your template.

## Building from Source

Requires [Go 1.25+](https://go.dev/dl/) and [Node.js 20+](https://nodejs.org/).

```bash
# Clone the repository
git clone https://github.com/rohan-bhautoo/templatr-setup.git
cd templatr-setup

# Build the embedded web UI
cd web && npm install && npm run build && cd ..

# Build the binary
go build -o templatr-setup .

# Run it
./templatr-setup doctor
```

### Running Tests

```bash
go test ./...
go vet ./...
```

### Snapshot Release (local testing)

```bash
goreleaser release --snapshot --clean
```

This builds all 6 platform binaries in `dist/` without publishing.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture details, and how to add new runtime installers.

## License

[MIT](LICENSE)
