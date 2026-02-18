package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update to the latest version from GitHub Releases",
	Long: `Downloads the latest version of templatr-setup for your OS and architecture,
verifies the SHA256 checksum, and replaces the current binary.

If you installed via a package manager (Homebrew, Scoop, winget),
the tool will detect this and suggest using the package manager instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Self-update — not yet implemented.")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
