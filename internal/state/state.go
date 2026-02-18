package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const stateFile = ".templatr/state.json"

// State tracks all installations performed by templatr-setup.
type State struct {
	Version           string             `json:"version"`
	Installations     []Installation     `json:"installations"`
	PathModifications []PathModification `json:"path_modifications"`
	EnvModifications  []EnvModification  `json:"env_modifications,omitempty"`
}

// Installation records a single runtime installation.
type Installation struct {
	Runtime         string `json:"runtime"`
	Version         string `json:"version"`
	Path            string `json:"path"`
	InstalledAt     string `json:"installed_at"`
	Template        string `json:"template,omitempty"`
	Checksum        string `json:"checksum,omitempty"`
	PreviousVersion string `json:"previous_version,omitempty"` // version before we installed (for revert messaging)
	PreviousPath    string `json:"previous_path,omitempty"`    // path to the previous installation
	Action          string `json:"action"`                     // "install" or "upgrade"
}

// PathModification records a PATH change made by the tool.
type PathModification struct {
	Method  string `json:"method"`            // "shell_rc" or "windows_env"
	File    string `json:"file,omitempty"`     // shell config file path (Unix)
	Line    string `json:"line,omitempty"`     // line added to shell config
	Value   string `json:"value"`             // the PATH directory value
	AddedAt string `json:"added_at"`
}

// EnvModification records an environment variable set by the tool (e.g., JAVA_HOME).
type EnvModification struct {
	Name    string `json:"name"`              // e.g. "JAVA_HOME", "GOROOT"
	Value   string `json:"value"`             // the value set
	Method  string `json:"method"`            // "shell_rc" or "windows_env"
	File    string `json:"file,omitempty"`     // shell config file path (Unix)
	AddedAt string `json:"added_at"`
}

// NewState creates an empty state.
func NewState() *State {
	return &State{
		Version:           "1.0.0",
		Installations:     []Installation{},
		PathModifications: []PathModification{},
	}
}

// stateFilePath returns the full path to the state file.
func stateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, stateFile), nil
}

// Load reads the state file from disk.
func Load() (*State, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewState(), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &s, nil
}

// Save writes the state to disk.
func (s *State) Save() error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// AddInstallation records a new installation.
func (s *State) AddInstallation(inst Installation) {
	inst.InstalledAt = time.Now().UTC().Format(time.RFC3339)
	s.Installations = append(s.Installations, inst)
}

// AddPathModification records a PATH change.
func (s *State) AddPathModification(mod PathModification) {
	mod.AddedAt = time.Now().UTC().Format(time.RFC3339)
	s.PathModifications = append(s.PathModifications, mod)
}

// RemoveInstallation removes an installation by runtime and version.
func (s *State) RemoveInstallation(runtime, version string) {
	var filtered []Installation
	for _, inst := range s.Installations {
		if inst.Runtime == runtime && inst.Version == version {
			continue
		}
		filtered = append(filtered, inst)
	}
	s.Installations = filtered
}

// RemovePathModification removes a PATH modification by value.
func (s *State) RemovePathModification(value string) {
	var filtered []PathModification
	for _, mod := range s.PathModifications {
		if mod.Value == value {
			continue
		}
		filtered = append(filtered, mod)
	}
	s.PathModifications = filtered
}

// AddEnvModification records an environment variable change.
func (s *State) AddEnvModification(mod EnvModification) {
	mod.AddedAt = time.Now().UTC().Format(time.RFC3339)
	s.EnvModifications = append(s.EnvModifications, mod)
}

// RemoveEnvModification removes an env modification by name.
func (s *State) RemoveEnvModification(name string) {
	var filtered []EnvModification
	for _, mod := range s.EnvModifications {
		if mod.Name == name {
			continue
		}
		filtered = append(filtered, mod)
	}
	s.EnvModifications = filtered
}

// GetInstallations returns all installations, optionally filtered by runtime.
func (s *State) GetInstallations(runtime string) []Installation {
	if runtime == "" {
		return s.Installations
	}
	var filtered []Installation
	for _, inst := range s.Installations {
		if inst.Runtime == runtime {
			filtered = append(filtered, inst)
		}
	}
	return filtered
}
