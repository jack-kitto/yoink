// cmd/audit.go (update the structs and functions)
package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type CommitInfo struct {
	SHA     string    `json:"sha"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
}

type GitHubUser struct {
	Login string `json:"login"`
}

type PRInfo struct {
	Number int        `json:"number"`
	Title  string     `json:"title"`
	Author GitHubUser `json:"author"`
	URL    string     `json:"url"`
}

func auditCmd() *cobra.Command {
	var asJSON bool
	var limit int
	var short bool

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Show vault history and pending pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			if asJSON {
				return outputAuditJSON(limit)
			}

			return outputAuditHuman(limit, short)
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output in JSON format")
	cmd.Flags().IntVar(&limit, "limit", 10, "Limit number of commits to show")
	cmd.Flags().BoolVar(&short, "short", false, "Show condensed output")

	return cmd
}

func outputAuditJSON(limit int) error {
	commits, err := getRecentCommits(limit)
	if err != nil {
		return err
	}

	prs, err := getPendingPRs()
	if err != nil {
		return err
	}

	output := map[string]interface{}{
		"vault":       projectCfg.VaultRepo,
		"commits":     commits,
		"pending_prs": prs,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func outputAuditHuman(limit int, short bool) error {
	repoName := extractRepoName(projectCfg.VaultRepo)
	fmt.Printf("📋 Vault Audit: %s\n", repoName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Show recent commits
	commits, err := getRecentCommits(limit)
	if err != nil {
		fmt.Printf("⚠️  Could not fetch commit history: %v\n", err)
	} else {
		fmt.Println("\n📝 Recent Activity:")
		for _, commit := range commits {
			if short {
				fmt.Printf("• %s  %s\n",
					commit.Date.Format("01-02"),
					truncate(commit.Message, 50))
			} else {
				fmt.Printf("• %s  %s  by @%s\n",
					commit.Date.Format("2006-01-02"),
					commit.Message,
					commit.Author)
			}
		}
	}

	// Show pending PRs
	prs, err := getPendingPRs()
	if err != nil {
		fmt.Printf("⚠️  Could not fetch pending PRs: %v\n", err)
	} else if len(prs) > 0 {
		fmt.Println("\n🔄 Pending Pull Requests:")
		for _, pr := range prs {
			if short {
				fmt.Printf("• #%d %s\n", pr.Number, truncate(pr.Title, 50))
			} else {
				fmt.Printf("• #%d %s  by @%s\n", pr.Number, pr.Title, pr.Author.Login)
			}
		}
	} else {
		fmt.Println("\n✅ No pending pull requests")
	}

	return nil
}

func getRecentCommits(limit int) ([]CommitInfo, error) {
	repoName := extractRepoName(projectCfg.VaultRepo)

	cmd := exec.Command("gh", "api",
		fmt.Sprintf("/repos/%s/commits", repoName),
		"--jq",
		fmt.Sprintf(".[:%d] | .[] | {sha: .sha, message: .commit.message, author: .commit.author.name, date: .commit.author.date}", limit))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits: %w", err)
	}

	var commits []CommitInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		var commit CommitInfo
		if err := json.Unmarshal([]byte(line), &commit); err == nil {
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

func getPendingPRs() ([]PRInfo, error) {
	repoName := extractRepoName(projectCfg.VaultRepo)

	cmd := exec.Command("gh", "pr", "list",
		"--repo", repoName,
		"--json", "number,title,author,url")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PRs: %w", err)
	}

	var prs []PRInfo
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR JSON: %w", err)
	}

	return prs, nil
}

func extractRepoName(gitURL string) string {
	if strings.Contains(gitURL, "github.com:") {
		parts := strings.Split(gitURL, ":")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	} else if strings.Contains(gitURL, "github.com/") {
		parts := strings.Split(gitURL, "github.com/")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], ".git")
		}
	}
	return gitURL
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
