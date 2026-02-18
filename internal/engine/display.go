package engine

import (
	"fmt"
	"strings"
)

// PrintSummary prints a human-readable summary table of the setup plan.
func PrintSummary(plan *SetupPlan) {
	m := plan.Manifest

	fmt.Printf("Template: %s (%s)\n", m.Template.Name, m.Template.Tier)
	if m.Template.Slug != "" {
		fmt.Printf("Docs:     %s\n", m.Meta.Docs)
	}
	fmt.Println()

	if len(plan.Runtimes) == 0 {
		fmt.Println("No runtimes required by this template.")
		return
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

	// Print header
	fmt.Printf("  %-*s  %-*s  %-*s  %s\n", nameW, "Runtime", reqW, "Required", curW, "Installed", "Action")
	fmt.Printf("  %s  %s  %s  %s\n",
		strings.Repeat("─", nameW),
		strings.Repeat("─", reqW),
		strings.Repeat("─", curW),
		strings.Repeat("─", actW),
	)

	// Print rows
	for _, r := range plan.Runtimes {
		cur := r.InstalledVersion
		if cur == "" {
			cur = "—"
		}

		icon := "  "
		switch r.Action {
		case ActionSkip:
			icon = "✓ "
		case ActionInstall:
			icon = "✗ "
		case ActionUpgrade:
			icon = "⬆ "
		}

		fmt.Printf("%s%-*s  %-*s  %-*s  %s\n",
			icon,
			nameW, r.DisplayName,
			reqW, r.RequiredVersion,
			curW, cur,
			r.Action.ActionIcon(),
		)
	}

	fmt.Println()

	// Summary line
	installs := 0
	upgrades := 0
	for _, r := range plan.Runtimes {
		switch r.Action {
		case ActionInstall:
			installs++
		case ActionUpgrade:
			upgrades++
		}
	}

	if installs == 0 && upgrades == 0 {
		fmt.Println("All runtimes are already installed and satisfy the requirements.")
	} else {
		parts := []string{}
		if installs > 0 {
			parts = append(parts, fmt.Sprintf("%d to install", installs))
		}
		if upgrades > 0 {
			parts = append(parts, fmt.Sprintf("%d to upgrade", upgrades))
		}
		fmt.Printf("Actions needed: %s\n", strings.Join(parts, ", "))
	}

	// Package manager info
	if plan.Packages != nil {
		fmt.Println()
		managerStatus := "not found"
		if plan.Packages.ManagerFound {
			managerStatus = "available"
		}
		fmt.Printf("Package manager: %s (%s)\n", plan.Packages.Manager, managerStatus)
		if plan.Packages.InstallCommand != "" {
			fmt.Printf("Install command: %s\n", plan.Packages.InstallCommand)
		}
	}

	// Env vars info
	if len(m.Env) > 0 {
		fmt.Println()
		required := 0
		for _, e := range m.Env {
			if e.Required {
				required++
			}
		}
		fmt.Printf("Environment variables: %d total (%d required)\n", len(m.Env), required)
	}

	// Config files info
	if len(m.Config) > 0 {
		fmt.Println()
		totalFields := 0
		for _, c := range m.Config {
			totalFields += len(c.Fields)
		}
		fmt.Printf("Config files: %d file(s), %d field(s) to configure\n", len(m.Config), totalFields)
	}

	fmt.Println()
}
