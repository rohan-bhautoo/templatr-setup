package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/install"
	"github.com/templatr/templatr-setup/internal/state"
)

var uninstallAll bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove runtimes previously installed by this tool",
	Long: `Reads the state file (~/.templatr/state.json) and removes all runtimes
that were installed by templatr-setup. Does not touch runtimes that
were already installed before the tool ran.

If a runtime was upgraded (e.g., Node.js 20 → 22), uninstalling removes
the newer version and your original installation becomes active again.`,
	Run: func(cmd *cobra.Command, args []string) {
		runUninstall()
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "Remove all installed runtimes without prompting")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall() {
	st, err := state.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %s\n", err)
		os.Exit(1)
	}

	if len(st.Installations) == 0 {
		fmt.Println("No runtimes were installed by templatr-setup. Nothing to uninstall.")
		return
	}

	fmt.Println("The following runtimes were installed by templatr-setup:")
	fmt.Println()
	for _, inst := range st.Installations {
		action := "installed"
		if inst.Action == "upgrade" {
			action = fmt.Sprintf("upgraded from %s", inst.PreviousVersion)
		}
		fmt.Printf("  %s %s (%s)\n", inst.Runtime, inst.Version, action)
		fmt.Printf("    Path: %s\n", inst.Path)
		if inst.PreviousVersion != "" {
			fmt.Printf("    Will revert to: %s (%s)\n", inst.PreviousVersion, inst.PreviousPath)
		}
	}
	fmt.Println()

	if !uninstallAll {
		fmt.Print("Remove all of these? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Uninstall cancelled.")
			return
		}
	}

	results, errs := st.UndoAll()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  Error: %s\n", e)
		}
	}

	// Remove PATH and env var modifications
	for _, result := range results {
		if result.PathMod != nil {
			if err := install.RemoveFromPath(*result.PathMod); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not remove PATH entry %s: %s\n", result.PathMod.Value, err)
			}
		}
		for _, envMod := range result.EnvMods {
			if err := install.RemoveEnvVar(envMod); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not remove %s: %s\n", envMod.Name, err)
			}
		}

		// Inform user about reverts
		if result.Previous != nil {
			fmt.Printf("  Reverted %s → %s at %s\n", result.Previous.Runtime, result.Previous.Version, result.Previous.Path)
		}
	}

	if err := st.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save state: %s\n", err)
	}

	fmt.Println()
	fmt.Println("Uninstall complete. Restart your terminal for PATH changes to take effect.")
}
