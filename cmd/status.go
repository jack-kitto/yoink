package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jack-kitto/yoink/internal/config"
	"github.com/jack-kitto/yoink/internal/project"
	"github.com/jack-kitto/yoink/internal/store"
	"github.com/jack-kitto/yoink/internal/util"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show diagnostic information about Yoink configuration and dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Yoink Status Check")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━")

			// Check dependencies
			if err := checkDependencies(); err != nil {
				fmt.Printf("❌ Dependencies: %v\n", err)
			} else {
				fmt.Println("✅ Dependencies: sops, age, git, gh")
			}

			// Check global config
			if _, err := config.LoadConfig(); err != nil {
				fmt.Printf("❌ Global config: %v\n", err)
				fmt.Println("   Run 'yoink init' to set up global configuration")
			} else {
				fmt.Println("✅ Global config loaded")
			}

			// Check Age key
			keyPath, _ := config.GetAgeKeyPath()
			if util.FileExists(keyPath) {
				fmt.Printf("✅ Age key found: %s\n", keyPath)
			} else {
				fmt.Printf("❌ Age key missing: %s\n", keyPath)
				fmt.Println("   Run 'yoink init' to generate Age key")
			}

			// Check project config
			if projectCfg, err := project.LoadProject(); err != nil {
				fmt.Printf("⚠️  Project config: %v\n", err)
				fmt.Println("   Run 'yoink vault-init' in a project directory")
			} else {
				fmt.Println("✅ Project config loaded")
				fmt.Printf("   Vault: %s\n", projectCfg.VaultRepo)

				// Test vault accessibility
				if err := testVaultAccess(projectCfg.VaultRepo); err != nil {
					fmt.Printf("❌ Vault access: %v\n", err)
				} else {
					fmt.Println("✅ Vault reachable")
				}

				// Test decryption capability
				if err := testDecryption(projectCfg.VaultRepo); err != nil {
					fmt.Printf("❌ Decryption test: %v\n", err)
				} else {
					fmt.Println("✅ Able to decrypt secrets")
				}
			}

			// Check GitHub auth status (only if gh is available)
			if _, err := exec.LookPath("gh"); err != nil {
				fmt.Println("⚠️  GitHub CLI (gh) not available")
			} else {
				if err := checkGitHubAuth(); err != nil {
					fmt.Printf("⚠️  GitHub auth: %v\n", err)
				} else {
					fmt.Println("✅ GitHub authenticated")
				}
			}

			return nil
		},
	}
}

func checkDependencies() error {
	return util.CheckDependencies()
}

func testVaultAccess(vaultRepo string) error {
	// Try to fetch a file (even if it doesn't exist, we'll get a proper HTTP response)
	_, err := util.FetchRawVaultFile(vaultRepo, "main", ".sops.yaml")
	if err != nil {
		// Fallback to git ls-remote
		cmd := exec.Command("git", "ls-remote", vaultRepo)
		return cmd.Run()
	}
	return nil
}

func testDecryption(vaultRepo string) error {
	// Try to fetch and decrypt secrets file
	fs := store.NewFast(vaultRepo)
	_, err := fs.Keys()
	return err
}

func checkGitHubAuth() error {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()

	// gh auth status returns exit code 1 even when authenticated
	// Check the output content instead
	outputStr := string(output)
	if strings.Contains(outputStr, "Logged in to github.com") ||
		strings.Contains(outputStr, "✓ Logged in to github.com") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("not authenticated - run 'gh auth login'")
	}

	return fmt.Errorf("authentication status unclear")
}
