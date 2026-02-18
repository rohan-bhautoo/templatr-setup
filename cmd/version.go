package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show tool version and check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("templatr-setup %s\n", versionStr)
		fmt.Printf("  commit: %s\n", commitStr)
		fmt.Printf("  built:  %s\n", dateStr)
		// TODO: check for updates via GitHub API
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
