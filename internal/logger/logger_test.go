package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger_Init(t *testing.T) {
	l := New()
	if err := l.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer l.Close()

	fp := l.FilePath()
	if fp == "" {
		t.Error("FilePath() is empty after Init()")
	}

	// Verify the file exists
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("Log file does not exist at %s", fp)
	}

	// Clean up the test log file
	defer os.Remove(fp)
}

func TestLogger_MaskSecrets(t *testing.T) {
	l := New()
	l.AddSecret("my-secret-key")
	l.AddSecret("another-secret")

	msg := l.maskSecrets("The key is my-secret-key and another-secret is here")
	if strings.Contains(msg, "my-secret-key") {
		t.Error("maskSecrets() did not mask first secret")
	}
	if strings.Contains(msg, "another-secret") {
		t.Error("maskSecrets() did not mask second secret")
	}
	if !strings.Contains(msg, "****") {
		t.Error("maskSecrets() should contain masked placeholders")
	}
}

func TestLogger_AddEmptySecret(t *testing.T) {
	l := New()
	l.AddSecret("") // should not panic or add empty secret
	if len(l.secrets) != 0 {
		t.Error("AddSecret(\"\") should not add empty string to secrets")
	}
}

func TestRecentLogFilesInDir_NoLogs(t *testing.T) {
	dir := t.TempDir()
	files, err := RecentLogFilesInDir(dir, 10)
	if err != nil {
		t.Fatalf("RecentLogFilesInDir() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("RecentLogFilesInDir() returned %d files, want 0", len(files))
	}
}

func TestRecentLogFilesInDir_WithLogs(t *testing.T) {
	dir := t.TempDir()

	// Create some fake log files
	for _, name := range []string{
		"setup-2026-02-18_100000.log",
		"setup-2026-02-18_110000.log",
		"setup-2026-02-18_120000.log",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := RecentLogFilesInDir(dir, 10)
	if err != nil {
		t.Fatalf("RecentLogFilesInDir() error = %v", err)
	}
	if len(files) != 3 {
		t.Errorf("RecentLogFilesInDir() returned %d files, want 3", len(files))
	}

	// Should be sorted newest first
	if len(files) >= 2 && files[0] < files[1] {
		t.Error("RecentLogFilesInDir() should return newest first")
	}
}

func TestRecentLogFilesInDir_Limit(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 5; i++ {
		name := filepath.Join(dir, "setup-2026-02-18_1"+string(rune('0'+i))+"0000.log")
		if err := os.WriteFile(name, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := RecentLogFilesInDir(dir, 2)
	if err != nil {
		t.Fatalf("RecentLogFilesInDir(2) error = %v", err)
	}
	if len(files) != 2 {
		t.Errorf("RecentLogFilesInDir(2) returned %d files, want 2", len(files))
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}
