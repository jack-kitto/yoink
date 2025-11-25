package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jack-kitto/yoink/internal/store"
	"github.com/jack-kitto/yoink/internal/vault"
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

			// Try fast fetch first
			fs := store.NewFast(projectCfg.VaultRepo)
			all, err := fs.All()
			if err != nil && verbose {
				fmt.Printf("⚠️  Fast fetch failed (%v), falling back to git clone...\n", err)
			}

			// Fallback to traditional method if fast fetch fails
			if err != nil {
				vman, err := vault.New(projectCfg.VaultRepo)
				if err != nil {
					return err
				}
				vman.Verbose = verbose

				defer vman.Cleanup()

				if err := vman.Sync(); err != nil {
					return err
				}

				encPath := filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml")
				s := store.New(encPath)
				all, err = s.All()
				if err != nil {
					return err
				}
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

				return os.WriteFile(envFile, data, 0o600)
			}

			// Export as .env format
			var envOutput string
			if fs != nil {
				envOutput, err = fs.ExportEnv()
			} else {
				// This shouldn't happen given the fallback above, but just in case
				return fmt.Errorf("no store available")
			}

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

			if err := os.WriteFile(envFile, []byte(envOutput), 0o600); err != nil {
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
