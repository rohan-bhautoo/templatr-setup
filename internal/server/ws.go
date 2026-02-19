package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/templatr/templatr-setup/internal/config"
	"github.com/templatr/templatr-setup/internal/engine"
	"github.com/templatr/templatr-setup/internal/install"
	"github.com/templatr/templatr-setup/internal/manifest"
	"github.com/templatr/templatr-setup/internal/packages"
)

// Message types sent from server to client.
const (
	MsgTypeStep     = "step"
	MsgTypeRuntime  = "runtime"
	MsgTypeDownload = "download"
	MsgTypeInstall  = "install"
	MsgTypeLog      = "log"
	MsgTypeComplete = "complete"
	MsgTypePlan     = "plan"
	MsgTypeError    = "error"
)

// ServerMessage is a message sent from the Go server to the web UI.
type ServerMessage struct {
	Type    string `json:"type"`
	Step    string `json:"step,omitempty"`
	Status  string `json:"status,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Action  string `json:"action,omitempty"`
	// Download progress fields
	Runtime  string  `json:"runtime,omitempty"`
	Progress float64 `json:"progress,omitempty"`
	Speed    string  `json:"speed,omitempty"`
	Total    string  `json:"total,omitempty"`
	// Log fields
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	// Complete fields
	Success bool `json:"success,omitempty"`
	// Plan data (sent once after manifest is loaded)
	Plan *PlanData `json:"plan,omitempty"`
}

// PlanData is the setup plan serialized for the web UI.
type PlanData struct {
	Template TemplateData  `json:"template"`
	Runtimes []RuntimeData `json:"runtimes"`
	Packages *PackageData  `json:"packages,omitempty"`
	EnvVars  []EnvVarData  `json:"envVars,omitempty"`
	Configs  []ConfigData  `json:"configs,omitempty"`
}

// TemplateData is template info for the web UI.
type TemplateData struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Tier     string `json:"tier"`
	Category string `json:"category"`
}

// RuntimeData is runtime plan info for the web UI.
type RuntimeData struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	RequiredVersion  string `json:"requiredVersion"`
	InstalledVersion string `json:"installedVersion"`
	Action           string `json:"action"`
}

// PackageData is package manager info for the web UI.
type PackageData struct {
	Manager        string `json:"manager"`
	InstallCommand string `json:"installCommand"`
	ManagerFound   bool   `json:"managerFound"`
}

// EnvVarData is an env var definition for the web UI form.
type EnvVarData struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	DocsURL     string `json:"docsUrl,omitempty"`
	File        string `json:"file,omitempty"`
}

// ConfigData is a config file definition for the web UI form.
type ConfigData struct {
	File        string          `json:"file"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Fields      []ConfigFieldUI `json:"fields"`
}

// ConfigFieldUI is a single config field for the web UI form.
type ConfigFieldUI struct {
	Path        string `json:"path"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Default     string `json:"default"`
}

// ClientMessage is a message sent from the web UI to the Go server.
type ClientMessage struct {
	Type   string            `json:"type"`
	Action string            `json:"action,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
	Config map[string]string `json:"config,omitempty"`
	// Manifest content for upload
	ManifestContent string `json:"manifestContent,omitempty"`
	ManifestPath    string `json:"manifestPath,omitempty"`
}

// Hub manages WebSocket connections and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan ServerMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
	hadClients bool   // true once at least one client has connected
	onEmpty    func() // called when all clients disconnect after at least one connected
}

// Client represents a single WebSocket connection.
type Client struct {
	conn *websocket.Conn
	send chan ServerMessage
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan ServerMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.hadClients = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			empty := h.hadClients && len(h.clients) == 0
			onEmpty := h.onEmpty
			h.mu.Unlock()

			if empty && onEmpty != nil {
				onEmpty()
			}

		case msg := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg ServerMessage) {
	h.broadcast <- msg
}

// handleWebSocket upgrades the HTTP connection to a WebSocket and processes messages.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // localhost only, so allow all origins
	})
	if err != nil {
		s.log.Error("WebSocket accept error: %s", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan ServerMessage, 256),
	}
	s.hub.register <- client

	// Writer goroutine
	go func() {
		defer conn.CloseNow()
		for msg := range client.send {
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			if err := conn.Write(r.Context(), websocket.MessageText, data); err != nil {
				return
			}
		}
	}()

	// Reader loop - process incoming messages
	defer func() {
		s.hub.unregister <- client
		conn.CloseNow()
	}()

	// If a manifest was specified on the command line, auto-load it
	if s.manifestPath != "" {
		go s.loadManifestAndSendPlan(s.manifestPath)
	}

	for {
		_, data, err := conn.Read(r.Context())
		if err != nil {
			return // connection closed
		}

		var msg ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			s.log.Warn("Invalid WebSocket message: %s", err)
			continue
		}

		s.handleClientMessage(client, msg)
	}
}

