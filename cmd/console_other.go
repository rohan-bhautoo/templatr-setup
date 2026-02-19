//go:build !windows

package cmd

// attachConsole is a no-op on non-Windows platforms.
func attachConsole() {}
