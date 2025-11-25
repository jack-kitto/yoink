package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jack-kitto/yoink/internal/util"
)

func CommitAndPush(path string, message string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%s not found", path)
	}

	// Get the directory containing the file
	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	cmds := [][]string{
		{"git", "-C", dir, "add", filename},
		{"git", "-C", dir, "commit", "-m", message},
		{"git", "-C", dir, "push"},
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git command failed: %s\nOutput: %s", strings.Join(c, " "), string(output))
		}
	}
	return nil
}

func CommitVault(vaultURL, secretsFile, msg string, dryRun bool) error {
	if dryRun {
		fmt.Printf("🔍 [DRY RUN] Would commit and push to vault: %s\n", vaultURL)
		return nil
	}

	dir, err := os.MkdirTemp("", "yoink-vault-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	// Clone the vault repository
	cloneCmd := exec.Command("git", "clone", vaultURL, dir)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone vault repository: %w", err)
	}

	// Copy secrets file to vault
	dest := filepath.Join(dir, filepath.Base(secretsFile))
	if err := util.CopyFile(secretsFile, dest); err != nil {
		return err
	}

	// Commit and push
	cmds := [][]string{
		{"git", "-C", dir, "add", "."},
		{"git", "-C", dir, "commit", "-m", msg},
		{"git", "-C", dir, "push"},
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git command failed: %s", strings.Join(c, " "))
		}
	}

	return nil
}

func CreatePR(repoURL, title, body, branch string) error {
	// Extract repository name from URL
	repoName := extractRepoFromURL(repoURL)
	if repoName == "" {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	// Create PR using gh CLI
	cmd := exec.Command("gh", "pr", "create",
		"--repo", repoName,
		"--title", title,
		"--body", body,
		"--head", branch,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	return nil
}

func ForkAndCreatePR(repoURL, title, body string) error {
	// Extract repository name from URL
	repoName := extractRepoFromURL(repoURL)
	if repoName == "" {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	// Fork the repository
	forkCmd := exec.Command("gh", "repo", "fork", repoName, "--clone=false")
	if err := forkCmd.Run(); err != nil {
		// Fork might already exist, continue
		fmt.Printf("⚠️  Fork might already exist: %v\n", err)
	}

	// Create PR
	prCmd := exec.Command("gh", "pr", "create",
		"--repo", repoName,
		"--title", title,
		"--body", body,
	)

	output, err := prCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create PR: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("✅ Pull request created: %s\n", string(output))
	return nil
}

func CheckRepoExists(repoURL string) error {
	repoName := extractRepoFromURL(repoURL)
	if repoName == "" {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	cmd := exec.Command("gh", "repo", "view", repoName)
	return cmd.Run()
}

func CreateRepo(repoURL, description string, private bool) error {
	repoName := extractRepoFromURL(repoURL)
	if repoName == "" {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	args := []string{"repo", "create", repoName}
	if description != "" {
		args = append(args, "--description", description)
	}
	if private {
		args = append(args, "--private")
	}

	cmd := exec.Command("gh", args...)
	return cmd.Run()
}

func extractRepoFromURL(repoURL string) string {
	// Handle SSH and HTTPS URLs
	if strings.Contains(repoURL, "github.com:") {
		// SSH: git@github.com:user/repo.git
		parts := strings.Split(repoURL, ":")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	} else if strings.Contains(repoURL, "github.com/") {
		// HTTPS: https://github.com/user/repo.git
		parts := strings.Split(repoURL, "github.com/")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	}
	return ""
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}
