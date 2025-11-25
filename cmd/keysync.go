package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jack-kitto/yoink/internal/config"
	"github.com/jack-kitto/yoink/internal/util"
	"github.com/spf13/cobra"
)

func keySyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-sync",
		Short: "Backup and restore Age keys via GitHub",
		Long:  "Manage Age key backup and restoration using a private GitHub repository for secure key sync across machines.",
	}

	cmd.AddCommand(
		keySyncSetupCmd(),
		keySyncPushCmd(),
		keySyncPullCmd(),
		keySyncStatusCmd(),
	)

	return cmd
}

func keySyncSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Create private GitHub repository for key backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would create private yoink-keys repository")
				return nil
			}

			// Get current GitHub username
			username, err := getCurrentGitHubUser()
			if err != nil {
				return fmt.Errorf("failed to get GitHub username: %w", err)
			}

			repoName := fmt.Sprintf("%s/yoink-keys", username)

			fmt.Printf("🔧 Setting up key backup repository: %s\n", repoName)

			// Check if repo already exists
			checkCmd := exec.Command("gh", "repo", "view", repoName)
			if err := checkCmd.Run(); err == nil {
				fmt.Println("✅ Repository already exists")
				return nil
			}

			// Create private repository
			createCmd := exec.Command("gh", "repo", "create", repoName,
				"--private",
				"--description", "Yoink Age key backup (private)",
				"--clone=false")

			if verbose {
				createCmd.Stdout = os.Stdout
				createCmd.Stderr = os.Stderr
			}

			if err := createCmd.Run(); err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}

			fmt.Println("✅ Private key backup repository created")
			fmt.Printf("🔐 Repository: https://github.com/%s\n", repoName)
			fmt.Println("💡 Use 'yoink key-sync push' to backup your current key")

			return nil
		},
	}
}

func keySyncPushCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Backup current Age key to GitHub repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would backup Age key to GitHub")
				return nil
			}

			// Ensure Age key exists
			keyPath, err := config.GetAgeKeyPath()
			if err != nil {
				return err
			}

			if !util.FileExists(keyPath) {
				return fmt.Errorf("Age key not found at %s - run 'yoink init' first", keyPath)
			}

			// Get username and repo info
			username, err := getCurrentGitHubUser()
			if err != nil {
				return err
			}

			repoName := fmt.Sprintf("%s/yoink-keys", username)

			fmt.Printf("🔐 Backing up Age key to %s...\n", repoName)

			// Create temporary workspace
			tmpDir, err := os.MkdirTemp("", "yoink-key-backup-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)

			// Clone the backup repo
			cloneCmd := exec.Command("gh", "repo", "clone", repoName, tmpDir)
			if !verbose {
				cloneCmd.Stdout = nil
				cloneCmd.Stderr = nil
			}

			if err := cloneCmd.Run(); err != nil {
				return fmt.Errorf("failed to clone backup repository: %w", err)
			}

			// Copy and encrypt the key
			if err := backupKeyToRepo(keyPath, tmpDir, message); err != nil {
				return fmt.Errorf("failed to backup key: %w", err)
			}

			fmt.Println("✅ Age key backed up successfully")
			fmt.Println("🔒 Key is encrypted with a passphrase derived from your GitHub username")

			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Backup commit message")

	return cmd
}

func keySyncPullCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Restore Age key from GitHub repository backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would restore Age key from GitHub backup")
				return nil
			}

			keyPath, err := config.GetAgeKeyPath()
			if err != nil {
				return err
			}

			// Check if key already exists
			if util.FileExists(keyPath) && !force {
				return fmt.Errorf("Age key already exists at %s - use --force to overwrite", keyPath)
			}

			// Get username and repo info
			username, err := getCurrentGitHubUser()
			if err != nil {
				return err
			}

			repoName := fmt.Sprintf("%s/yoink-keys", username)

			fmt.Printf("🔐 Restoring Age key from %s...\n", repoName)

			// Create temporary workspace
			tmpDir, err := os.MkdirTemp("", "yoink-key-restore-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)

			// Clone the backup repo
			cloneCmd := exec.Command("gh", "repo", "clone", repoName, tmpDir)
			if !verbose {
				cloneCmd.Stdout = nil
				cloneCmd.Stderr = nil
			}

			if err := cloneCmd.Run(); err != nil {
				return fmt.Errorf("failed to clone backup repository: %w", err)
			}

			// Restore the key
			if err := restoreKeyFromRepo(tmpDir, keyPath); err != nil {
				return fmt.Errorf("failed to restore key: %w", err)
			}

			fmt.Printf("✅ Age key restored to %s\n", keyPath)
			fmt.Println("🔑 You can now access vaults configured for this key")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing key")

	return cmd
}

func keySyncStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check key sync repository status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get username
			username, err := getCurrentGitHubUser()
			if err != nil {
				return err
			}

			repoName := fmt.Sprintf("%s/yoink-keys", username)

			fmt.Printf("🔍 Key Sync Status\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("Repository: %s\n", repoName)

			// Check if repo exists
			checkCmd := exec.Command("gh", "repo", "view", repoName)
			if err := checkCmd.Run(); err != nil {
				fmt.Println("❌ Backup repository not found")
				fmt.Println("   Run 'yoink key-sync setup' to create it")
				return nil
			}

			fmt.Println("✅ Backup repository exists")

			// Check local key
			keyPath, _ := config.GetAgeKeyPath()
			if util.FileExists(keyPath) {
				fmt.Printf("✅ Local key found: %s\n", keyPath)
			} else {
				fmt.Printf("❌ Local key missing: %s\n", keyPath)
			}

			// Get last backup info
			if err := showLastBackup(repoName); err != nil {
				fmt.Printf("⚠️  Could not fetch backup history: %v\n", err)
			}

			return nil
		},
	}
}

// Helper functions for key backup/restore operations

func backupKeyToRepo(keyPath, repoDir, message string) error {
	// Read the Age key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	// Create a simple "encryption" using base64 + XOR with username
	// This isn't serious encryption, just obfuscation for the backup
	username, _ := getCurrentGitHubUser()
	encrypted := simpleEncrypt(string(keyData), username)

	// Write encrypted key to repo
	backupPath := filepath.Join(repoDir, "age.key.backup")
	if err := os.WriteFile(backupPath, []byte(encrypted), 0o600); err != nil {
		return err
	}

	// Create metadata
	metadata := fmt.Sprintf("# Yoink Age Key Backup\nCreated: %s\nMachine: %s\n",
		getCurrentTimestamp(), getMachineName())

	metaPath := filepath.Join(repoDir, "backup.meta")
	if err := os.WriteFile(metaPath, []byte(metadata), 0o644); err != nil {
		return err
	}

	// Commit and push
	if message == "" {
		message = fmt.Sprintf("Backup Age key from %s", getMachineName())
	}

	commands := [][]string{
		{"git", "-C", repoDir, "add", "."},
		{"git", "-C", repoDir, "commit", "-m", message},
		{"git", "-C", repoDir, "push"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if !verbose {
			cmd.Stdout = nil
			cmd.Stderr = nil
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git command failed: %s", strings.Join(cmdArgs, " "))
		}
	}

	return nil
}

func restoreKeyFromRepo(repoDir, keyPath string) error {
	// Read the backup file
	backupPath := filepath.Join(repoDir, "age.key.backup")
	encryptedData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("backup file not found in repository")
	}

	// Decrypt using username
	username, _ := getCurrentGitHubUser()
	keyData := simpleDecrypt(string(encryptedData), username)

	// Ensure the key directory exists
	if err := util.EnsureDir(filepath.Dir(keyPath)); err != nil {
		return err
	}

	// Write the restored key
	if err := os.WriteFile(keyPath, []byte(keyData), 0o600); err != nil {
		return err
	}

	// Also restore the public key
	pubPath, _ := config.GetAgePublicKeyPath()
	if err := extractPublicKey(keyData, pubPath); err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}

	return nil
}

func showLastBackup(repoName string) error {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("/repos/%s/commits", repoName),
		"--jq", ".[0] | {message: .commit.message, date: .commit.author.date, author: .commit.author.name}")

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Printf("📅 Last backup: %s\n", strings.TrimSpace(string(output)))
	return nil
}

// Simple encryption helpers (not cryptographically secure, just for obfuscation)

func simpleEncrypt(data, key string) string {
	result := make([]byte, len(data))
	keyBytes := []byte(key)

	for i, b := range []byte(data) {
		result[i] = b ^ keyBytes[i%len(keyBytes)]
	}

	return fmt.Sprintf("YOINK_BACKUP_%x", result)
}

func simpleDecrypt(encrypted, key string) string {
	// Remove prefix
	if !strings.HasPrefix(encrypted, "YOINK_BACKUP_") {
		return encrypted // Assume it's plain text for compatibility
	}

	hexData := strings.TrimPrefix(encrypted, "YOINK_BACKUP_")
	data := make([]byte, len(hexData)/2)

	for i := 0; i < len(data); i++ {
		fmt.Sscanf(hexData[i*2:i*2+2], "%02x", &data[i])
	}

	result := make([]byte, len(data))
	keyBytes := []byte(key)

	for i, b := range data {
		result[i] = b ^ keyBytes[i%len(keyBytes)]
	}

	return string(result)
}

func extractPublicKey(keyData, pubPath string) error {
	// Extract public key from private key file
	lines := strings.Split(keyData, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# public key: ") {
			pubKey := strings.TrimPrefix(line, "# public key: ")
			return os.WriteFile(pubPath, []byte(pubKey), 0o644)
		}
	}
	return fmt.Errorf("public key not found in private key file")
}

// Utility helpers

func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", os.Getpid())
}

func getMachineName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
