package tui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // purple
	colorSuccess   = lipgloss.Color("#22C55E") // green
	colorWarning   = lipgloss.Color("#EAB308") // yellow
	colorDanger    = lipgloss.Color("#EF4444") // red
	colorInfo      = lipgloss.Color("#3B82F6") // blue
	colorMuted     = lipgloss.Color("#6B7280") // gray
	colorHighlight = lipgloss.Color("#A78BFA") // light purple
)

// Text styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	infoStyle = lipgloss.NewStyle().
			Foreground(colorInfo)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	boldStyle = lipgloss.NewStyle().
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(colorHighlight)
)

// Table styles.
var (
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				PaddingRight(2)

	tableCellStyle = lipgloss.NewStyle().
			PaddingRight(2)

	tableBorderStyle = lipgloss.NewStyle().
				Foreground(colorMuted)
)

// Box styles.
var (
	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(1, 2)

	activeBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)
)

// Status icons.
const (
	iconOK      = "✓"
	iconMissing = "✗"
	iconUpgrade = "⬆"
	iconArrow   = "→"
	iconDot     = "●"
	iconCheck   = "✔"
	iconCross   = "✘"
)
