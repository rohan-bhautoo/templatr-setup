package install

import (
	"fmt"
	"path/filepath"

	"github.com/templatr/templatr-setup/internal/engine"
	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/state"
)

// Installer is the interface that each runtime installer must implement.
type Installer interface {
	// Name returns the runtime key (e.g., "node", "python").
	Name() string

	// ResolveVersion resolves a version requirement (e.g., ">=20.0.0", "latest")
	// to an exact version string (e.g., "22.14.0").
	ResolveVersion(requirement string) (string, error)

	// Install downloads and installs the runtime to targetDir.
	// targetDir is the final location, e.g. ~/.templatr/runtimes/node/22.14.0/
	Install(version, targetDir string, progress ProgressFunc) error

	// BinDir returns the path to the directory containing executables
	// within an installation directory.
	BinDir(installDir string) string

	// EnvVars returns environment variables that should be set for this runtime.
	// For example, Java returns {"JAVA_HOME": installDir}.
	// Return nil if no env vars are needed beyond PATH.
	EnvVars(installDir string) map[string]string
}

// registry holds all registered installers.
var registry = map[string]Installer{}

// Register adds an installer to the registry.
func Register(i Installer) {
	registry[i.Name()] = i
}

// GetInstaller returns the installer for the given runtime name, or nil.
func GetInstaller(name string) Installer {
	return registry[name]
}

func init() {
	Register(&NodeInstaller{})
	Register(&PythonInstaller{})
	Register(&FlutterInstaller{})
	Register(&JavaInstaller{})
	Register(&GoInstaller{})
	Register(&RustInstaller{})
	Register(&RubyInstaller{})
	Register(&PHPInstaller{})
	Register(&DotnetInstaller{})
}

// InstallResult records what was installed for a single runtime.
type InstallResult struct {
	Runtime     string
	Version     string
	InstallPath string
	BinDir      string
}

// ExecutePlan runs the installation plan: resolves versions, downloads,
// installs, updates PATH, and records state.
func ExecutePlan(plan *engine.SetupPlan, log *logger.Logger, progress ProgressFunc) ([]InstallResult, error) {
	runtimesBase, err := RuntimesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine runtimes directory: %w", err)
	}

	st, err := state.Load()
	if err != nil {
		log.Warn("Could not load state file, starting fresh: %s", err)
		st = state.NewState()
	}

	var results []InstallResult

	for _, rp := range plan.Runtimes {
		if rp.Action == engine.ActionSkip {
			continue
		}

		installer := GetInstaller(rp.Name)
		if installer == nil {
			return results, fmt.Errorf("no installer available for runtime %q", rp.Name)
		}

		log.Info("Resolving latest version for %s (requires %s)...", rp.DisplayName, rp.RequiredVersion)

		version, err := installer.ResolveVersion(rp.RequiredVersion)
		if err != nil {
			return results, fmt.Errorf("failed to resolve version for %s: %w", rp.DisplayName, err)
		}
		log.Info("Will install %s %s", rp.DisplayName, version)

		targetDir := filepath.Join(runtimesBase, rp.Name, version)
		log.Info("Installing %s %s to %s...", rp.DisplayName, version, targetDir)

		if err := installer.Install(version, targetDir, progress); err != nil {
			return results, fmt.Errorf("failed to install %s %s: %w", rp.DisplayName, version, err)
		}

		binDir := installer.BinDir(targetDir)
		log.Info("Adding %s to PATH...", binDir)

		pathEntry, err := AddToPath(binDir)
		if err != nil {
			log.Warn("Failed to add %s to PATH: %s", binDir, err)
			log.Warn("You may need to manually add %s to your PATH", binDir)
		} else if pathEntry != nil {
			st.AddPathModification(*pathEntry)
		}

		// Set runtime-specific env vars (e.g., JAVA_HOME, GOROOT)
		envVars := installer.EnvVars(targetDir)
		for envName, envValue := range envVars {
			log.Info("Setting %s=%s", envName, envValue)
			envEntry, err := SetEnvVar(envName, envValue)
			if err != nil {
				log.Warn("Failed to set %s: %s", envName, err)
			} else if envEntry != nil {
				st.AddEnvModification(*envEntry)
			}
		}

		st.AddInstallation(state.Installation{
			Runtime:         rp.Name,
			Version:         version,
			Path:            targetDir,
			Template:        plan.Manifest.Template.Slug,
			Action:          string(rp.Action),
			PreviousVersion: rp.InstalledVersion,
			PreviousPath:    rp.InstalledPath,
		})

		results = append(results, InstallResult{
			Runtime:     rp.Name,
			Version:     version,
			InstallPath: targetDir,
			BinDir:      binDir,
		})

		log.Info("%s %s installed successfully", rp.DisplayName, version)
	}

	if err := st.Save(); err != nil {
		log.Warn("Failed to save state file: %s", err)
	}

	return results, nil
}
