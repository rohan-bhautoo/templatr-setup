package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// RustInstaller handles Rust installation via rustup.
type RustInstaller struct{}

func (r *RustInstaller) Name() string { return "rust" }

func (r *RustInstaller) ResolveVersion(requirement string) (string, error) {
	// Rust always installs the latest stable via rustup.
	// The actual version is determined by rustup at install time.
	return "stable", nil
}

func (r *RustInstaller) Install(version, targetDir string, progress ProgressFunc) error {
	target := rustTarget()

	if runtime.GOOS == "windows" {
		return r.installWindows(targetDir, target, progress)
	}
	return r.installUnix(targetDir, target, progress)
}

func (r *RustInstaller) installUnix(targetDir, target string, progress ProgressFunc) error {
	// Download rustup-init
	url := fmt.Sprintf("https://static.rust-lang.org/rustup/dist/%s/rustup-init", target)
	tmpFile := filepath.Join(os.TempDir(), "rustup-init")
	defer os.Remove(tmpFile)

	if err := DownloadFile(url, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download rustup: %w", err)
	}

	// Make executable
	if err := os.Chmod(tmpFile, 0o755); err != nil {
		return err
	}

	// Run rustup-init with custom paths
	cargoHome := filepath.Join(targetDir, ".cargo")
	rustupHome := filepath.Join(targetDir, ".rustup")

	cmd := exec.Command(tmpFile, "--default-toolchain", "stable", "-y", "--no-modify-path")
	cmd.Env = append(os.Environ(),
		"CARGO_HOME="+cargoHome,
		"RUSTUP_HOME="+rustupHome,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rustup-init failed: %w", err)
	}

	return nil
}

func (r *RustInstaller) installWindows(targetDir, target string, progress ProgressFunc) error {
	url := fmt.Sprintf("https://static.rust-lang.org/rustup/dist/%s/rustup-init.exe", target)
	tmpFile := filepath.Join(os.TempDir(), "rustup-init.exe")
	defer os.Remove(tmpFile)

	if err := DownloadFile(url, tmpFile, progress); err != nil {
		return fmt.Errorf("failed to download rustup: %w", err)
	}

	cargoHome := filepath.Join(targetDir, ".cargo")
	rustupHome := filepath.Join(targetDir, ".rustup")

	cmd := exec.Command(tmpFile, "--default-toolchain", "stable", "-y", "--no-modify-path")
	cmd.Env = append(os.Environ(),
		"CARGO_HOME="+cargoHome,
		"RUSTUP_HOME="+rustupHome,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rustup-init failed: %w", err)
	}

	return nil
}

func (r *RustInstaller) BinDir(installDir string) string {
	return filepath.Join(installDir, ".cargo", "bin")
}

func (r *RustInstaller) EnvVars(installDir string) map[string]string { return nil }

func rustTarget() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch {
	case os == "linux" && arch == "amd64":
		return "x86_64-unknown-linux-gnu"
	case os == "linux" && arch == "arm64":
		return "aarch64-unknown-linux-gnu"
	case os == "darwin" && arch == "amd64":
		return "x86_64-apple-darwin"
	case os == "darwin" && arch == "arm64":
		return "aarch64-apple-darwin"
	case os == "windows" && arch == "amd64":
		return "x86_64-pc-windows-msvc"
	case os == "windows" && arch == "arm64":
		return "aarch64-pc-windows-msvc"
	default:
		return "x86_64-unknown-linux-gnu"
	}
}
