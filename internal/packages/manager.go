package packages

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/manifest"
)

// RunInstall executes the package manager install command from the manifest.
func RunInstall(m *manifest.Manifest, log *logger.Logger) error {
	if m.Packages.InstallCommand == "" {
		log.Info("No install command specified, skipping package installation")
		return nil
	}

	log.Info("Running: %s", m.Packages.InstallCommand)

	parts := strings.Fields(m.Packages.InstallCommand)
	if len(parts) == 0 {
		return fmt.Errorf("empty install command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("package install failed: %w", err)
	}

	return nil
}

// RunGlobalInstalls installs global packages if specified in the manifest.
func RunGlobalInstalls(m *manifest.Manifest, log *logger.Logger) error {
	if len(m.Packages.Global) == 0 {
		return nil
	}

	manager := m.Packages.Manager
	var installCmd string

	switch manager {
	case "npm":
		installCmd = "npm install -g"
	case "pnpm":
		installCmd = "pnpm add -g"
	case "yarn":
		installCmd = "yarn global add"
	case "bun":
		installCmd = "bun add -g"
	case "pip":
		installCmd = "pip install"
	default:
		log.Warn("Global package installation not supported for %s", manager)
		return nil
	}

	for _, pkg := range m.Packages.Global {
		fullCmd := installCmd + " " + pkg
		log.Info("Running: %s", fullCmd)

		parts := strings.Fields(fullCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Warn("Failed to install global package %s: %s", pkg, err)
		}
	}

	return nil
}

// RunPostSetup executes the post_setup commands from the manifest.
func RunPostSetup(m *manifest.Manifest, log *logger.Logger) error {
	if len(m.PostSetup.Commands) == 0 {
		return nil
	}

	for _, cmdStr := range m.PostSetup.Commands {
		log.Info("Running post-setup: %s", cmdStr)

		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("post-setup command %q failed: %w", cmdStr, err)
		}
	}

	return nil
}
