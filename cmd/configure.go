package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure .env and site.ts files for your template",
	Long: `Reads the configuration definitions from .templatr.toml and presents
an interactive form to fill out .env variables and site.ts fields.
Values are written directly to the template files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configure step — not yet implemented.")
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
