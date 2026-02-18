package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/engine"
	"github.com/templatr/templatr-setup/internal/manifest"
)

var (
	manifestFile string
	dryRun       bool
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
	rootCmd.AddCommand(setupCmd)
}

func runSetupCommand() {
	fmt.Println("templatr-setup — Template dependency installer")
	fmt.Printf("Version: %s (commit: %s, built: %s)\n\n", versionStr, commitStr, dateStr)

	// Load manifest
	m, err := manifest.Load(manifestFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Validate manifest
	errs := manifest.Validate(m)
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Manifest validation errors:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(1)
	}

	// Build plan
	plan, err := engine.BuildPlan(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building setup plan: %s\n", err)
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

	// TODO: implement actual installation flow
	fmt.Println("Installation not yet implemented. Use --dry-run to preview the plan.")
}
