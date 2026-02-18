package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/templatr/templatr-setup/internal/manifest"
)

func TestWriteEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	envDefs := []manifest.EnvVar{
		{Key: "SITE_URL", Label: "Site URL", Description: "Your site URL", Default: "http://localhost:3000", Type: "url"},
		{Key: "API_KEY", Label: "API Key", Description: "Get from dashboard", Type: "secret", DocsURL: "https://example.com/docs"},
		{Key: "CONTACT_EMAIL", Label: "Contact", Type: "email"},
	}

	values := map[string]string{
		"SITE_URL":      "https://example.com",
		"API_KEY":       "sk_test_12345",
		"CONTACT_EMAIL": "hello@example.com",
	}

	if err := WriteEnvFile(path, envDefs, values); err != nil {
		t.Fatalf("WriteEnvFile failed: %s", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %s", err)
	}

	content := string(data)

	if !strings.Contains(content, "SITE_URL=https://example.com") {
		t.Error("expected SITE_URL in output")
	}
	if !strings.Contains(content, "API_KEY=sk_test_12345") {
		t.Error("expected API_KEY in output")
	}
	if !strings.Contains(content, "CONTACT_EMAIL=hello@example.com") {
		t.Error("expected CONTACT_EMAIL in output")
	}
	if !strings.Contains(content, "# Your site URL") {
		t.Error("expected description comment")
	}
	if !strings.Contains(content, "# Docs: https://example.com/docs") {
		t.Error("expected docs URL comment")
	}
}

func TestWriteEnvFile_QuotesSpecialChars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	envDefs := []manifest.EnvVar{
		{Key: "MSG", Label: "Message", Type: "text"},
	}
	values := map[string]string{
		"MSG": "hello world with spaces",
	}

	if err := WriteEnvFile(path, envDefs, values); err != nil {
		t.Fatalf("WriteEnvFile failed: %s", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `MSG="hello world with spaces"`) {
		t.Errorf("expected quoted value, got: %s", string(data))
	}
}

func TestReadEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `# comment
SITE_URL=https://example.com
API_KEY="sk_test_12345"
EMPTY=
QUOTED='single quotes'
`
	os.WriteFile(path, []byte(content), 0o644)

	vals, err := ReadEnvFile(path)
	if err != nil {
		t.Fatalf("ReadEnvFile failed: %s", err)
	}

	tests := map[string]string{
		"SITE_URL": "https://example.com",
		"API_KEY":  "sk_test_12345",
		"EMPTY":    "",
		"QUOTED":   "single quotes",
	}

	for key, expected := range tests {
		if got := vals[key]; got != expected {
			t.Errorf("key %s: expected %q, got %q", key, expected, got)
		}
	}
}

func TestReadEnvFile_NotFound(t *testing.T) {
	vals, err := ReadEnvFile("/nonexistent/.env")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %s", err)
	}
	if len(vals) != 0 {
		t.Errorf("expected empty map, got %d entries", len(vals))
	}
}
