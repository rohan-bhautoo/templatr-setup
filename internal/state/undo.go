package state

import (
	"fmt"
	"os"
	"strings"
)

// UndoResult holds the cleanup info from undoing an installation.
type UndoResult struct {
	PathMod  *PathModification
	EnvMods  []EnvModification
	Previous *Installation // what was there before (for messaging)
}

// UndoInstallation removes an installed runtime from disk and cleans up state.
func (s *State) UndoInstallation(runtime, version string) (*UndoResult, error) {
	// Find the installation
	var target *Installation
	for i := range s.Installations {
		if s.Installations[i].Runtime == runtime && s.Installations[i].Version == version {
			target = &s.Installations[i]
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("installation not found: %s %s", runtime, version)
	}

	result := &UndoResult{}

	// Record previous version info for messaging
	if target.PreviousVersion != "" {
		result.Previous = &Installation{
			Runtime: target.Runtime,
			Version: target.PreviousVersion,
			Path:    target.PreviousPath,
		}
	}

	// Remove the runtime directory
	if target.Path != "" {
		if err := os.RemoveAll(target.Path); err != nil {
			return nil, fmt.Errorf("failed to remove %s: %w", target.Path, err)
		}
	}

	// Find associated PATH modification
	for _, mod := range s.PathModifications {
		if target.Path != "" && strings.HasPrefix(mod.Value, target.Path) {
			result.PathMod = &mod
			break
		}
	}

	// Find associated env modifications (e.g., JAVA_HOME pointing into our install dir)
	for _, mod := range s.EnvModifications {
		if target.Path != "" && strings.HasPrefix(mod.Value, target.Path) {
			result.EnvMods = append(result.EnvMods, mod)
		}
	}

	// Clean up state
	s.RemoveInstallation(runtime, version)
	if result.PathMod != nil {
		s.RemovePathModification(result.PathMod.Value)
	}
	for _, envMod := range result.EnvMods {
		s.RemoveEnvModification(envMod.Name)
	}

	return result, nil
}

// UndoAll removes all installations tracked in state.
func (s *State) UndoAll() ([]UndoResult, []error) {
	var results []UndoResult
	var errs []error

	// Work on a copy since we're modifying the slice
	installations := make([]Installation, len(s.Installations))
	copy(installations, s.Installations)

	for _, inst := range installations {
		result, err := s.UndoInstallation(inst.Runtime, inst.Version)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s %s: %w", inst.Runtime, inst.Version, err))
			continue
		}
		results = append(results, *result)
	}

	return results, errs
}
