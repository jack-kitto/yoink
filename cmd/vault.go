package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/jack-kitto/yoink/internal/config"
	"github.com/jack-kitto/yoink/internal/project"
	"github.com/jack-kitto/yoink/internal/store"
	"github.com/jack-kitto/yoink/internal/util"
)

func vaultInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-init",
		Short: "Initialize a per-project vault configuration (.yoink.yaml)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would initialize project vault")
				return nil
			}

			// Ensure we're in a git repository
			if !util.FileExists(".git") {
				return fmt.Errorf("not in a git repository - run 'git init' first")
			}

			// Check if already initialized
			if util.FileExists(".yoink.yaml") {
				return fmt.Errorf("vault already initialized (.yoink.yaml exists)")
			}

			// Generate age key if it doesn't exist
			if err := ensureAgeKey(); err != nil {
				return fmt.Errorf("failed to set up age key: %w", err)
			}

			// Initialize the project
			if err := project.InitProject(); err != nil {
				return err
			}

			// Initialize SOPS configuration
			if err := initSOPSForProject(); err != nil {
				return fmt.Errorf("failed to initialize SOPS: %w", err)
			}

			return nil
		},
	}
}

func ensureAgeKey() error {
	keyPath, err := config.GetAgeKeyPath()
	if err != nil {
		return err
	}

	pubPath, err := config.GetAgePublicKeyPath()
	if err != nil {
		return err
	}

	// Check if key already exists
	if util.FileExists(keyPath) {
		fmt.Println("🔑 Using existing Age key")
		return nil
	}

	fmt.Println("🔑 Generating new Age key pair...")

	// Ensure directory exists
	keyDir := filepath.Dir(keyPath)
	if err := util.EnsureDir(keyDir); err != nil {
		return err
	}

	// Generate key using age-keygen
	cmd := exec.Command("age-keygen", "-o", keyPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("age-keygen failed: %w\nOutput: %s", err, string(output))
	}

	// Extract public key from the private key file
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	// Find the public key in the comment
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# public key: ") {
			pubKey := strings.TrimPrefix(line, "# public key: ")
			if err := os.WriteFile(pubPath, []byte(pubKey), 0644); err != nil {
				return err
			}
			break
		}
	}

	fmt.Printf("✅ Age key generated: %s\n", keyPath)
	return nil
}

func initSOPSForProject() error {
	// Read the public key
	pubPath, err := config.GetAgePublicKeyPath()
	if err != nil {
		return err
	}

	pubKeyData, err := os.ReadFile(pubPath)
	if err != nil {
		return err
	}

	publicKey := strings.TrimSpace(string(pubKeyData))

	// Create .sops.yaml in the vault directory
	vaultDir, err := project.GetVaultDir()
	if err != nil {
		return err
	}

	if err := util.EnsureDir(vaultDir); err != nil {
		return err
	}

	return store.InitSOPSForVault(vaultDir, []string{publicKey})
}
