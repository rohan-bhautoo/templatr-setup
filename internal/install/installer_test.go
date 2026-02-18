package install

import (
	"path/filepath"
	"testing"
)

func TestGetInstaller_Registered(t *testing.T) {
	runtimes := []string{"node", "python", "flutter", "java", "go", "rust", "ruby", "php", "dotnet"}

	for _, name := range runtimes {
		installer := GetInstaller(name)
		if installer == nil {
			t.Errorf("expected installer for %q, got nil", name)
			continue
		}
		if installer.Name() != name {
			t.Errorf("expected installer name %q, got %q", name, installer.Name())
		}
	}
}

func TestGetInstaller_NotRegistered(t *testing.T) {
	installer := GetInstaller("brainfuck")
	if installer != nil {
		t.Error("expected nil for unregistered runtime")
	}
}

func TestNodeInstaller_BinDir(t *testing.T) {
	n := &NodeInstaller{}
	dir := n.BinDir("/home/user/.templatr/runtimes/node/22.14.0")
	if dir == "" {
		t.Error("expected non-empty bin dir")
	}
}

func TestGoInstaller_BinDir(t *testing.T) {
	g := &GoInstaller{}
	dir := g.BinDir("/home/user/.templatr/runtimes/go/1.22.5")
	if dir == "" {
		t.Error("expected non-empty bin dir")
	}
}

func TestRustInstaller_BinDir(t *testing.T) {
	r := &RustInstaller{}
	base := filepath.Join("home", "user", ".templatr", "runtimes", "rust", "stable")
	dir := r.BinDir(base)
	expected := filepath.Join(base, ".cargo", "bin")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestPythonInstaller_ExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"cpython-3.13.2+20250212-x86_64-unknown-linux-gnu-install_only_stripped.tar.gz", "3.13.2"},
		{"cpython-3.12.0+20231002-aarch64-apple-darwin-install_only.tar.gz", "3.12.0"},
		{"cpython-3.11.7+20231212-x86_64-pc-windows-msvc-install_only.tar.gz", "3.11.7"},
		{"random-file.tar.gz", ""},
		{"cpython-noplus.tar.gz", ""},
	}

	for _, tt := range tests {
		result := extractPythonVersion(tt.name)
		if result != tt.expected {
			t.Errorf("extractPythonVersion(%q) = %q, want %q", tt.name, result, tt.expected)
		}
	}
}

func TestRubyInstaller_ResolveVersion_NotImplemented(t *testing.T) {
	r := &RubyInstaller{}
	_, err := r.ResolveVersion("latest")
	if err == nil {
		t.Error("expected error from unimplemented ruby installer")
	}
}

func TestPHPInstaller_ResolveVersion_NotImplemented(t *testing.T) {
	p := &PHPInstaller{}
	_, err := p.ResolveVersion("latest")
	if err == nil {
		t.Error("expected error from unimplemented php installer")
	}
}

func TestDotnetInstaller_ResolveVersion_NotImplemented(t *testing.T) {
	d := &DotnetInstaller{}
	_, err := d.ResolveVersion("latest")
	if err == nil {
		t.Error("expected error from unimplemented dotnet installer")
	}
}
