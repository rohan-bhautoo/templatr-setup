package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/engine"
	"github.com/templatr/templatr-setup/internal/install"
	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/manifest"
	"github.com/templatr/templatr-setup/internal/packages"
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
	fmt.Println("templatr-setup — Template dependency installer")
	fmt.Printf("Version: %s (commit: %s, built: %s)\n\n", versionStr, commitStr, dateStr)

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

	// Display summary
	engine.PrintSummary(plan)

	if dryRun {
		fmt.Println("Dry run mode — no changes were made.")
		return
	}

	if !plan.NeedsAction() {
		fmt.Println("Nothing to install — all requirements are satisfied.")
		fmt.Println("Run 'templatr-setup configure' to set up environment variables and config files.")
		return
	}

	// Confirm with user
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

	// Execute installation plan
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
			fmt.Fprintf(os.Stderr, "See log file for details: %s\n", log.FilePath())
		}
		os.Exit(1)
	}

	fmt.Println() // newline after progress bar

	// Install global packages
	if err := packages.RunGlobalInstalls(m, log); err != nil {
		log.Warn("Global package installation had issues: %s", err)
	}

	// Run package install command
	fmt.Println()
	if m.Packages.InstallCommand != "" {
		log.Info("Running package install: %s", m.Packages.InstallCommand)
		if err := packages.RunInstall(m, log); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
			log.Warn("Package install failed: %s", err)
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("Installation complete!")
	fmt.Println()
	for _, r := range results {
		fmt.Printf("  ✓ %s %s → %s\n", r.Runtime, r.Version, r.InstallPath)
	}

	// Post-setup hint
	fmt.Println()
	if len(m.Env) > 0 || len(m.Config) > 0 {
		fmt.Println("Run 'templatr-setup configure' to set up environment variables and config files.")
	}

	// Post-setup commands
	if len(m.PostSetup.Commands) > 0 {
		fmt.Println()
		log.Info("Running post-setup commands...")
		if err := packages.RunPostSetup(m, log); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: post-setup command failed: %s\n", err)
			log.Warn("Post-setup failed: %s", err)
		}
	}

	// Success message
	if m.PostSetup.Message != "" {
		fmt.Println()
		fmt.Println(strings.TrimSpace(m.PostSetup.Message))
	}

	if log.FilePath() != "" {
		fmt.Printf("\nLog file: %s\n", log.FilePath())
	}

	log.Info("Setup completed successfully")
}
