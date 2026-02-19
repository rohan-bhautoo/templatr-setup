package cmd

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procAttach                = kernel32.NewProc("AttachConsole")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

const attachParentProcess = ^uintptr(0) // ATTACH_PARENT_PROCESS = (DWORD)-1

// parentConsoleAttached is true when the process was launched from a terminal
// (cmd.exe, PowerShell, Windows Terminal), false when double-clicked from Explorer.
var parentConsoleAttached bool

// attachConsole detects whether the process was launched from a terminal or
// double-clicked from Explorer, and reattaches stdout/stderr when needed.
//
// For -H windowsgui builds (release): the process starts without a console.
// AttachConsole succeeds when run from a terminal, fails when double-clicked.
//
// For normal console builds (development): the process already has a console.
// AttachConsole always fails, so we fall back to GetConsoleProcessList - if
// multiple processes share the console, we inherited it from a terminal shell.
// If only our process is attached, Windows created a fresh console (double-click).
func attachConsole() {
	r, _, _ := procAttach.Call(attachParentProcess)
	if r != 0 {
		// Successfully attached to parent console (-H windowsgui, run from terminal)
		parentConsoleAttached = true

		// Reopen stdout/stderr to the attached console
		con, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
		if err != nil {
			return
		}
		os.Stdout = con
		os.Stderr = con
		return
	}

	// AttachConsole failed - either no console (-H windowsgui, double-click)
	// or already have one (console subsystem build). Use GetConsoleProcessList
	// to check if we share the console with a parent shell process.
	var pids [16]uint32
	count, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)
	// count > 1 means other processes share our console (cmd.exe, powershell, etc.)
	// count <= 1 means we're alone (double-click created a fresh console, or no console)
	parentConsoleAttached = count > 1
}

// launchedFromConsole returns true if the process was started from a terminal
// (cmd.exe, PowerShell, Windows Terminal), false if double-clicked from Explorer.
func launchedFromConsole() bool {
	return parentConsoleAttached
}
