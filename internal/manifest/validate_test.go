package manifest

import (
	"testing"
)

func TestValidate_ValidManifest(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{
			Name:    "Test",
			Version: "1.0.0",
		},
		Runtimes: map[string]string{
			"node": ">=20.0.0",
		},
		Packages: PackageConfig{
			Manager: "npm",
		},
		Env: []EnvVar{
			{Key: "API_KEY", Type: "secret"},
		},
	}

	errs := Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0: %v", len(errs), errs)
	}
}

func TestValidate_MissingTemplateName(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{
			Version: "1.0.0",
		},
	}

	errs := Validate(m)
	found := false
	for _, e := range errs {
		if e.Error() == "[template] name is required" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should report missing template name")
	}
}

func TestValidate_MissingTemplateVersion(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{
			Name: "Test",
		},
	}

	errs := Validate(m)
	found := false
	for _, e := range errs {
		if e.Error() == "[template] version is required" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should report missing template version")
	}
}

func TestValidate_UnknownRuntime(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Runtimes: map[string]string{
			"unknown_runtime": "1.0.0",
		},
	}

	errs := Validate(m)
	if len(errs) == 0 {
		t.Error("Validate() should report unknown runtime")
	}
}

func TestValidate_UnknownPackageManager(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Packages: PackageConfig{
			Manager: "unknown_manager",
		},
	}

	errs := Validate(m)
	if len(errs) == 0 {
		t.Error("Validate() should report unknown package manager")
	}
}

func TestValidate_UnknownFieldType(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Env: []EnvVar{
			{Key: "TEST", Type: "invalid_type"},
		},
	}

	errs := Validate(m)
	if len(errs) == 0 {
		t.Error("Validate() should report unknown field type")
	}
}

func TestValidate_EmptyEnvKey(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Env: []EnvVar{
			{Key: "", Type: "text"},
		},
	}

	errs := Validate(m)
	found := false
	for _, e := range errs {
		if e.Error() == "[env.0] key is required" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should report missing env key")
	}
}

func TestValidate_EmptyConfigFilePath(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Config: []ConfigFile{
			{File: "", Fields: []ConfigField{{Path: "a.b", Type: "text"}}},
		},
	}

	errs := Validate(m)
	found := false
	for _, e := range errs {
		if e.Error() == "[config.0] file path is required" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should report missing config file path")
	}
}

func TestValidate_EmptyConfigFieldPath(t *testing.T) {
	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Config: []ConfigFile{
			{File: "site.ts", Fields: []ConfigField{{Path: "", Type: "text"}}},
		},
	}

	errs := Validate(m)
	found := false
	for _, e := range errs {
		if e.Error() == "[config.0.fields.0] path is required" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should report missing config field path")
	}
}

func TestValidate_AllValidRuntimes(t *testing.T) {
	runtimes := map[string]string{
		"node": ">=20.0.0", "python": ">=3.12.0", "flutter": ">=3.22.0",
		"java": ">=21", "go": ">=1.22.0", "rust": "latest",
		"ruby": ">=3.3.0", "php": ">=8.3.0", "dotnet": ">=8.0.0",
	}

	m := &Manifest{
		Template: TemplateInfo{Name: "T", Version: "1.0.0"},
		Runtimes: runtimes,
	}

	errs := Validate(m)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors for valid runtimes: %v", len(errs), errs)
	}
}

func TestValidate_AllValidManagers(t *testing.T) {
	managers := []string{"npm", "pnpm", "yarn", "bun", "pip", "pub", "composer", "cargo", "go"}

	for _, mgr := range managers {
		m := &Manifest{
			Template: TemplateInfo{Name: "T", Version: "1.0.0"},
			Packages: PackageConfig{Manager: mgr},
		}
		errs := Validate(m)
		if len(errs) != 0 {
			t.Errorf("Validate() returned errors for valid manager %q: %v", mgr, errs)
		}
	}
}

func TestValidate_AllValidFieldTypes(t *testing.T) {
	types := []string{"text", "url", "email", "secret", "number", "boolean"}

	for _, typ := range types {
		m := &Manifest{
			Template: TemplateInfo{Name: "T", Version: "1.0.0"},
			Env:      []EnvVar{{Key: "TEST", Type: typ}},
		}
		errs := Validate(m)
		if len(errs) != 0 {
			t.Errorf("Validate() returned errors for valid type %q: %v", typ, errs)
		}
	}
}
