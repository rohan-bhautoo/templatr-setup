package manifest

import (
	"fmt"
	"strings"
)

// validRuntimes is the set of runtimes the tool knows how to install.
var validRuntimes = map[string]bool{
	"node":   true,
	"python": true,
	"flutter": true,
	"java":   true,
	"go":     true,
	"rust":   true,
	"ruby":   true,
	"php":    true,
	"dotnet": true,
}

// validManagers is the set of supported package managers.
var validManagers = map[string]bool{
	"npm":      true,
	"pnpm":     true,
	"yarn":     true,
	"bun":      true,
	"pip":      true,
	"pub":      true,
	"composer": true,
	"cargo":    true,
	"go":       true,
}

// validFieldTypes is the set of supported form field types.
var validFieldTypes = map[string]bool{
	"text":    true,
	"url":     true,
	"email":   true,
	"secret":  true,
	"number":  true,
	"boolean": true,
}

// Validate checks the manifest for required fields and valid values.
func Validate(m *Manifest) []error {
	var errs []error

	// Template section
	if m.Template.Name == "" {
		errs = append(errs, fmt.Errorf("[template] name is required"))
	}
	if m.Template.Version == "" {
		errs = append(errs, fmt.Errorf("[template] version is required"))
	}

	// Runtimes
	for name := range m.Runtimes {
		if !validRuntimes[strings.ToLower(name)] {
			errs = append(errs, fmt.Errorf("[runtimes] unknown runtime %q — supported: %s", name, runtimeList()))
		}
	}

	// Packages
	if m.Packages.Manager != "" && !validManagers[m.Packages.Manager] {
		errs = append(errs, fmt.Errorf("[packages] unknown manager %q — supported: %s", m.Packages.Manager, managerList()))
	}

	// Env vars
	for i, env := range m.Env {
		if env.Key == "" {
			errs = append(errs, fmt.Errorf("[env.%d] key is required", i))
		}
		if env.Type != "" && !validFieldTypes[env.Type] {
			errs = append(errs, fmt.Errorf("[env.%d] unknown type %q — supported: %s", i, env.Type, fieldTypeList()))
		}
	}

	// Config files
	for i, cfg := range m.Config {
		if cfg.File == "" {
			errs = append(errs, fmt.Errorf("[config.%d] file path is required", i))
		}
		for j, field := range cfg.Fields {
			if field.Path == "" {
				errs = append(errs, fmt.Errorf("[config.%d.fields.%d] path is required", i, j))
			}
			if field.Type != "" && !validFieldTypes[field.Type] {
				errs = append(errs, fmt.Errorf("[config.%d.fields.%d] unknown type %q", i, j, field.Type))
			}
		}
	}

	return errs
}

func runtimeList() string {
	names := make([]string, 0, len(validRuntimes))
	for k := range validRuntimes {
		names = append(names, k)
	}
	return strings.Join(names, ", ")
}

func managerList() string {
	names := make([]string, 0, len(validManagers))
	for k := range validManagers {
		names = append(names, k)
	}
	return strings.Join(names, ", ")
}

func fieldTypeList() string {
	names := make([]string, 0, len(validFieldTypes))
	for k := range validFieldTypes {
		names = append(names, k)
	}
	return strings.Join(names, ", ")
}
