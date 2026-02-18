package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/templatr/templatr-setup/internal/config"
	"github.com/templatr/templatr-setup/internal/engine"
	"github.com/templatr/templatr-setup/internal/install"
	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/packages"
)

// phase tracks the current TUI state.
type phase int

const (
	phaseSummary   phase = iota // Show plan summary
	phaseConfirm                // Wait for user confirmation
	phaseInstall                // Installing runtimes
	phasePackages               // Installing packages
	phaseConfigure              // Configure .env and config files
	phaseComplete               // Done
)

// Custom messages for async operations.
type (
	runtimeResolvingMsg struct{ name string }
	runtimeResolvedMsg  struct{ name, version string }
	downloadProgressMsg struct{ downloaded, total int64 }
	runtimeInstalledMsg struct {
		name, version, installPath, binDir string
	}
	runtimeFailedMsg struct{ err error }
	installDoneMsg   struct {
		results []install.InstallResult
	}
	packagesDoneMsg struct{ err error }
	configDoneMsg   struct{ err error }
)

// Model is the main Bubbletea model for the setup flow.
type Model struct {
	phase       phase
	plan        *engine.SetupPlan
	log         *logger.Logger
	skipConfirm bool
	width       int
	height      int

	// Sub-models
	progressModel   progressModel
	configureModel  configureModel
	packagesSpinner spinner.Model
	packagesRunning bool

	// Install state
	installResults []install.InstallResult

	// Completion state
	finalErr    error
	logFilePath string
}

// New creates a new TUI model.
func New(plan *engine.SetupPlan, log *logger.Logger, skipConfirm bool) Model {
	// Collect runtime names for progress model
	var names, displayNames []string
	for _, r := range plan.Runtimes {
		if r.Action != engine.ActionSkip {
			names = append(names, r.Name)
			displayNames = append(displayNames, r.DisplayName)
		}
	}

	ps := spinner.New()
	ps.Spinner = spinner.Dot
	ps.Style = highlightStyle

	m := Model{
		plan:            plan,
		log:             log,
		skipConfirm:     skipConfirm,
		progressModel:   newProgressModel(names, displayNames),
		configureModel:  newConfigureModel(plan.Manifest),
		packagesSpinner: ps,
		logFilePath:     log.FilePath(),
	}

	if !plan.NeedsAction() {
		if len(m.configureModel.fields) > 0 {
			m.phase = phaseConfigure
		} else {
			m.phase = phaseComplete
		}
	} else if skipConfirm {
		m.phase = phaseInstall
	} else {
		m.phase = phaseSummary
	}

	return m
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.progressModel.spinner.Tick,
		m.packagesSpinner.Tick,
	}

	if m.phase == phaseInstall {
		cmds = append(cmds, m.installRuntimeCmd(0))
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.phase == phaseComplete || m.phase == phaseSummary || m.phase == phaseConfirm {
				return m, tea.Quit
			}
		}

		switch m.phase {
		case phaseSummary:
			m.phase = phaseConfirm
			return m, nil

		case phaseConfirm:
			switch msg.String() {
			case "y", "Y":
				m.phase = phaseInstall
				return m, m.installRuntimeCmd(0)
			case "n", "N", "esc":
				return m, tea.Quit
			}
			return m, nil

		case phaseConfigure:
			if msg.String() == "esc" {
				m.configureModel.skipped = true
				m.phase = phaseComplete
				return m, nil
			}

			var cmd tea.Cmd
			m.configureModel, cmd = m.configureModel.Update(msg)
			if m.configureModel.done {
				return m, m.writeConfigCmd()
			}
			return m, cmd

		case phaseComplete:
			return m, tea.Quit
		}
	}

	// Handle async messages
	switch msg := msg.(type) {
	case runtimeResolvingMsg, runtimeResolvedMsg, downloadProgressMsg:
		var cmd tea.Cmd
		m.progressModel, cmd = m.progressModel.Update(msg)
		return m, cmd

	case runtimeInstalledMsg:
		m.installResults = append(m.installResults, install.InstallResult{
			Runtime:     msg.name,
			Version:     msg.version,
			InstallPath: msg.installPath,
			BinDir:      msg.binDir,
		})
		var cmd tea.Cmd
		m.progressModel, cmd = m.progressModel.Update(msg)

		// Start next runtime or move to packages
		if m.progressModel.current < len(m.progressModel.runtimes) {
			return m, tea.Batch(cmd, m.installRuntimeCmd(m.progressModel.current))
		}
		m.phase = phasePackages
		m.packagesRunning = true
		return m, tea.Batch(cmd, m.runPackagesCmd())

	case runtimeFailedMsg:
		m.progressModel, _ = m.progressModel.Update(msg)
		m.finalErr = msg.err
		m.phase = phaseComplete
		return m, nil

	case packagesDoneMsg:
		m.packagesRunning = false
		if msg.err != nil {
			m.log.Warn("Package install had issues: %s", msg.err)
		}
		if len(m.configureModel.fields) > 0 {
			m.phase = phaseConfigure
		} else {
			m.phase = phaseComplete
		}
		return m, nil

	case configDoneMsg:
		if msg.err != nil {
			m.log.Warn("Config write failed: %s", msg.err)
		}
		m.phase = phaseComplete
		return m, nil

	case spinner.TickMsg:
		var cmd1, cmd2 tea.Cmd
		m.progressModel.spinner, cmd1 = m.progressModel.spinner.Update(msg)
		m.packagesSpinner, cmd2 = m.packagesSpinner.Update(msg)
		return m, tea.Batch(cmd1, cmd2)
	}

	// Forward progress model updates
	var cmd tea.Cmd
	m.progressModel, cmd = m.progressModel.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	width := m.width
	if width == 0 {
		width = 80
	}

	var b strings.Builder

	switch m.phase {
	case phaseSummary:
		b.WriteString(renderSummary(m.plan, width))
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("Press any key to continue..."))

	case phaseConfirm:
		b.WriteString(renderSummary(m.plan, width))
		b.WriteString("\n")
		b.WriteString(highlightStyle.Render("Proceed with installation? "))
		b.WriteString(boldStyle.Render("[y/n] "))

	case phaseInstall:
		b.WriteString(m.progressModel.View())

	case phasePackages:
		b.WriteString(m.progressModel.View())
		b.WriteString("\n")
		if m.packagesRunning {
			b.WriteString(fmt.Sprintf("  %s Running package install...\n", m.packagesSpinner.View()))
		} else {
			b.WriteString(fmt.Sprintf("  %s Packages installed\n", successStyle.Render(iconCheck)))
		}

	case phaseConfigure:
		b.WriteString(m.configureModel.View())

	case phaseComplete:
		b.WriteString(m.renderComplete())
	}

	return b.String()
}

