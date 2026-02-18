package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidManifest(t *testing.T) {
	content := `
[template]
name = "Test Template"
version = "1.0.0"
tier = "starter"
category = "website"
slug = "test-template"

[runtimes]
node = ">=20.0.0"

[packages]
manager = "npm"
install_command = "npm install"

[[env]]
key = "API_KEY"
label = "API Key"
default = ""
required = true
type = "secret"

[post_setup]
commands = ["npm run build"]
message = "Done!"

[meta]
min_tool_version = "1.0.0"
docs = "https://example.com"
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".templatr.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if m.Template.Name != "Test Template" {
		t.Errorf("Template.Name = %q, want %q", m.Template.Name, "Test Template")
	}
	if m.Template.Version != "1.0.0" {
		t.Errorf("Template.Version = %q, want %q", m.Template.Version, "1.0.0")
	}
	if m.Template.Tier != "starter" {
		t.Errorf("Template.Tier = %q, want %q", m.Template.Tier, "starter")
	}
	if m.Runtimes["node"] != ">=20.0.0" {
		t.Errorf("Runtimes[node] = %q, want %q", m.Runtimes["node"], ">=20.0.0")
	}
	if m.Packages.Manager != "npm" {
		t.Errorf("Packages.Manager = %q, want %q", m.Packages.Manager, "npm")
	}
	if len(m.Env) != 1 {
		t.Errorf("len(Env) = %d, want 1", len(m.Env))
	}
	if m.Env[0].Key != "API_KEY" {
		t.Errorf("Env[0].Key = %q, want %q", m.Env[0].Key, "API_KEY")
	}
	if m.Env[0].Type != "secret" {
		t.Errorf("Env[0].Type = %q, want %q", m.Env[0].Type, "secret")
	}
	if m.Meta.MinToolVersion != "1.0.0" {
		t.Errorf("Meta.MinToolVersion = %q, want %q", m.Meta.MinToolVersion, "1.0.0")
	}
}

func TestLoad_MultipleRuntimes(t *testing.T) {
	content := `
[template]
name = "Multi"
version = "1.0.0"

[runtimes]
node = ">=20.0.0"
python = ">=3.12.0"
flutter = ">=3.22.0"
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".templatr.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(m.Runtimes) != 3 {
		t.Errorf("len(Runtimes) = %d, want 3", len(m.Runtimes))
	}
	if m.Runtimes["python"] != ">=3.12.0" {
		t.Errorf("Runtimes[python] = %q, want %q", m.Runtimes["python"], ">=3.12.0")
	}
}

func TestLoad_ConfigFields(t *testing.T) {
	content := `
[template]
name = "Config Test"
version = "1.0.0"

[[config]]
file = "src/config/site.ts"
label = "Site Config"
description = "Site configuration"

  [[config.fields]]
  path = "siteConfig.name"
  label = "Site Name"
  type = "text"
  default = "MySite"

  [[config.fields]]
  path = "siteConfig.url"
  label = "Site URL"
  type = "url"
  default = "https://example.com"
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".templatr.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(m.Config) != 1 {
		t.Fatalf("len(Config) = %d, want 1", len(m.Config))
	}
	if m.Config[0].File != "src/config/site.ts" {
		t.Errorf("Config[0].File = %q, want %q", m.Config[0].File, "src/config/site.ts")
	}
	if len(m.Config[0].Fields) != 2 {
		t.Fatalf("len(Config[0].Fields) = %d, want 2", len(m.Config[0].Fields))
	}
	if m.Config[0].Fields[0].Path != "siteConfig.name" {
		t.Errorf("Fields[0].Path = %q, want %q", m.Config[0].Fields[0].Path, "siteConfig.name")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/.templatr.toml")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent file")
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".templatr.toml")
	if err := os.WriteFile(path, []byte("this is not valid toml [[["), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for invalid TOML")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".templatr.toml")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error for empty file: %v", err)
	}
	// Empty file should parse but produce empty manifest
	if m.Template.Name != "" {
		t.Errorf("Template.Name = %q, want empty", m.Template.Name)
	}
}

func TestLoad_AutoDetect(t *testing.T) {
	// Create a temp dir with .templatr.toml and cd into it
	dir := t.TempDir()
	content := `
[template]
name = "Auto Detect"
version = "1.0.0"
`
	path := filepath.Join(dir, ".templatr.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Save current dir and restore after test
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	m, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error = %v", err)
	}
	if m.Template.Name != "Auto Detect" {
		t.Errorf("Template.Name = %q, want %q", m.Template.Name, "Auto Detect")
	}
}
