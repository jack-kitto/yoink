package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func exportCmd() *cobra.Command {
	var envFile string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export decrypted secrets as .env or JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			all, err := secretStore.All()
			if err != nil {
				return err
			}

			if asJSON {
				data, err := json.MarshalIndent(all, "", "  ")
				if err != nil {
					return err
				}

				if envFile == "" {
					fmt.Println(string(data))
					return nil
				}

				if dryRun {
					fmt.Printf("🔍 [DRY RUN] Would write JSON to: %s\n", envFile)
					return nil
				}

				return os.WriteFile(envFile, data, 0600)
			}

			// Export as .env format
			envOutput, err := secretStore.ExportEnv()
			if err != nil {
				return err
			}

			if envFile == "" {
				fmt.Print(envOutput)
				return nil
			}

			if dryRun {
				fmt.Printf("🔍 [DRY RUN] Would write .env to: %s\n", envFile)
				return nil
			}

			if err := os.WriteFile(envFile, []byte(envOutput), 0600); err != nil {
				return err
			}

			fmt.Printf("✅ Secrets exported to %s\n", envFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&envFile, "env-file", "", "write output to .env file")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output in JSON format")
	return cmd
}
