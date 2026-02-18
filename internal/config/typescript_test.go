package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "site.ts")

	original := `export const siteConfig = {
  name: "SaaSify",
  description: "Ship your SaaS faster",
  url: "https://your-domain.com",
  contact: {
    email: "hello@your-domain.com",
  },
  links: {
    twitter: "https://twitter.com/yourusername",
    github: "https://github.com/yourusername",
  },
};
`
	os.WriteFile(path, []byte(original), 0o644)

	fieldValues := map[string]string{
		"siteConfig.name":            "My SaaS Product",
		"siteConfig.description":     "The best SaaS ever built",
		"siteConfig.url":             "https://mysaas.com",
		"siteConfig.contact.email":   "contact@mysaas.com",
		"siteConfig.links.twitter":   "https://twitter.com/mysaas",
		"siteConfig.links.github":    "https://github.com/mysaas",
	}

	if err := UpdateConfigFile(path, fieldValues); err != nil {
		t.Fatalf("UpdateConfigFile failed: %s", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	tests := map[string]string{
		`name: "My SaaS Product"`:             "name",
		`description: "The best SaaS ever built"`: "description",
		`url: "https://mysaas.com"`:            "url",
		`email: "contact@mysaas.com"`:          "email",
		`twitter: "https://twitter.com/mysaas"`: "twitter",
		`github: "https://github.com/mysaas"`:  "github",
	}

	for expected, field := range tests {
		if !strings.Contains(content, expected) {
			t.Errorf("field %s: expected %q in output, got:\n%s", field, expected, content)
		}
	}
}

func TestUpdateConfigFile_SingleQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.ts")

	original := `const config = {
  title: 'Default Title',
  apiUrl: 'https://api.example.com',
};
`
	os.WriteFile(path, []byte(original), 0o644)

	err := UpdateConfigFile(path, map[string]string{
		"config.title": "New Title",
	})
	if err != nil {
		t.Fatalf("UpdateConfigFile failed: %s", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// Should preserve single quotes
	if !strings.Contains(content, `title: 'New Title'`) {
		t.Errorf("expected single-quoted replacement, got:\n%s", content)
	}
	// Should not touch apiUrl
	if !strings.Contains(content, `apiUrl: 'https://api.example.com'`) {
		t.Errorf("apiUrl should be unchanged, got:\n%s", content)
	}
}

func TestUpdateConfigFile_NotFound(t *testing.T) {
	err := UpdateConfigFile("/nonexistent/site.ts", map[string]string{"a": "b"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReplaceFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		key      string
		newValue string
		expected string
	}{
		{
			name:     "double quotes",
			content:  `  name: "old",`,
			key:      "name",
			newValue: "new",
			expected: `  name: "new",`,
		},
		{
			name:     "single quotes",
			content:  `  name: 'old',`,
			key:      "name",
			newValue: "new",
			expected: `  name: 'new',`,
		},
		{
			name:     "with spaces",
			content:  `  name:   "old"  ,`,
			key:      "name",
			newValue: "new",
			expected: `  name:   "new"  ,`,
		},
		{
			name:     "escapes quotes in value",
			content:  `  name: "old",`,
			key:      "name",
			newValue: `he said "hello"`,
			expected: `  name: "he said \"hello\"",`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceFieldValue(tt.content, tt.key, tt.newValue)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
