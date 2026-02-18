package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// runtimeStatus tracks the install state of a single runtime.
type runtimeStatus struct {
	name        string
	displayName string
	version     string // resolved version (set after resolve)
	state       installState
	err         error
}

type installState int

const (
	statePending installState = iota
	stateResolving
	stateDownloading
	stateInstalling
	stateDone
	stateFailed
)

// progressModel manages the installation progress display.
type progressModel struct {
	runtimes []runtimeStatus
	current  int
	spinner  spinner.Model
	progress progress.Model
	dlBytes  int64
	dlTotal  int64
	done     bool
	err      error
}

func newProgressModel(runtimeNames []string, displayNames []string) progressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = highlightStyle

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	runtimes := make([]runtimeStatus, len(runtimeNames))
	for i, name := range runtimeNames {
		runtimes[i] = runtimeStatus{
			name:        name,
			displayName: displayNames[i],
			state:       statePending,
		}
	}

	return progressModel{
		runtimes: runtimes,
		spinner:  s,
		progress: p,
	}
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case runtimeResolvingMsg:
		if m.current < len(m.runtimes) {
			m.runtimes[m.current].state = stateResolving
		}
		return m, nil

	case runtimeResolvedMsg:
		if m.current < len(m.runtimes) {
			m.runtimes[m.current].version = msg.version
			m.runtimes[m.current].state = stateDownloading
		}
		return m, nil

	case downloadProgressMsg:
		m.dlBytes = msg.downloaded
		m.dlTotal = msg.total
		if msg.total > 0 {
			pct := float64(msg.downloaded) / float64(msg.total)
			return m, m.progress.SetPercent(pct)
		}
		return m, nil

	case runtimeInstalledMsg:
		if m.current < len(m.runtimes) {
			m.runtimes[m.current].state = stateDone
			m.runtimes[m.current].version = msg.version
		}
		m.current++
		m.dlBytes = 0
		m.dlTotal = 0
		return m, m.progress.SetPercent(0)

	case runtimeFailedMsg:
		if m.current < len(m.runtimes) {
			m.runtimes[m.current].state = stateFailed
			m.runtimes[m.current].err = msg.err
		}
		m.err = msg.err
		m.done = true
		return m, nil

	case installDoneMsg:
		m.done = true
		return m, nil
	}

	return m, nil
}

func (m progressModel) View() string {
	var b strings.Builder

	b.WriteString(boldStyle.Render("Installation Progress"))
	b.WriteString("\n\n")

	for i, rt := range m.runtimes {
		var icon, status string
		switch rt.state {
		case statePending:
			icon = mutedStyle.Render("○")
			status = mutedStyle.Render("pending")
		case stateResolving:
			icon = m.spinner.View()
			status = infoStyle.Render("resolving version...")
		case stateDownloading:
			icon = m.spinner.View()
			if m.dlTotal > 0 {
				status = infoStyle.Render(fmt.Sprintf("downloading %s... %s / %s",
					rt.version, formatBytes(m.dlBytes), formatBytes(m.dlTotal)))
			} else {
				status = infoStyle.Render(fmt.Sprintf("downloading %s...", rt.version))
			}
		case stateInstalling:
			icon = m.spinner.View()
			status = infoStyle.Render(fmt.Sprintf("installing %s...", rt.version))
		case stateDone:
			icon = successStyle.Render(iconCheck)
			status = successStyle.Render(fmt.Sprintf("%s installed", rt.version))
		case stateFailed:
			icon = errorStyle.Render(iconCross)
			status = errorStyle.Render(fmt.Sprintf("failed: %s", rt.err))
		}

		b.WriteString(fmt.Sprintf("  %s %s  %s\n", icon, boldStyle.Render(rt.displayName), status))

		// Show progress bar for the current downloading runtime
		if i == m.current && (rt.state == stateDownloading || rt.state == stateInstalling) && m.dlTotal > 0 {
			b.WriteString(fmt.Sprintf("    %s\n", m.progress.View()))
		}
	}

	return b.String()
}

func formatBytes(b int64) string {
	switch {
	case b >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(b)/(1024*1024*1024))
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
