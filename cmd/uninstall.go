package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove runtimes previously installed by this tool",
	Long: `Reads the state file (~/.templatr/state.json) and removes all runtimes
that were installed by templatr-setup. Does not touch runtimes that
were already installed before the tool ran.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Uninstall — not yet implemented.")
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
