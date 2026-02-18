package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/engine"
	"github.com/templatr/templatr-setup/internal/install"
	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/manifest"
	"github.com/templatr/templatr-setup/internal/packages"
	"github.com/templatr/templatr-setup/internal/tui"
)

var (
	manifestFile string
	dryRun       bool
	yesFlag      bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install all dependencies defined in .templatr.toml",
	Long: `Reads the .templatr.toml manifest file from the current directory
(or a specified path) and installs all required runtimes and packages.

After installation, optionally runs the configure step to set up
.env and site.ts files through an interactive form.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSetupCommand()
	},
}

func init() {
	setupCmd.Flags().StringVarP(&manifestFile, "file", "f", "", "Path to .templatr.toml manifest file")
	setupCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without installing")
	setupCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.AddCommand(setupCmd)
}

func runSetupCommand() {
	// Initialize logger
	log := logger.New()
	if err := log.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize logger: %s\n", err)
	} else {
		defer log.Close()
		log.Info("templatr-setup %s started", versionStr)
	}

	// Load manifest
	m, err := manifest.Load(manifestFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		log.Error("Failed to load manifest: %s", err)
		os.Exit(1)
	}
	log.Info("Loaded manifest: %s (%s)", m.Template.Name, m.Template.Tier)

	// Validate manifest
	errs := manifest.Validate(m)
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Manifest validation errors:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
			log.Error("Validation: %s", e)
		}
		os.Exit(1)
	}

	// Build plan
	plan, err := engine.BuildPlan(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building setup plan: %s\n", err)
		log.Error("Failed to build plan: %s", err)
		os.Exit(1)
	}

	// Dry run: print summary and exit
	if dryRun {
		engine.PrintSummary(plan)
		fmt.Println("Dry run mode — no changes were made.")
		return
	}

	// Interactive TUI mode when running in a terminal
	if isTerminal() {
		tuiModel := tui.New(plan, log, yesFlag)
		p := tea.NewProgram(tuiModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %s\n", err)
			log.Error("TUI error: %s", err)
			os.Exit(1)
		}
		return
	}

	// Fallback: non-interactive plain text mode (CI, pipes, etc.)
	runSetupPlainText(plan, m, log)
}

// runSetupPlainText is the non-TUI fallback for non-interactive environments.
func runSetupPlainText(plan *engine.SetupPlan, m *manifest.Manifest, log *logger.Logger) {
	fmt.Println("templatr-setup — Template dependency installer")
	fmt.Printf("Version: %s\n\n", versionStr)

	engine.PrintSummary(plan)

	if !plan.NeedsAction() {
		fmt.Println("Nothing to install — all requirements are satisfied.")
		return
	}

	// Confirm
	if !yesFlag {
		fmt.Print("Proceed with installation? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Installation cancelled.")
			return
		}
	}

	fmt.Println()
	log.Info("Starting installation...")

	progress := func(downloaded, total int64) {
		if total > 0 {
			pct := float64(downloaded) / float64(total) * 100
			fmt.Printf("\r  Downloading... %.0f%% (%d / %d MB)", pct, downloaded/(1024*1024), total/(1024*1024))
		} else {
			fmt.Printf("\r  Downloading... %d MB", downloaded/(1024*1024))
		}
	}

	results, err := install.ExecutePlan(plan, log, progress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
		log.Error("Installation failed: %s", err)
		if log.FilePath() != "" {
			fmt.Fprintf(os.Stderr, "See log file: %s\n", log.FilePath())
		}
		os.Exit(1)
	}

	fmt.Println()

	if err := packages.RunGlobalInstalls(m, log); err != nil {
		log.Warn("Global install issues: %s", err)
	}

	if m.Packages.InstallCommand != "" {
		log.Info("Running: %s", m.Packages.InstallCommand)
		if err := packages.RunInstall(m, log); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
	}

	fmt.Println()
	fmt.Println("Installation complete!")
	for _, r := range results {
		fmt.Printf("  ✓ %s %s → %s\n", r.Runtime, r.Version, r.InstallPath)
	}

	if len(m.PostSetup.Commands) > 0 {
		fmt.Println()
		log.Info("Running post-setup commands...")
		if err := packages.RunPostSetup(m, log); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
	}

	if m.PostSetup.Message != "" {
		fmt.Println()
		fmt.Println(strings.TrimSpace(m.PostSetup.Message))
	}

	if log.FilePath() != "" {
		fmt.Printf("\nLog file: %s\n", log.FilePath())
	}
}
