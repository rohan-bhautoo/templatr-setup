# templatr-setup

A cross-platform CLI tool that reads `.templatr.toml` manifest files from [Templatr](https://templatr.co) templates and automatically installs all required runtimes, packages, and configuration.

**For developers:** Run in your terminal for an interactive TUI experience.
**For everyone else:** Double-click the downloaded file - a visual setup wizard opens in your browser. No terminal knowledge required.

## Installation

### macOS / Linux

```bash
brew install rohan-bhautoo/tap/templatr-setup
```

### Windows

```bash
scoop bucket add templatr https://github.com/rohan-bhautoo/scoop-bucket
scoop install templatr-setup
```

Or via winget:

```bash
winget install Templatr.TemplatrSetup
```

### Direct Download

Download the latest release for your OS from the [Releases](https://github.com/rohan-bhautoo/templatr-setup/releases/latest) page.

## Quick Start

1. Download a template from [templatr.co](https://templatr.co)
2. Navigate to the template directory
3. Run the setup tool:

```bash
templatr-setup
```

The tool reads the `.templatr.toml` file included in your template, detects what's already installed on your system, and installs only what's missing.

## Usage

```bash
templatr-setup                     # Auto-detect: TUI or Web UI
templatr-setup --ui                # Force web dashboard in browser
templatr-setup setup --dry-run     # Preview what would be installed
templatr-setup doctor              # Check system status and installed runtimes
templatr-setup configure           # Set up .env and site configuration
templatr-setup uninstall           # Remove runtimes installed by this tool
templatr-setup update              # Self-update to the latest version
templatr-setup logs                # View recent log files
```

### Dry Run

Preview what the tool will install without making any changes:

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
```

## Supported Runtimes

| Runtime | Source                  |
| ------- | ----------------------- |
| Node.js | nodejs.org              |
| Python  | python-build-standalone |
| Flutter | flutter.dev             |
| Java    | Adoptium (Temurin)      |
| Go      | go.dev                  |
| Rust    | rustup.rs               |
| Ruby    | ruby-builder            |
| PHP     | php.net                 |
| .NET    | dot.net                 |

All runtimes are downloaded from official sources and verified with SHA256 checksums.

## The `.templatr.toml` Manifest

Every Templatr template includes a `.templatr.toml` file at the root that describes its requirements:

```toml
[template]
name = "SaaS Landing Page Template"
version = "1.0.0"
tier = "business"
category = "website"

[runtimes]
node = ">=20.0.0"

[packages]
manager = "npm"
install_command = "npm install"

[[env]]
key = "NEXT_PUBLIC_SITE_URL"
label = "Site URL"
default = "http://localhost:3000"
required = true
type = "url"

[post_setup]
commands = ["npm run build"]
message = "Your template is ready! Run 'npm run dev' to start."
```

See the [examples/](examples/) directory for more manifest examples.

## How It Works

1. **Parse** - Reads `.templatr.toml` and validates the manifest
2. **Detect** - Scans your system for installed runtimes and their versions
3. **Compare** - Checks installed versions against the manifest requirements using semver
4. **Summarize** - Shows you exactly what will be installed before doing anything
5. **Install** - Downloads and installs only what's missing to `~/.templatr/runtimes/`
6. **Configure** - Helps you fill out `.env` and site configuration through interactive forms
7. **Done** - Runs post-setup commands and you're ready to go

## Security

- **Open source** - Inspect every line of code before running
- **SHA256 checksums** - Every download is verified against official checksums
- **Official sources only** - Runtimes are downloaded from nodejs.org, flutter.dev, go.dev, etc.
- **No telemetry** - Zero data leaves your machine
- **User-space installation** - Installs to `~/.templatr/runtimes/`, no root/admin required
- **Reversible** - `templatr-setup uninstall` cleanly removes everything the tool installed

## Contributing

Contributions are welcome! Please open an issue first to discuss what you'd like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## Building from Source

Requires [Go 1.25+](https://go.dev/dl/) and [Node.js 20+](https://nodejs.org/).

```bash
git clone https://github.com/rohan-bhautoo/templatr-setup.git
cd templatr-setup

# Build the web UI
cd web && npm install && npm run build && cd ..

# Build the binary
go build -o templatr-setup .
```

### Running Tests

```bash
go test ./...
```

## License

[MIT](LICENSE)
