package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckDependencies verifies that required external tools are available
func CheckDependencies() error {
	deps := []string{"sops", "age", "gh", "git"}
	missing := []string{}

	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required tools: %s\n\nPlease install:\n- sops: https://github.com/mozilla/sops\n- age: https://github.com/FiloSottile/age\n- gh: https://cli.github.com\n- git: https://git-scm.com",
			strings.Join(missing, ", "))
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0700)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// WriteGitignore adds entries to .gitignore to prevent committing plaintext secrets
func WriteGitignore(entries []string) error {
	gitignorePath := ".gitignore"
	var content string

	// Read existing .gitignore if it exists
	if FileExists(gitignorePath) {
		data, err := os.ReadFile(gitignorePath)
		if err != nil {
			return err
		}
		content = string(data)
	}

	// Check if we need to add our entries
	needsUpdate := false
	for _, entry := range entries {
		if !strings.Contains(content, entry) {
			content += "\n# Yoink - prevent committing plaintext secrets\n" + entry + "\n"
			needsUpdate = true
			break
		}
	}

	if needsUpdate {
		return os.WriteFile(gitignorePath, []byte(content), 0644)
	}

	return nil
}

// GetGitRepoName returns the name of the current git repository
func GetGitRepoName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return filepath.Base(strings.TrimSpace(string(out))), nil
}

// GetGitRemoteURL returns the remote URL of the current git repository
func GetGitRemoteURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
