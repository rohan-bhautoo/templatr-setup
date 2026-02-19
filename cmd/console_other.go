//go:build !windows

package cmd

// attachConsole is a no-op on non-Windows platforms.
func attachConsole() {}

// launchedFromConsole returns true if the process is running in a terminal.
// On Unix, this is determined by checking if stdin is a terminal device.
func launchedFromConsole() bool {
	return isTerminal()
}
