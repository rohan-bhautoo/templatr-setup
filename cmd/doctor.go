package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/detect"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system status: installed runtimes, versions, and PATH",
	Long:  `Scans your system for installed runtimes and reports their versions, locations, and PATH status.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("templatr-setup doctor - System Health Check\n")
		fmt.Printf("Version: %s (commit: %s, built: %s)\n\n", versionStr, commitStr, dateStr)

		sysInfo := detect.GetSystemInfo()
		fmt.Printf("OS:           %s\n", sysInfo.OS)
		fmt.Printf("Architecture: %s\n", sysInfo.Arch)
		fmt.Printf("Home:         %s\n\n", sysInfo.HomeDir)

		fmt.Println("Runtime Detection:")
		fmt.Println("─────────────────────────────────────────────────")

		runtimes := detect.ScanRuntimes()
		for _, r := range runtimes {
			status := "not found"
			if r.Installed {
				status = r.Version
			}
			icon := "✗"
			if r.Installed {
				icon = "✓"
			}
			fmt.Printf("  %s %-12s %s", icon, r.Name, status)
			if r.Installed && r.Path != "" {
				fmt.Printf("  (%s)", r.Path)
			}
			fmt.Println()
		}

		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
