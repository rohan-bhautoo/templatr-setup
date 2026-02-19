package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/logger"
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
		if uiFlag || !isTerminal() {
			launchWebUI()
			return
		}
		runSetup(cmd, args)
	},
}

// Execute is the entry point called from main.
func Execute() {
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
		// Don't wait — check hasn't finished yet, skip the notice
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&uiFlag, "ui", false, "Launch the visual web dashboard in your browser")
	rootCmd.PersistentFlags().StringVarP(&manifestFile, "file", "f", "", "Path to .templatr.toml manifest file")
}

// isTerminal checks if stdin is connected to a terminal.
// Returns false when the binary is double-clicked (no terminal attached).
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

func runSetup(cmd *cobra.Command, args []string) {
	// When invoked via the root command (no subcommand), run setup
	runSetupCommand()
}
