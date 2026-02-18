package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/templatr/templatr-setup/internal/state"
)

// SetEnvVar sets a persistent user-level environment variable (e.g., JAVA_HOME).
// Returns the state entry for tracking.
func SetEnvVar(name, value string) (*state.EnvModification, error) {
	if runtime.GOOS == "windows" {
		return setEnvVarWindows(name, value)
	}
	return setEnvVarUnix(name, value)
}

// RemoveEnvVar removes a persistent user-level environment variable.
func RemoveEnvVar(entry state.EnvModification) error {
	if runtime.GOOS == "windows" {
		return removeEnvVarWindows(entry)
	}
	return removeEnvVarUnix(entry)
}

// AddToPath adds a directory to the user's PATH.
// Returns the state entry for tracking, or nil if no modification was needed.
func AddToPath(binDir string) (*state.PathModification, error) {
	if runtime.GOOS == "windows" {
		return addToPathWindows(binDir)
	}
	return addToPathUnix(binDir)
}

// RemoveFromPath removes a directory from the user's PATH.
func RemoveFromPath(entry state.PathModification) error {
	if runtime.GOOS == "windows" {
		return removeFromPathWindows(entry)
	}
	return removeFromPathUnix(entry)
}

// addToPathWindows adds to the user-level PATH on Windows via PowerShell.
func addToPathWindows(binDir string) (*state.PathModification, error) {
	// Read current user PATH
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`[Environment]::GetEnvironmentVariable("PATH", "User")`)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read user PATH: %w", err)
	}

	currentPath := strings.TrimSpace(string(out))

	// Check if already in PATH
	for _, p := range strings.Split(currentPath, ";") {
		if strings.EqualFold(strings.TrimSpace(p), binDir) {
			return nil, nil // already there
		}
	}

	// Prepend to PATH
	var newPath string
	if currentPath == "" {
		newPath = binDir
	} else {
		newPath = binDir + ";" + currentPath
	}

	cmd = exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("PATH", "%s", "User")`,
			strings.ReplaceAll(newPath, `"`, `\"`)))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to set user PATH: %w", err)
	}

	// Also update current process PATH
	os.Setenv("PATH", binDir+";"+os.Getenv("PATH"))

	return &state.PathModification{
		Method: "windows_env",
		Value:  binDir,
	}, nil
}

// removeFromPathWindows removes a directory from user-level PATH on Windows.
func removeFromPathWindows(entry state.PathModification) error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`[Environment]::GetEnvironmentVariable("PATH", "User")`)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read user PATH: %w", err)
	}

	currentPath := strings.TrimSpace(string(out))
	parts := strings.Split(currentPath, ";")
	var filtered []string
	for _, p := range parts {
		if !strings.EqualFold(strings.TrimSpace(p), entry.Value) {
			filtered = append(filtered, p)
		}
	}

	newPath := strings.Join(filtered, ";")
	cmd = exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("PATH", "%s", "User")`,
			strings.ReplaceAll(newPath, `"`, `\"`)))
	return cmd.Run()
}

// shellConfigFiles returns the shell config files to modify on the current system.
func shellConfigFiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var files []string

	// Check which shells are in use
	shell := os.Getenv("SHELL")

	if strings.Contains(shell, "zsh") || fileExists(filepath.Join(home, ".zshrc")) {
		files = append(files, filepath.Join(home, ".zshrc"))
	}
	if strings.Contains(shell, "bash") || fileExists(filepath.Join(home, ".bashrc")) {
		files = append(files, filepath.Join(home, ".bashrc"))
	}

	// If no shell detected, default to both common configs
	if len(files) == 0 {
		if runtime.GOOS == "darwin" {
			files = append(files, filepath.Join(home, ".zshrc"))
		} else {
			files = append(files, filepath.Join(home, ".bashrc"))
		}
	}

	return files
}

