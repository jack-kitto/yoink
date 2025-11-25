package util

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// FetchRawVaultFile attempts to fetch a file from GitHub via HTTPS
// Returns the file content as string, or error if not possible
func FetchRawVaultFile(vaultRepo, branch, file string) (string, error) {
	// Extract org/repo from git URL
	repoPath := extractGitHubRepo(vaultRepo)
	if repoPath == "" {
		return "", fmt.Errorf("invalid GitHub repository URL")
	}

	// Try GitHub Raw API first
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repoPath, branch, file)

	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		data, err := io.ReadAll(resp.Body)
		return string(data), err
	}

	// If that fails, try gh CLI as fallback
	return fetchViaGH(repoPath, branch, file)
}

func fetchViaGH(repoPath, branch, file string) (string, error) {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("/repos/%s/contents/%s", repoPath, file),
		"--jq", ".content | @base64d")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch via gh CLI: %w", err)
	}

	return string(output), nil
}

func extractGitHubRepo(gitURL string) string {
	if strings.Contains(gitURL, "github.com:") {
		// SSH: git@github.com:user/repo.git
		parts := strings.Split(gitURL, ":")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	} else if strings.Contains(gitURL, "github.com/") {
		// HTTPS: https://github.com/user/repo.git
		parts := strings.Split(gitURL, "github.com/")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	}
	return ""
}
