package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/logger"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show recent log files",
	Long:  `Lists and displays recent templatr-setup log files stored in ~/.templatr/logs/.`,
	Run: func(cmd *cobra.Command, args []string) {
		files, err := logger.RecentLogFiles(10)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading log files: %s\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No log files found.")
			fmt.Println("Log files are created when you run 'templatr-setup setup'.")
			return
		}

		fmt.Println("Recent log files:")
		fmt.Println()
		for i, f := range files {
			info, err := os.Stat(f)
			if err != nil {
				continue
			}
			fmt.Printf("  %d. %s  (%s)\n", i+1, filepath.Base(f), formatSize(info.Size()))
		}

		fmt.Printf("\nLog directory: %s\n", filepath.Dir(files[0]))
		fmt.Printf("\nTo view the latest log:\n  cat %s\n", files[0])
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