func (m Model) renderComplete() string {
	var b strings.Builder

	if m.finalErr != nil {
		b.WriteString(errorStyle.Render("Installation failed"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("  %s %s\n", errorStyle.Render(iconCross), m.finalErr))
	} else {
		b.WriteString(successStyle.Render("Setup complete!"))
		b.WriteString("\n\n")

		for _, r := range m.installResults {
			b.WriteString(fmt.Sprintf("  %s %s %s %s %s\n",
				successStyle.Render(iconCheck),
				boldStyle.Render(r.Runtime),
				r.Version,
				mutedStyle.Render(iconArrow),
				mutedStyle.Render(r.InstallPath),
			))
		}

		if m.configureModel.done && !m.configureModel.skipped {
			b.WriteString(fmt.Sprintf("\n  %s Configuration saved\n", successStyle.Render(iconCheck)))
		}
	}

	if m.plan.Manifest.PostSetup.Message != "" && m.finalErr == nil {
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(m.plan.Manifest.PostSetup.Message))
		b.WriteString("\n")
	}

	if m.logFilePath != "" {
		b.WriteString(fmt.Sprintf("\n%s %s\n", mutedStyle.Render("Log file:"), mutedStyle.Render(m.logFilePath)))
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Press q to exit"))

	return b.String()
}

// --- Async commands ---

func (m Model) installRuntimeCmd(idx int) tea.Cmd {
	actionRuntimes := m.actionRuntimes()
	if idx >= len(actionRuntimes) {
		return func() tea.Msg {
			return installDoneMsg{}
		}
	}

	rp := actionRuntimes[idx]
	log := m.log
	slug := m.plan.Manifest.Template.Slug

	return func() tea.Msg {
		result, err := install.InstallSingleRuntime(rp, slug, log, nil)
		if err != nil {
			return runtimeFailedMsg{err: err}
		}

		return runtimeInstalledMsg{
			name:        result.Runtime,
			version:     result.Version,
			installPath: result.InstallPath,
			binDir:      result.BinDir,
		}
	}
}

func (m Model) runPackagesCmd() tea.Cmd {
	mf := m.plan.Manifest
	log := m.log

	return func() tea.Msg {
		if err := packages.RunGlobalInstalls(mf, log); err != nil {
			log.Warn("Global install issues: %s", err)
		}

		var err error
		if mf.Packages.InstallCommand != "" {
			log.Info("Running: %s", mf.Packages.InstallCommand)
			err = packages.RunInstall(mf, log)
		}

		if len(mf.PostSetup.Commands) > 0 {
			log.Info("Running post-setup commands...")
			if postErr := packages.RunPostSetup(mf, log); postErr != nil && err == nil {
				err = postErr
			}
		}

		return packagesDoneMsg{err: err}
	}
}

func (m Model) writeConfigCmd() tea.Cmd {
	mf := m.plan.Manifest
	vals := m.configureModel.Values()
	log := m.log

	return func() tea.Msg {
		// Write .env
		envVals := make(map[string]string)
		for _, env := range mf.Env {
			if v, ok := vals[env.Key]; ok {
				envVals[env.Key] = v
			}
		}
		if len(envVals) > 0 {
			log.Info("Writing .env file...")
			if err := config.WriteEnvFile(".env", mf.Env, envVals); err != nil {
				return configDoneMsg{err: fmt.Errorf("failed to write .env: %w", err)}
			}
		}

		// Write config files
		for _, cfg := range mf.Config {
			fieldVals := make(map[string]string)
			for _, f := range cfg.Fields {
				if v, ok := vals[f.Path]; ok {
					fieldVals[f.Path] = v
				}
			}
			if len(fieldVals) > 0 {
				log.Info("Updating %s...", cfg.File)
				if err := config.UpdateConfigFile(cfg.File, fieldVals); err != nil {
					log.Warn("Failed to update %s: %s", cfg.File, err)
				}
			}
		}

		return configDoneMsg{}
	}
}

func (m Model) actionRuntimes() []engine.RuntimePlan {
	var runtimes []engine.RuntimePlan
	for _, r := range m.plan.Runtimes {
		if r.Action != engine.ActionSkip {
			runtimes = append(runtimes, r)
		}
	}
	return runtimes
}
