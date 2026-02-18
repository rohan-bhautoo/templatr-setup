package engine

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/templatr/templatr-setup/internal/detect"
	"github.com/templatr/templatr-setup/internal/manifest"
)

// ActionType describes what needs to happen for a runtime.
type ActionType string

const (
	ActionSkip    ActionType = "skip"    // Already installed and satisfies requirement
	ActionInstall ActionType = "install" // Not installed at all
	ActionUpgrade ActionType = "upgrade" // Installed but doesn't satisfy version requirement
)

// RuntimePlan describes the action needed for a single runtime.
type RuntimePlan struct {
	Name             string     // e.g. "node", "python"
	DisplayName      string     // e.g. "Node.js", "Python"
	RequiredVersion  string     // from manifest, e.g. ">=20.0.0" or "latest"
	InstalledVersion string     // from detection, e.g. "25.2.1" or ""
	Action           ActionType // skip, install, upgrade
	InstalledPath    string     // path to existing binary, if any
}

// SetupPlan contains the full plan for a setup operation.
type SetupPlan struct {
	Manifest *manifest.Manifest
	Runtimes []RuntimePlan
	Packages *PackagePlan
}

// PackagePlan describes the package installation step.
type PackagePlan struct {
	Manager        string
	InstallCommand string
	ManagerFound   bool
}

// runtimeDisplayNames maps manifest runtime keys to human-readable names.
var runtimeDisplayNames = map[string]string{
	"node":    "Node.js",
	"python":  "Python",
	"flutter": "Flutter",
	"java":    "Java (Temurin)",
	"go":      "Go",
	"rust":    "Rust",
	"ruby":    "Ruby",
	"php":     "PHP",
	"dotnet":  ".NET",
}

// runtimeDetectNames maps manifest runtime keys to detection names used by detect.ScanRuntimes.
var runtimeDetectNames = map[string]string{
	"node":    "Node.js",
	"python":  "Python",
	"flutter": "Flutter",
	"java":    "Java",
	"go":      "Go",
	"rust":    "Rust",
	"ruby":    "Ruby",
	"php":     "PHP",
	"dotnet":  ".NET",
}

// managerDetectNames maps package managers to detection names.
var managerDetectNames = map[string]string{
	"npm":  "npm",
	"pnpm": "pnpm",
	"yarn": "yarn",
	"bun":  "bun",
	"pip":  "pip",
	"pub":  "Dart", // pub comes with dart
}

// BuildPlan creates a setup plan by comparing manifest requirements against detected runtimes.
func BuildPlan(m *manifest.Manifest) (*SetupPlan, error) {
	// Detect what's installed on the system
	detected := detect.ScanRuntimes()
	detectedMap := make(map[string]detect.RuntimeInfo, len(detected))
	for _, r := range detected {
		detectedMap[r.Name] = r
	}

	plan := &SetupPlan{
		Manifest: m,
		Runtimes: make([]RuntimePlan, 0, len(m.Runtimes)),
	}

	// Compare each required runtime against what's installed
	for name, required := range m.Runtimes {
		detectName, ok := runtimeDetectNames[name]
		if !ok {
			detectName = name
		}
		displayName, ok := runtimeDisplayNames[name]
		if !ok {
			displayName = name
		}

		rp := RuntimePlan{
			Name:            name,
			DisplayName:     displayName,
			RequiredVersion: required,
		}

		info, found := detectedMap[detectName]
		if found && info.Installed {
			rp.InstalledVersion = info.Version
			rp.InstalledPath = info.Path

			// Check if installed version satisfies the requirement
			satisfied, err := versionSatisfies(info.Version, required)
			if err != nil {
				// If we can't parse versions, flag as needing action
				rp.Action = ActionUpgrade
			} else if satisfied {
				rp.Action = ActionSkip
			} else {
				rp.Action = ActionUpgrade
			}
		} else {
			rp.Action = ActionInstall
		}

		plan.Runtimes = append(plan.Runtimes, rp)
	}

	// Check package manager availability
	if m.Packages.Manager != "" {
		pp := &PackagePlan{
			Manager:        m.Packages.Manager,
			InstallCommand: m.Packages.InstallCommand,
		}
		managerDetect, ok := managerDetectNames[m.Packages.Manager]
		if ok {
			info, found := detectedMap[managerDetect]
			pp.ManagerFound = found && info.Installed
		}
		plan.Packages = pp
	}

	return plan, nil
}

// versionSatisfies checks if an installed version satisfies a requirement string.
// Requirement can be: "latest", ">=20.0.0", "^20.0.0", "~20.0.0", "20.0.0", etc.
func versionSatisfies(installed, required string) (bool, error) {
	if required == "latest" {
		// "latest" means any installed version is acceptable for the plan display,
		// but we'll resolve and potentially upgrade in the install step.
		return true, nil
	}

	// Clean installed version string
	installed = strings.TrimSpace(installed)
	if strings.Contains(installed, "(version unknown)") {
		return false, fmt.Errorf("version unknown")
	}

	v, err := semver.NewVersion(installed)
	if err != nil {
		return false, fmt.Errorf("cannot parse installed version %q: %w", installed, err)
	}

	c, err := semver.NewConstraint(required)
	if err != nil {
		return false, fmt.Errorf("cannot parse version requirement %q: %w", required, err)
	}

	return c.Check(v), nil
}

// NeedsAction returns true if the plan has any runtimes that need installation or upgrade.
func (p *SetupPlan) NeedsAction() bool {
	for _, r := range p.Runtimes {
		if r.Action != ActionSkip {
			return true
		}
	}
	return false
}

// ActionIcon returns a display icon for the action type.
func (a ActionType) ActionIcon() string {
	switch a {
	case ActionSkip:
		return "OK"
	case ActionInstall:
		return "Install"
	case ActionUpgrade:
		return "Upgrade"
	default:
		return "?"
	}
}

// StatusIcon returns a colored status indicator.
func (a ActionType) StatusIcon() string {
	switch a {
	case ActionSkip:
		return "[OK]"
	case ActionInstall:
		return "[MISSING]"
	case ActionUpgrade:
		return "[UPGRADE]"
	default:
		return "[?]"
	}
}
