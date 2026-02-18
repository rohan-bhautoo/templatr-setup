package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/templatr/templatr-setup/internal/engine"
)

// renderSummary builds the summary table view for the plan.
func renderSummary(plan *engine.SetupPlan, width int) string {
	var b strings.Builder
	m := plan.Manifest

	// Template header
	b.WriteString(titleStyle.Render("templatr-setup"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Template dependency installer"))
	b.WriteString("\n\n")

	b.WriteString(boldStyle.Render("Template: "))
	b.WriteString(fmt.Sprintf("%s (%s)\n", m.Template.Name, m.Template.Tier))
	if m.Meta.Docs != "" {
		b.WriteString(mutedStyle.Render(fmt.Sprintf("Docs: %s", m.Meta.Docs)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(plan.Runtimes) == 0 {
		b.WriteString(mutedStyle.Render("No runtimes required by this template."))
		return b.String()
	}

	// Calculate column widths
	nameW, reqW, curW, actW := 10, 10, 10, 8
	for _, r := range plan.Runtimes {
		if len(r.DisplayName) > nameW {
			nameW = len(r.DisplayName)
		}
		if len(r.RequiredVersion) > reqW {
			reqW = len(r.RequiredVersion)
		}
		cur := r.InstalledVersion
		if cur == "" {
			cur = "—"
		}
		if len(cur) > curW {
			curW = len(cur)
		}
		act := r.Action.ActionIcon()
		if len(act) > actW {
			actW = len(act)
		}
	}

	// Header row
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		nameW, "Runtime",
		reqW, "Required",
		curW, "Installed",
		"Action",
	)
	b.WriteString(tableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Separator
	sep := fmt.Sprintf("  %s  %s  %s  %s",
		strings.Repeat("─", nameW),
		strings.Repeat("─", reqW),
		strings.Repeat("─", curW),
		strings.Repeat("─", actW),
	)
	b.WriteString(tableBorderStyle.Render(sep))
	b.WriteString("\n")

	// Rows
	for _, r := range plan.Runtimes {
		cur := r.InstalledVersion
		if cur == "" {
			cur = "—"
		}

		var icon string
		var actionStyled string
		switch r.Action {
		case engine.ActionSkip:
			icon = successStyle.Render(iconOK)
			actionStyled = successStyle.Render("OK")
		case engine.ActionInstall:
			icon = errorStyle.Render(iconMissing)
			actionStyled = warningStyle.Render("Install")
		case engine.ActionUpgrade:
			icon = warningStyle.Render(iconUpgrade)
			actionStyled = warningStyle.Render("Upgrade")
		}

		row := fmt.Sprintf("%s %-*s  %-*s  %-*s  %s",
			icon,
			nameW, r.DisplayName,
			reqW, r.RequiredVersion,
			curW, cur,
			actionStyled,
		)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Actions summary
	installs, upgrades := 0, 0
	for _, r := range plan.Runtimes {
		switch r.Action {
		case engine.ActionInstall:
			installs++
		case engine.ActionUpgrade:
			upgrades++
		}
	}

	if installs == 0 && upgrades == 0 {
		b.WriteString(successStyle.Render("All runtimes are already installed and satisfy the requirements."))
	} else {
		var parts []string
		if installs > 0 {
			parts = append(parts, fmt.Sprintf("%d to install", installs))
		}
		if upgrades > 0 {
			parts = append(parts, fmt.Sprintf("%d to upgrade", upgrades))
		}
		b.WriteString(boldStyle.Render("Actions needed: "))
		b.WriteString(warningStyle.Render(strings.Join(parts, ", ")))
	}
	b.WriteString("\n")

	// Package manager
	if plan.Packages != nil {
		b.WriteString("\n")
		status := errorStyle.Render("not found")
		if plan.Packages.ManagerFound {
			status = successStyle.Render("available")
		}
		b.WriteString(fmt.Sprintf("Package manager: %s (%s)\n",
			boldStyle.Render(plan.Packages.Manager), status))
	}

	// Env vars info
	if len(m.Env) > 0 {
		required := 0
		for _, e := range m.Env {
			if e.Required {
				required++
			}
		}
		b.WriteString(fmt.Sprintf("Environment variables: %d total (%d required)\n", len(m.Env), required))
	}

	// Config files
	if len(m.Config) > 0 {
		totalFields := 0
		for _, c := range m.Config {
			totalFields += len(c.Fields)
		}
		b.WriteString(fmt.Sprintf("Config files: %d file(s), %d field(s)\n", len(m.Config), totalFields))
	}

	return lipgloss.NewStyle().MaxWidth(width).Render(b.String())
}