// loadManifestAndSendPlan loads a manifest file and broadcasts the plan.
func (s *Server) loadManifestAndSendPlan(path string) {
	m, err := manifest.Load(path)
	if err != nil {
		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeError,
			Message: fmt.Sprintf("Failed to load manifest: %s", err),
		})
		return
	}

	if errs := manifest.Validate(m); len(errs) > 0 {
		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeError,
			Message: fmt.Sprintf("Manifest validation failed: %s", errs[0]),
		})
		return
	}

	s.validateAndBroadcastPlan(m)
}

// loadManifestFromContent parses uploaded TOML content and broadcasts the plan.
func (s *Server) loadManifestFromContent(content string) {
	m, err := manifest.Parse([]byte(content))
	if err != nil {
		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeError,
			Message: fmt.Sprintf("Failed to parse manifest: %s", err),
		})
		return
	}

	s.validateAndBroadcastPlan(m)
}

// validateAndBroadcastPlan validates a manifest, stores it, builds a plan, and broadcasts it.
func (s *Server) validateAndBroadcastPlan(m *manifest.Manifest) {
	if errs := manifest.Validate(m); len(errs) > 0 {
		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeError,
			Message: fmt.Sprintf("Manifest validation failed: %s", errs[0]),
		})
		return
	}

	plan, err := engine.BuildPlan(m)
	if err != nil {
		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeError,
			Message: fmt.Sprintf("Failed to build plan: %s", err),
		})
		return
	}

	s.loadedManifest = m

	s.hub.Broadcast(ServerMessage{
		Type: MsgTypePlan,
		Plan: buildPlanData(plan),
	})
}

// handleClientMessage processes a message from a web UI client.
func (s *Server) handleClientMessage(_ *Client, msg ClientMessage) {
	switch msg.Type {
	case "load_manifest":
		if msg.ManifestContent != "" {
			go s.loadManifestFromContent(msg.ManifestContent)
		} else {
			go s.loadManifestAndSendPlan(msg.ManifestPath)
		}

	case "confirm":
		go s.runInstallation()

	case "configure":
		go s.runConfigure(msg)

	case "cancel":
		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeComplete,
			Success: false,
			Message: "Setup cancelled by user.",
		})
	}
}

// runInstallation performs the full installation flow and broadcasts progress.
func (s *Server) runInstallation() {
	m := s.loadedManifest
	if m == nil {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeError, Message: "No manifest loaded. Please upload a .templatr.toml file first."})
		return
	}

	plan, err := engine.BuildPlan(m)
	if err != nil {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeError, Message: err.Error()})
		return
	}

	s.hub.Broadcast(ServerMessage{Type: MsgTypeStep, Step: "install", Status: "running"})

	// Install runtimes one at a time with progress
	for _, rp := range plan.Runtimes {
		if rp.Action == engine.ActionSkip {
			s.hub.Broadcast(ServerMessage{
				Type:    MsgTypeRuntime,
				Name:    rp.Name,
				Version: rp.InstalledVersion,
				Status:  "installed",
				Action:  "skip",
			})
			continue
		}

		s.hub.Broadcast(ServerMessage{
			Type:   MsgTypeRuntime,
			Name:   rp.Name,
			Status: "installing",
			Action: string(rp.Action),
		})

		progress := func(downloaded, total int64) {
			if total > 0 {
				pct := float64(downloaded) / float64(total) * 100
				s.hub.Broadcast(ServerMessage{
					Type:     MsgTypeDownload,
					Runtime:  rp.Name,
					Progress: pct,
					Total:    formatBytes(total),
				})
			}
		}

		result, err := install.InstallSingleRuntime(rp, m.Template.Slug, s.log, progress)
		if err != nil {
			s.hub.Broadcast(ServerMessage{
				Type:    MsgTypeError,
				Message: fmt.Sprintf("Failed to install %s: %s", rp.DisplayName, err),
			})
			s.hub.Broadcast(ServerMessage{
				Type:    MsgTypeComplete,
				Success: false,
				Message: fmt.Sprintf("Installation failed: %s", err),
			})
			return
		}

		s.hub.Broadcast(ServerMessage{
			Type:    MsgTypeInstall,
			Runtime: rp.Name,
			Version: result.Version,
			Status:  "complete",
		})
	}

	// Run packages
	s.hub.Broadcast(ServerMessage{Type: MsgTypeStep, Step: "packages", Status: "running"})
	s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "info", Message: "Installing packages..."})

	if err := packages.RunGlobalInstalls(m, s.log); err != nil {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "warn", Message: fmt.Sprintf("Global install warning: %s", err)})
	}

	if err := packages.RunInstall(m, s.log); err != nil {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "warn", Message: fmt.Sprintf("Package install warning: %s", err)})
	}

	s.hub.Broadcast(ServerMessage{Type: MsgTypeStep, Step: "packages", Status: "complete"})

	// Check if configure step is needed
	if len(m.Env) > 0 || len(m.Config) > 0 {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeStep, Step: "configure", Status: "ready"})
	} else {
		// Run post-setup and complete
		s.runPostSetupAndComplete(m)
	}
}

