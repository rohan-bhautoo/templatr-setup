package detect

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v20.0.0", "20.0.0"},
		{"v25.2.1\n", "25.2.1"},
		{"Python 3.12.0", "3.12.0"},
		{"python 3.12.0", "3.12.0"},
		{"go version go1.26.0 windows/amd64", "1.26.0"},
		{"rustc 1.75.0 (some hash)", "1.75.0"},
		{"ruby 3.3.0 (2023-12-25 revision 5124f9ac75)", "3.3.0"},
		{"git version 2.52.0.windows.1", "2.52.0.windows.1"},
		{"Dart SDK version: 3.3.0 (stable)", "3.3.0"},
		{"php 8.3.0 (cli)", "8.3.0"},
		{"10.8.0\n", "10.8.0"},
		{"10.8.0", "10.8.0"},
		{"  v20.0.0  \n", "20.0.0"},
		// Multiline output
		{"Flutter 3.22.0\nEngine • revision abc\nTools • Dart 3.3.0", "3.22.0"},
	}

	for _, tt := range tests {
		got := parseVersion(tt.input)
		if got != tt.want {
			t.Errorf("parseVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsWindowsStub(t *testing.T) {
	tests := []struct {
		output string
		want   bool
	}{
		{"Python was not found; run without arguments to install from the Microsoft Store", true},
		{"App execution aliases", true},
		{"Python 3.12.0", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isWindowsStub(tt.output)
		if got != tt.want {
			t.Errorf("isWindowsStub(%q) = %v, want %v", tt.output, got, tt.want)
		}
	}
}

func TestGetSystemInfo(t *testing.T) {
	info := GetSystemInfo()

	if info.OS == "" {
		t.Error("GetSystemInfo().OS is empty")
	}
	if info.Arch == "" {
		t.Error("GetSystemInfo().Arch is empty")
	}
	if info.HomeDir == "" {
		t.Error("GetSystemInfo().HomeDir is empty")
	}
}

func TestScanRuntimes(t *testing.T) {
	runtimes := ScanRuntimes()

	if len(runtimes) == 0 {
		t.Error("ScanRuntimes() returned no results")
	}

	// Should always include at least these entries
	names := make(map[string]bool)
	for _, r := range runtimes {
		names[r.Name] = true
	}

	expected := []string{"Node.js", "npm", "Python", "Git", "Go"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("ScanRuntimes() missing expected runtime %q", name)
		}
	}
}
