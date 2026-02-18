package detect

import (
	"os"
	"runtime"
)

// SystemInfo holds information about the user's system.
type SystemInfo struct {
	OS      string
	Arch    string
	HomeDir string
}

// GetSystemInfo returns the current system's OS, architecture, and home directory.
func GetSystemInfo() SystemInfo {
	homeDir, _ := os.UserHomeDir()
	return SystemInfo{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		HomeDir: homeDir,
	}
}
