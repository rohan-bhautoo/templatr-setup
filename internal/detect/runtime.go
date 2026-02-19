package detect

import (
	"os/exec"
	"strings"
)

// RuntimeInfo holds the detection result for a single runtime.
type RuntimeInfo struct {
	Name      string
	Installed bool
	Version   string
	Path      string
}

// runtimeCheck defines how to detect a runtime.
type runtimeCheck struct {
	Name       string
	Binary     string
	VersionArg string
}

var checks = []runtimeCheck{
	{Name: "Node.js", Binary: "node", VersionArg: "--version"},
	{Name: "npm", Binary: "npm", VersionArg: "--version"},
	{Name: "pnpm", Binary: "pnpm", VersionArg: "--version"},
	{Name: "yarn", Binary: "yarn", VersionArg: "--version"},
	{Name: "bun", Binary: "bun", VersionArg: "--version"},
	{Name: "Python", Binary: "python3", VersionArg: "--version"},
	{Name: "pip", Binary: "pip3", VersionArg: "--version"},
	{Name: "Flutter", Binary: "flutter", VersionArg: "--version"},
	{Name: "Dart", Binary: "dart", VersionArg: "--version"},
	{Name: "Java", Binary: "java", VersionArg: "--version"},
	{Name: "Go", Binary: "go", VersionArg: "version"},
	{Name: "Rust", Binary: "rustc", VersionArg: "--version"},
	{Name: "Cargo", Binary: "cargo", VersionArg: "--version"},
	{Name: "Ruby", Binary: "ruby", VersionArg: "--version"},
	{Name: "PHP", Binary: "php", VersionArg: "--version"},
	{Name: ".NET", Binary: "dotnet", VersionArg: "--version"},
	{Name: "Git", Binary: "git", VersionArg: "--version"},
}

// ScanRuntimes checks all known runtimes and returns their status.
func ScanRuntimes() []RuntimeInfo {
	results := make([]RuntimeInfo, 0, len(checks))

	for _, c := range checks {
		info := detectRuntime(c)
		results = append(results, info)
	}

	return results
}

// DetectRuntime checks a specific runtime by binary name.
func DetectRuntime(binary, versionArg string) RuntimeInfo {
	return detectRuntime(runtimeCheck{
		Name:       binary,
		Binary:     binary,
		VersionArg: versionArg,
	})
}

func detectRuntime(c runtimeCheck) RuntimeInfo {
	info := RuntimeInfo{Name: c.Name}

	// Find the binary on PATH.
	path, err := exec.LookPath(c.Binary)
	if err != nil {
		// On Windows, python3 might not exist - try "python" as fallback.
		switch c.Binary {
		case "python3":
			path, err = exec.LookPath("python")
			if err != nil {
				return info
			}
		case "pip3":
			path, err = exec.LookPath("pip")
			if err != nil {
				return info
			}
		default:
			return info
		}
	}

	// Get version.
	out, err := exec.Command(path, c.VersionArg).CombinedOutput()
	outStr := string(out)

	if err != nil {
		// Windows has stub executables (e.g. python3.exe in WindowsApps) that
		// appear on PATH but aren't actually installed. Detect these false positives.
		if isWindowsStub(outStr) {
			return info // not installed
		}
		// dotnet exists but may have no SDK installed.
		if c.Binary == "dotnet" && strings.Contains(outStr, "No .NET SDKs were found") {
			return info // not usable
		}
		info.Installed = true
		info.Path = path
		info.Version = "installed (version unknown)"
		return info
	}

	info.Installed = true
	info.Path = path
	info.Version = parseVersion(outStr)
	return info
}

// isWindowsStub detects Windows Store stub executables that look installed
// but just redirect to the Microsoft Store.
func isWindowsStub(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "microsoft store") ||
		strings.Contains(lower, "was not found") ||
		strings.Contains(lower, "app execution aliases")
}

// parseVersion extracts a clean version string from command output.
func parseVersion(output string) string {
	output = strings.TrimSpace(output)

	// Handle multiline output - take first line only.
	if idx := strings.IndexByte(output, '\n'); idx != -1 {
		output = output[:idx]
	}

	// Strip common prefixes: "v", "go version go", "Python ", etc.
	output = strings.TrimPrefix(output, "v")
	output = strings.TrimPrefix(output, "go version go")
	output = strings.TrimPrefix(output, "Python ")
	output = strings.TrimPrefix(output, "python ")
	output = strings.TrimPrefix(output, "Flutter ")
	output = strings.TrimPrefix(output, "flutter ")
	output = strings.TrimPrefix(output, "rustc ")
	output = strings.TrimPrefix(output, "ruby ")
	output = strings.TrimPrefix(output, "php ")
	output = strings.TrimPrefix(output, "git version ")
	output = strings.TrimPrefix(output, "Dart SDK version: ")

	// Trim trailing info after space (e.g., "3.12.0 (default, ...)")
	if idx := strings.IndexByte(output, ' '); idx != -1 {
		output = output[:idx]
	}

	return strings.TrimSpace(output)
}
