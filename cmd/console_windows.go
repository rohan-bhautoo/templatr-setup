package cmd

import (
	"os"
	"syscall"
)

var (
	kernel32   = syscall.NewLazyDLL("kernel32.dll")
	procAttach = kernel32.NewProc("AttachConsole")
)

const attachParentProcess = ^uintptr(0) // ATTACH_PARENT_PROCESS = (DWORD)-1

// attachConsole attaches to the parent process's console so stdout/stderr
// work when running from cmd/PowerShell. This is needed because we build
// with -H windowsgui to avoid a console window on double-click.
func attachConsole() {
	r, _, _ := procAttach.Call(attachParentProcess)
	if r == 0 {
		return // no parent console (double-clicked) - nothing to attach
	}

	// Reopen stdout/stderr to the attached console
	con, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return
	}
	os.Stdout = con
	os.Stderr = con
}
