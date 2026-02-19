# templatr-setup Web UI

The embedded web dashboard for `templatr-setup`. This is a React single-page application built with Vite that gets compiled to static assets and embedded into the Go binary at build time via `//go:embed all:web/dist`.

## Stack

| Technology   | Version | Purpose                                         |
| ------------ | ------- | ----------------------------------------------- |
| React        | 19.2    | UI framework                                    |
| TypeScript   | 5.9     | Type safety                                     |
| Vite         | 7.3     | Build tool and dev server                       |
| Tailwind CSS | 4.2     | Styling (via `@tailwindcss/vite` plugin)        |
| Shadcn UI    | 3.8     | Pre-built accessible components                 |
| Tabler Icons | 3.36    | Icon library (consistent with Templatr website) |
| Radix UI     | 1.4     | Headless UI primitives (via Shadcn)             |

## Development

```bash
# Install dependencies
npm install

# Start dev server with hot reload (http://localhost:5173)
npm run dev

# Build for production (outputs to dist/)
npm run build

# Preview production build
npm run preview

# Lint
npm run lint
```

During development, run `npm run dev` in this directory and the Go backend separately (`go run . --ui`). The web UI dev server runs on port 5173 and you'll need to configure the WebSocket URL to point to the Go backend port (19532).

For production, the built `dist/` directory is embedded into the Go binary. The Go HTTP server serves these static files and handles the WebSocket connection on the same port.

## Architecture

### How It Works

1. The Go binary starts an HTTP server on `localhost:19532` (auto-finds an available port in range 19532-19631)
2. The SPA is served from the embedded filesystem (`web/dist/`)
3. The React app connects to the Go backend via WebSocket at `ws://localhost:{port}/ws`
4. All setup operations (manifest parsing, runtime detection, installation, configuration) happen in Go
5. The web UI is purely a display/input layer - it sends user actions and receives progress events

### WebSocket Protocol

The web UI communicates with the Go backend using JSON messages over WebSocket.

**Server to client (progress events):**

```typescript
// Step lifecycle
{ type: "step", step: "install" | "packages" | "configure", status: "running" | "complete" | "ready" }

// Plan loaded (after manifest is parsed)
{ type: "plan", plan: { template: {...}, runtimes: [...], packages: {...}, envVars: [...], configs: [...] } }

// Per-runtime events
{ type: "runtime", name: "node", version: "22.14.0", status: "installed" | "installing", action: "install" | "skip" }

// Download progress (0-100)
{ type: "download", runtime: "python", progress: 45.2, total: "28.3 MB" }

// Runtime installation complete
{ type: "install", runtime: "python", version: "3.12.8", status: "complete" }

// Log messages
{ type: "log", level: "info" | "warn" | "error", message: "Running npm install..." }

// Setup complete
{ type: "complete", success: true, message: "Your template is ready!" }

// Error
{ type: "error", message: "Failed to download Node.js" }
```

**Client to server (user actions):**

```typescript
// Load a manifest (either by path or by uploading content)
{ type: "load_manifest", manifestPath: "/path/to/.templatr.toml" }
{ type: "load_manifest", manifestContent: "..." }  // for drag-and-drop upload

// Confirm installation plan
{ type: "confirm" }

// Submit configuration values
{ type: "configure", env: { "SITE_URL": "https://example.com" }, config: { "siteConfig.name": "My SaaS" } }

// Cancel the operation
{ type: "cancel" }
```

## Components

### Step Components (`src/components/steps/`)

The UI is a 5-step wizard:

| Component           | Step | Description                                                 |
| ------------------- | ---- | ----------------------------------------------------------- |
| `WelcomeStep.tsx`   | 1    | Auto-detects manifest or lets user upload via drag-and-drop |
| `SummaryStep.tsx`   | 2    | Shows runtime table, package manager info, download sizes   |
| `InstallStep.tsx`   | 3    | Real-time progress bars for each runtime download/install   |
| `ConfigureStep.tsx` | 4    | Form inputs for `.env` variables and site config fields     |
| `CompleteStep.tsx`  | 5    | Success message, next steps, log file location              |

### UI Components (`src/components/ui/`)

Shadcn UI components installed via `npx shadcn@latest add`:

- `button` - Primary actions, download links
- `card` - Container for summary sections
- `input` - Text fields in configure step
- `badge` - Status indicators (OK, Install, Upgrade)
- `progress` - Download progress bar

### Hooks (`src/hooks/`)

| Hook               | Purpose                                                        |
| ------------------ | -------------------------------------------------------------- |
| `useWebSocket.ts`  | Manages WebSocket connection, reconnection, message parsing    |
| `useSetupState.ts` | Global state: current step, plan data, progress, config values |

### Lib (`src/lib/`)

| File       | Purpose                                  |
| ---------- | ---------------------------------------- |
| `utils.ts` | `cn()` helper for Tailwind class merging |

## Build Output

`npm run build` produces static files in `dist/`:

```
dist/
├── index.html
├── assets/
│   ├── index-{hash}.js      # Bundled React app
│   └── index-{hash}.css     # Bundled Tailwind CSS
└── vite.svg
```

These are embedded into the Go binary by the `//go:embed all:web/dist` directive in `embed.go` at the project root. When `dist/` doesn't exist (development without building the web UI), the Go server returns a fallback HTML page with instructions to build the web UI.

## Styling

- Dark mode by default (Shadcn UI's `dark` class on `<html>`)
- Tailwind CSS v4 with the `@tailwindcss/vite` plugin (no `tailwind.config.ts` needed)
- `tw-animate-css` for entrance animations
- All colors follow the Shadcn UI CSS variable system
