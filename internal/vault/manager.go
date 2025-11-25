package vault

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jack-kitto/yoink/internal/util"
)

// Manager handles cloning, committing, and cleaning up vault repositories
// within ~/.config/yoink/vaults/<repoName>.
type Manager struct {
	RepoURL string
	BaseDir string
	WorkDir string
}

// New creates or reuses a local workspace for the given vault.
func New(repoURL string) (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	base := filepath.Join(home, ".config", "yoink", "vaults")
	workdir := filepath.Join(base, sanitizeRepoName(repoURL))

	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return nil, err
	}

	return &Manager{RepoURL: repoURL, BaseDir: base, WorkDir: workdir}, nil
}

func sanitizeRepoName(repoURL string) string {
	name := strings.ReplaceAll(repoURL, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

// Git helper
func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) Sync() error {
	// If repo exists, pull latest; else clone fresh
	dir := filepath.Join(m.WorkDir, "repo")
	if _, err := os.Stat(dir); err == nil {
		cmd := exec.Command("git", "-C", dir, "pull", "--rebase")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("⚠️  Could not pull latest, recloning...")
			os.RemoveAll(dir)
		}
	}

	if _, err := os.Stat(filepath.Join(m.WorkDir, "repo")); os.IsNotExist(err) {
		fmt.Println("🌀 Cloning vault...")
		cmd := exec.Command("git", "clone", m.RepoURL, "repo")
		cmd.Dir = m.WorkDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone vault: %w", err)
		}
	}

	return nil
}

// CommitAndPush commits the given file to the vault and optionally opens a PR.
func (m *Manager) CommitAndPush(fileName, msg string, createPR bool) error {
	repoDir := filepath.Join(m.WorkDir, "repo")

	if createPR {
		// Create branch first, before making changes
		branch := "yoink-update-" + time.Now().Format("20060102150405")

		// Make sure we're on main first
		cmd := exec.Command("git", "-C", repoDir, "checkout", "main")
		if err := cmd.Run(); err != nil {
			// If main doesn't exist, we might be on master or the repo might be empty
			exec.Command("git", "-C", repoDir, "checkout", "-b", "main").Run()
		}

		// Create and checkout new branch
		cmd = exec.Command("git", "-C", repoDir, "checkout", "-b", branch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Now make changes and commit
	cmds := [][]string{
		{"git", "-C", repoDir, "add", fileName},
		{"git", "-C", repoDir, "commit", "-m", msg},
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Only fail on actual errors, not "no changes to commit"
			if !strings.Contains(string(output), "nothing to commit") {
				return fmt.Errorf("git command failed: %s\nOutput: %s", strings.Join(c, " "), string(output))
			}
		}
	}

	if createPR {
		branch := "yoink-update-" + time.Now().Format("20060102150405")

		// Push the branch
		cmd := exec.Command("git", "-C", repoDir, "push", "origin", branch)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to push branch: %w\nOutput: %s", err, string(output))
		}

		// Create PR - run from the repo directory
		title := fmt.Sprintf("chore(secrets): %s", msg)
		body := fmt.Sprintf("Automated secret update from Yoink.\n\n_Commit message:_ %s", msg)

		cmd = exec.Command("gh", "pr", "create",
			"--title", title,
			"--body", body,
			"--head", branch,
			"--base", "main")
		cmd.Dir = repoDir

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create PR: %w\nOutput: %s", err, string(output))
		}

		fmt.Printf("🔗 Pull request created successfully\n")
	} else {
		// Direct push to main
		cmd := exec.Command("git", "-C", repoDir, "push")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to push: %w\nOutput: %s", err, string(output))
		}
	}

	return nil
}

func (m *Manager) Cleanup() {
	if util.FileExists(m.WorkDir) {
		os.RemoveAll(m.WorkDir)
	}
}

func extractRepoFromURL(repoURL string) string {
	if strings.Contains(repoURL, "github.com:") {
		return strings.TrimSuffix(strings.Split(repoURL, ":")[1], ".git")
	}
	if strings.Contains(repoURL, "github.com/") {
		return strings.TrimSuffix(strings.Split(repoURL, "github.com/")[1], ".git")
	}
	return repoURL
}
