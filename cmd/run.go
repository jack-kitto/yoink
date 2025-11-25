package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run -- <command>",
		Short: "Run a command with secrets injected as environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would run command with injected secrets:")
				envMap, err := secretStore.All()
				if err != nil {
					return err
				}
				for k := range envMap {
					fmt.Printf("  %s=***\n", k)
				}
				fmt.Printf("Command: %v\n", args)
				return nil
			}

			envMap, err := secretStore.All()
			if err != nil {
				return err
			}

			// Start with current environment
			env := os.Environ()
			
			// Add secrets as environment variables
			for k, v := range envMap {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}

			// Execute the command
			c := exec.Command(args[0], args[1:]...)
			c.Env = env
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin

			fmt.Printf("🚀 Running command with %d injected secrets...\n", len(envMap))
			return c.Run()
		},
	}
	return cmd
}
