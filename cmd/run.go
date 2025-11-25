package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jack-kitto/yoink/internal/store"
	"github.com/jack-kitto/yoink/internal/vault"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run -- <command>",
		Short: "Run a command with secrets injected as environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Make sure configuration is set up
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			// Clone or refresh the vault
			vman, err := vault.New(projectCfg.VaultRepo)
			if err != nil {
				return fmt.Errorf("failed to initialize vault: %w", err)
			}
			if err := vman.Sync(); err != nil {
				return fmt.Errorf("failed to sync vault: %w", err)
			}

			encPath := filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml")
			s := store.New(encPath)

			// Load all secrets from the vault
			envMap, err := s.All()
			if err != nil {
				vman.Cleanup()
				return fmt.Errorf("failed to load secrets: %w", err)
			}

			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would run command with injected secrets:")
				for k := range envMap {
					fmt.Printf("  %s=***\n", k)
				}
				fmt.Printf("Command: %v\n", args)
				vman.Cleanup()
				return nil
			}

			// Start with current environment
			env := os.Environ()

			// Add secrets as environment variables
			for k, v := range envMap {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}

			// Execute the command with inherited streams
			c := exec.Command(args[0], args[1:]...)
			c.Env = env
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin

			fmt.Printf("🚀 Running command with %d injected secrets...\n", len(envMap))
			err = c.Run()

			vman.Cleanup()
			return err
		},
	}

	return cmd
}
