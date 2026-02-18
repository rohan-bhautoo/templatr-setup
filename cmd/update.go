package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/selfupdate"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update to the latest version from GitHub Releases",
	Long: `Downloads the latest version of templatr-setup for your OS and architecture,
verifies the SHA256 checksum, and replaces the current binary.

If you installed via a package manager (Homebrew, Scoop, winget),
the tool will detect this and suggest using the package manager instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", versionStr)
		fmt.Println("Checking for updates...")

		if err := selfupdate.DoUpdate(versionStr); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Update successful! Restart templatr-setup to use the new version.")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
