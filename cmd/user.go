package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/jack-kitto/yoink/internal/config"
	"github.com/jack-kitto/yoink/internal/git"
	"github.com/jack-kitto/yoink/internal/project"
)

func onboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "onboard",
		Short: "Add yourself as a user by creating a GitHub pull request with your Age key",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure we're in a project
			projectCfg, err := project.LoadProject()
			if err != nil {
				return fmt.Errorf("not in a yoink project: %w", err)
			}

			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would create onboarding PR")
				return nil
			}

			// Ensure age key exists
			if err := ensureAgeKey(); err != nil {
				return fmt.Errorf("failed to set up age key: %w", err)
			}

			// Read public key
			pubPath, err := config.GetAgePublicKeyPath()
			if err != nil {
				return err
			}

			pubKeyData, err := os.ReadFile(pubPath)
			if err != nil {
				return err
			}

			publicKey := strings.TrimSpace(string(pubKeyData))

			// Get current user info
			username, err := getCurrentGitHubUser()
			if err != nil {
				return fmt.Errorf("failed to get GitHub username: %w", err)
			}

			// Create PR
			title := fmt.Sprintf("Add %s to vault access", username)
			body := fmt.Sprintf("Adding new user to vault access.\n\n"+
				"**User:** %s\n"+
				"**Public Key:**\n"+
				"```\n"+
				"%s\n"+
				"```\n\n"+
				"Please add this public key to the .sops.yaml configuration to grant vault access.", username, publicKey)

			fmt.Println("📤 Creating GitHub pull request...")

			if err := git.ForkAndCreatePR(projectCfg.VaultRepo, title, body); err != nil {
				return fmt.Errorf("failed to create PR: %w", err)
			}

			fmt.Println("✅ Pull request created!")
			fmt.Println("💭 Maintainers must merge the PR to grant you access to the vault.")

			return nil
		},
	}
}

func removeUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-user <public-key>",
		Short: "Create a PR to remove a user's access to the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure we're in a project
			projectCfg, err := project.LoadProject()
			if err != nil {
				return fmt.Errorf("not in a yoink project: %w", err)
			}

			if dryRun {
				fmt.Printf("🔍 [DRY RUN] Would create PR to remove user: %s\n", args[0])
				return nil
			}

			publicKey := strings.TrimSpace(args[0])

			// Get current user info
			username, err := getCurrentGitHubUser()
			if err != nil {
				return fmt.Errorf("failed to get GitHub username: %w", err)
			}

			// Create PR
			title := "Remove user from vault access"
			body := fmt.Sprintf("Removing user from vault access.\n\n"+
				"**Requested by:** %s\n"+
				"**Public Key to Remove:**\n"+
				"```\n"+
				"%s\n"+
				"```\n\n"+
				"Please remove this public key from the .sops.yaml configuration.", username, publicKey)

			fmt.Println("📤 Creating GitHub pull request...")

			if err := git.ForkAndCreatePR(projectCfg.VaultRepo, title, body); err != nil {
				return fmt.Errorf("failed to create PR: %w", err)
			}

			fmt.Println("✅ Pull request created!")
			fmt.Println("💭 Maintainers must merge the PR to remove user access.")

			return nil
		},
	}
}

func getCurrentGitHubUser() (string, error) {
	// Try to get username from gh CLI
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("please authenticate with 'gh auth login' first")
	}

	username := strings.TrimSpace(string(output))
	if username == "" {
		return "", fmt.Errorf("could not determine GitHub username")
	}

	return username, nil
}