// runConfigure writes config values and completes the setup.
func (s *Server) runConfigure(msg ClientMessage) {
	m := s.loadedManifest
	if m == nil {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeError, Message: "No manifest loaded."})
		return
	}

	// Write env files (grouped by target file)
	if len(msg.Env) > 0 && len(m.Env) > 0 {
		// Mask secrets
		for _, envDef := range m.Env {
			if envDef.Type == "secret" {
				if v, ok := msg.Env[envDef.Key]; ok && v != "" {
					s.log.AddSecret(v)
				}
			}
		}

		grouped, fileOrder := config.GroupEnvByFile(m.Env)
		for _, file := range fileOrder {
			s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "info", Message: fmt.Sprintf("Writing %s...", file)})
			defs := grouped[file]
			if err := config.WriteEnvFile(file, defs, msg.Env); err != nil {
				s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "error", Message: fmt.Sprintf("Failed to write %s: %s", file, err)})
			}
		}
	}

	// Write config files
	if len(msg.Config) > 0 {
		for _, cfg := range m.Config {
			s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "info", Message: fmt.Sprintf("Updating %s...", cfg.File)})

			fieldValues := make(map[string]string)
			for _, field := range cfg.Fields {
				if v, ok := msg.Config[field.Path]; ok {
					fieldValues[field.Path] = v
				}
			}

			if len(fieldValues) > 0 {
				if err := config.UpdateConfigFile(cfg.File, fieldValues); err != nil {
					s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "error", Message: fmt.Sprintf("Failed to update %s: %s", cfg.File, err)})
				}
			}
		}
	}

	s.runPostSetupAndComplete(m)
}

// runPostSetupAndComplete runs post-setup commands and sends the completion message.
func (s *Server) runPostSetupAndComplete(m *manifest.Manifest) {
	if len(m.PostSetup.Commands) > 0 {
		s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "info", Message: "Running post-setup commands..."})
		if err := packages.RunPostSetup(m, s.log); err != nil {
			s.hub.Broadcast(ServerMessage{Type: MsgTypeLog, Level: "warn", Message: fmt.Sprintf("Post-setup warning: %s", err)})
		}
	}

	completeMsg := "Setup complete!"
	if m.PostSetup.Message != "" {
		completeMsg = m.PostSetup.Message
	}

	s.hub.Broadcast(ServerMessage{
		Type:    MsgTypeComplete,
		Success: true,
		Message: completeMsg,
	})
}

// buildPlanData converts an engine.SetupPlan to a PlanData for the web UI.
func buildPlanData(plan *engine.SetupPlan) *PlanData {
	pd := &PlanData{
		Template: TemplateData{
			Name:     plan.Manifest.Template.Name,
			Version:  plan.Manifest.Template.Version,
			Tier:     plan.Manifest.Template.Tier,
			Category: plan.Manifest.Template.Category,
		},
	}

	for _, rp := range plan.Runtimes {
		pd.Runtimes = append(pd.Runtimes, RuntimeData{
			Name:             rp.Name,
			DisplayName:      rp.DisplayName,
			RequiredVersion:  rp.RequiredVersion,
			InstalledVersion: rp.InstalledVersion,
			Action:           string(rp.Action),
		})
	}

	if plan.Packages != nil {
		pd.Packages = &PackageData{
			Manager:        plan.Packages.Manager,
			InstallCommand: plan.Packages.InstallCommand,
			ManagerFound:   plan.Packages.ManagerFound,
		}
	}

	for _, env := range plan.Manifest.Env {
		pd.EnvVars = append(pd.EnvVars, EnvVarData{
			Key:         env.Key,
			Label:       env.Label,
			Description: env.Description,
			Default:     env.Default,
			Required:    env.Required,
			Type:        env.Type,
			DocsURL:     env.DocsURL,
			File:        env.File,
		})
	}

	for _, cfg := range plan.Manifest.Config {
		cd := ConfigData{
			File:        cfg.File,
			Label:       cfg.Label,
			Description: cfg.Description,
		}
		for _, field := range cfg.Fields {
			cd.Fields = append(cd.Fields, ConfigFieldUI{
				Path:        field.Path,
				Label:       field.Label,
				Description: field.Description,
				Type:        field.Type,
				Default:     field.Default,
			})
		}
		pd.Configs = append(pd.Configs, cd)
	}

	return pd
}

// formatBytes formats bytes into a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