// addToPathUnix appends an export line to shell config files.
func addToPathUnix(binDir string) (*state.PathModification, error) {
	exportLine := fmt.Sprintf(`export PATH="%s:$PATH"`, binDir)
	marker := fmt.Sprintf("# templatr-setup: %s", binDir)
	fullLine := marker + "\n" + exportLine

	files := shellConfigFiles()
	if len(files) == 0 {
		return nil, fmt.Errorf("no shell config files found")
	}

	var modifiedFile string
	for _, rcFile := range files {
		content, err := os.ReadFile(rcFile)
		if err != nil && !os.IsNotExist(err) {
			continue
		}

		// Skip if already added
		if strings.Contains(string(content), marker) {
			continue
		}

		f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			continue
		}

		if _, err := fmt.Fprintf(f, "\n%s\n", fullLine); err != nil {
			f.Close()
			continue
		}
		f.Close()
		modifiedFile = rcFile
	}

	// Also update current process PATH
	os.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	if modifiedFile != "" {
		return &state.PathModification{
			Method: "shell_rc",
			File:   modifiedFile,
			Line:   fullLine,
			Value:  binDir,
		}, nil
	}

	return nil, nil
}

// removeFromPathUnix removes the export line from shell config files.
func removeFromPathUnix(entry state.PathModification) error {
	if entry.File == "" {
		return nil
	}

	content, err := os.ReadFile(entry.File)
	if err != nil {
		return err
	}

	marker := fmt.Sprintf("# templatr-setup: %s", entry.Value)
	lines := strings.Split(string(content), "\n")
	var filtered []string
	skipNext := false

	for _, line := range lines {
		if strings.Contains(line, marker) {
			skipNext = true
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		filtered = append(filtered, line)
	}

	return os.WriteFile(entry.File, []byte(strings.Join(filtered, "\n")), 0o644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// --- Environment variable management ---

func setEnvVarWindows(name, value string) (*state.EnvModification, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("%s", "%s", "User")`,
			name, strings.ReplaceAll(value, `"`, `\""`)))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to set %s: %w", name, err)
	}

	os.Setenv(name, value)

	return &state.EnvModification{
		Name:   name,
		Value:  value,
		Method: "windows_env",
	}, nil
}

func removeEnvVarWindows(entry state.EnvModification) error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("%s", $null, "User")`, entry.Name))
	return cmd.Run()
}

func setEnvVarUnix(name, value string) (*state.EnvModification, error) {
	exportLine := fmt.Sprintf(`export %s="%s"`, name, value)
	marker := fmt.Sprintf("# templatr-setup: %s", name)
	fullLine := marker + "\n" + exportLine

	files := shellConfigFiles()
	if len(files) == 0 {
		return nil, fmt.Errorf("no shell config files found")
	}

	var modifiedFile string
	for _, rcFile := range files {
		content, err := os.ReadFile(rcFile)
		if err != nil && !os.IsNotExist(err) {
			continue
		}

		if strings.Contains(string(content), marker) {
			continue
		}

		f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			continue
		}

		if _, err := fmt.Fprintf(f, "\n%s\n", fullLine); err != nil {
			f.Close()
			continue
		}
		f.Close()
		modifiedFile = rcFile
	}

	os.Setenv(name, value)

	if modifiedFile != "" {
		return &state.EnvModification{
			Name:   name,
			Value:  value,
			Method: "shell_rc",
			File:   modifiedFile,
		}, nil
	}

	return nil, nil
}

func removeEnvVarUnix(entry state.EnvModification) error {
	if entry.File == "" {
		return nil
	}

	content, err := os.ReadFile(entry.File)
	if err != nil {
		return err
	}

	marker := fmt.Sprintf("# templatr-setup: %s", entry.Name)
	lines := strings.Split(string(content), "\n")
	var filtered []string
	skipNext := false

	for _, line := range lines {
		if strings.Contains(line, marker) {
			skipNext = true
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		filtered = append(filtered, line)
	}

	return os.WriteFile(entry.File, []byte(strings.Join(filtered, "\n")), 0o644)
}
