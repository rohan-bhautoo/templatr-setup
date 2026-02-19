package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/manifest"
	"github.com/templatr/templatr-setup/internal/selfupdate"
	"github.com/templatr/templatr-setup/internal/server"
	"golang.org/x/term"
)

var (
	versionStr string
	commitStr  string
	dateStr    string
	uiFlag     bool
	webAssets  embed.FS
)

// SetVersionInfo sets the version info from ldflags.
func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
}

// SetWebAssets sets the embedded web UI assets.
func SetWebAssets(assets embed.FS) {
	webAssets = assets
}

var rootCmd = &cobra.Command{
	Use:   "templatr-setup",
	Short: "Template setup and dependency installer for Templatr templates",
	Long: `templatr-setup reads a .templatr.toml manifest file from your template
directory and automatically installs all required runtimes, packages,
and dependencies. It also helps configure .env and site.ts files
through an interactive interface.

For developers: run in your terminal for an interactive TUI experience.
For everyone else: double-click the downloaded file to open the visual
web dashboard in your browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		if uiFlag {
			launchWebUI()
			return
		}
		if !hasManifestAvailable() {
			printNoManifestHelp()
			return
		}
		runSetup(cmd, args)
	},
}

// Execute is the entry point called from main.
func Execute() {
	// Attach to parent console first. On Windows with -H windowsgui, this
	// detects whether we were launched from a terminal (cmd/PowerShell) or
	// double-clicked from Explorer. Must run before shouldLaunchWebUI().
	attachConsole()

	// If double-clicked from Explorer (no terminal), launch the web UI
	// directly, bypassing cobra.
	if shouldLaunchWebUI() {
		launchWebUI()
		return
	}

	// Non-blocking update check (runs in background, prints notice after command)
	updateCh := make(chan *selfupdate.CheckResult, 1)
	go func() {
		updateCh <- selfupdate.CheckForUpdate(versionStr)
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	// Show update notice if available
	select {
	case result := <-updateCh:
		if result != nil && result.UpdateAvail {
			fmt.Fprintf(os.Stderr, "\nA new version of templatr-setup is available: %s (current: %s)\n", result.LatestVersion, result.CurrentVersion)
			fmt.Fprintf(os.Stderr, "Run 'templatr-setup update' to upgrade, or visit https://templatr.io/tools/setup\n")
		}
	default:
		// Don't wait - check hasn't finished yet, skip the notice
	}
}

// shouldLaunchWebUI checks if we should bypass cobra and launch the web UI.
// Returns true when there are no CLI args and the process was not launched
// from a terminal - the typical double-click from file explorer scenario.
// On Windows, this is detected by AttachConsole (set by attachConsole()).
// On Unix, this is detected by checking if stdin is a terminal.
func shouldLaunchWebUI() bool {
	// If there are CLI args beyond the exe name, let cobra handle them
	if len(os.Args) > 1 {
		return false
	}
	// No args - launch web UI only when not from a terminal (double-click)
	return !launchedFromConsole()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&uiFlag, "ui", false, "Launch the visual web dashboard in your browser")
	rootCmd.PersistentFlags().StringVarP(&manifestFile, "file", "f", "", "Path to .templatr.toml manifest file")
}

// hasManifestAvailable checks if a manifest file is available for CLI mode.
// Returns true if --file was specified or .templatr.toml exists in the CWD.
// When false, the web UI is launched instead (e.g. double-click from Explorer).
func hasManifestAvailable() bool {
	if manifestFile != "" {
		return true
	}
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(cwd, manifest.DefaultManifestName))
	return err == nil
}

// isTerminal checks if stdin is connected to a terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func launchWebUI() {
	log := logger.New()
	if err := log.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize logger: %s\n", err)
	} else {
		defer log.Close()
	}

	srv := server.New(webAssets, log, manifestFile)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func printNoManifestHelp() {
	fmt.Println("No .templatr.toml manifest found in the current directory.")
	fmt.Println()
	fmt.Println("Navigate to a template directory containing a .templatr.toml file,")
	fmt.Println("or specify one with the -f flag:")
	fmt.Println()
	fmt.Println("  templatr-setup -f <path>    Use a specific manifest file")
	fmt.Println("  templatr-setup --ui         Open the visual web dashboard")
	fmt.Println("  templatr-setup doctor       Check system and installed runtimes")
	fmt.Println("  templatr-setup help         Show all available commands")
}

func runSetup(_ *cobra.Command, _ []string) {
	// When invoked via the root command (no subcommand), run setup
	runSetupCommand()
}
