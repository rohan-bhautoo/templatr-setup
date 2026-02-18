package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/selfupdate"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show tool version and check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("templatr-setup %s\n", versionStr)
		fmt.Printf("  commit: %s\n", commitStr)
		fmt.Printf("  built:  %s\n", dateStr)

		if result := selfupdate.CheckForUpdate(versionStr); result != nil && result.UpdateAvail {
			fmt.Println()
			fmt.Printf("A new version is available: %s (current: %s)\n", result.LatestVersion, result.CurrentVersion)
			fmt.Println("Run 'templatr-setup update' to upgrade, or visit https://templatr.io/tools/setup")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
