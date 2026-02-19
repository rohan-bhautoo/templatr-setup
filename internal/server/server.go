package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/manifest"
)

const defaultPort = 19532

// Server is the local HTTP server that serves the embedded web UI
// and provides WebSocket/REST APIs for the setup wizard.
type Server struct {
	assets   embed.FS
	log      *logger.Logger
	hub            *Hub
	port           int
	srv            *http.Server
	manifestPath   string // path to manifest file (from --file flag)
	loadedManifest *manifest.Manifest // parsed manifest (from file or upload)
}

// New creates a new server with the embedded web assets.
func New(assets embed.FS, log *logger.Logger, manifestFile string) *Server {
	return &Server{
		assets:       assets,
		log:          log,
		hub:          NewHub(),
		port:         defaultPort,
		manifestPath: manifestFile,
	}
}

// Start starts the HTTP server and opens the browser.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/status", s.handleStatus)

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Serve embedded SPA assets
	mux.Handle("/", s.spaHandler())

	// Find available port
	port, err := s.findPort()
	if err != nil {
		return fmt.Errorf("could not find available port: %w", err)
	}
	s.port = port

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://%s", addr)

	s.log.Info("Starting web dashboard on %s", url)

	// Shut down when all browser tabs disconnect or on OS signal
	s.hub.onEmpty = func() {
		s.log.Info("All clients disconnected, shutting down")
		go s.Shutdown()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		s.Shutdown()
	}()

	// Open browser after a short delay to let server start
	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser(url)
	}()

	// Start the hub for WebSocket connections
	go s.hub.Run()

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// spaHandler returns an HTTP handler that serves the embedded SPA.
// It serves files from web/dist/ and falls back to index.html for
// client-side routing.
func (s *Server) spaHandler() http.Handler {
	// Try to get the web/dist subdirectory from the embedded FS
	distFS, err := fs.Sub(s.assets, "web/dist")
	if err != nil {
		// No web/dist directory — return a fallback page
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, fallbackHTML)
		})
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if the file exists
		f, err := distFS.Open(path[1:]) // strip leading /
		if err != nil {
			// File not found — serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}

// handleStatus returns a simple health check response.
func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","manifest":"%s"}`, s.manifestPath)
}

// findPort tries the default port, then scans upward for an available one.
func (s *Server) findPort() (int, error) {
	for port := defaultPort; port < defaultPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range %d-%d", defaultPort, defaultPort+100)
}

// openBrowser opens the given URL in the user's default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux and others
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

const fallbackHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>templatr-setup</title>
  <style>
    body { font-family: system-ui, sans-serif; background: #0a0a0b; color: #e4e4e7; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
    .container { text-align: center; max-width: 480px; padding: 2rem; }
    h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
    p { color: #a1a1aa; line-height: 1.6; }
    code { background: #27272a; padding: 0.2em 0.5em; border-radius: 4px; font-size: 0.9em; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Web UI Not Built</h1>
    <p>The web assets have not been compiled yet. Run the following to build:</p>
    <p><code>cd web && npm install && npm run build</code></p>
    <p>Then restart the tool with <code>templatr-setup --ui</code></p>
  </div>
</body>
</html>`
