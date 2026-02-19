package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/templatr/templatr-setup/internal/config"
	"github.com/templatr/templatr-setup/internal/logger"
	"github.com/templatr/templatr-setup/internal/manifest"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure .env and site.ts files for your template",
	Long: `Reads the configuration definitions from .templatr.toml and presents
an interactive form to fill out .env variables and site.ts fields.
Values are written directly to the template files.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConfigure()
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

func runConfigure() {
	log := logger.New()
	if err := log.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize logger: %s\n", err)
	} else {
		defer log.Close()
	}

	m, err := manifest.Load(manifestFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if len(m.Env) == 0 && len(m.Config) == 0 {
		fmt.Println("No configuration fields defined in the manifest.")
		return
	}

	// Read existing env values from all target files to pre-fill
	existingEnv := make(map[string]string)
	grouped, fileOrder := config.GroupEnvByFile(m.Env)
	for _, file := range fileOrder {
		existing, _ := config.ReadEnvFile(file)
		for k, v := range existing {
			existingEnv[k] = v
		}
	}

	// Plain text interactive mode
	reader := bufio.NewReader(os.Stdin)
	envValues := make(map[string]string)

	if len(m.Env) > 0 {
		currentFile := ""
		for _, env := range m.Env {
			// Show file header when switching to a new target file
			target := config.EnvFileTarget(env)
			if target != currentFile {
				if currentFile != "" {
					fmt.Println()
				}
				currentFile = target
				fmt.Printf("Environment Variables (%s)\n", target)
				fmt.Println(strings.Repeat("─", 40))
				fmt.Println()
			}

			defaultVal := env.Default
			if existing, ok := existingEnv[env.Key]; ok && existing != "" {
				defaultVal = existing
			}

			label := env.Label
			if env.Required {
				label += " *"
			}

			fmt.Printf("  %s\n", label)
			if env.Description != "" {
				fmt.Printf("  %s\n", env.Description)
			}
			if defaultVal != "" {
				fmt.Printf("  [default: %s]\n", defaultVal)
			}
			fmt.Print("  > ")

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				input = defaultVal
			}
			envValues[env.Key] = input
			fmt.Println()
		}

		log.Info("Writing env files...")
		for _, file := range fileOrder {
			defs := grouped[file]
			if err := config.WriteEnvFile(file, defs, envValues); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", file, err)
				os.Exit(1)
			}
			fmt.Printf("  ✓ %s written\n", file)
		}
	}

	// Config files
	for _, cfg := range m.Config {
		fmt.Printf("\n%s (%s)\n", cfg.Label, cfg.File)
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println()

		fieldValues := make(map[string]string)
		for _, f := range cfg.Fields {
			label := f.Label
			fmt.Printf("  %s\n", label)
			if f.Description != "" {
				fmt.Printf("  %s\n", f.Description)
			}
			if f.Default != "" {
				fmt.Printf("  [default: %s]\n", f.Default)
			}
			fmt.Print("  > ")

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				input = f.Default
			}
			fieldValues[f.Path] = input
			fmt.Println()
		}

		log.Info("Updating %s...", cfg.File)
		if err := config.UpdateConfigFile(cfg.File, fieldValues); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not update %s: %s\n", cfg.File, err)
			log.Warn("Failed to update %s: %s", cfg.File, err)
		} else {
			fmt.Printf("  ✓ %s updated\n", cfg.File)
		}
	}

	fmt.Println("\nConfiguration complete!")
}
