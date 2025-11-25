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

type Manager struct {
	RepoURL string
	BaseDir string
	WorkDir string
	Verbose bool
}

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

	return &Manager{
		RepoURL: repoURL,
		BaseDir: base,
		WorkDir: workdir,
		Verbose: false, // Will be set by commands
	}, nil
}

func sanitizeRepoName(repoURL string) string {
	name := strings.ReplaceAll(repoURL, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

// quietRun executes a command with output control based on verbose flag
func (m *Manager) quietRun(dir string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	if m.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func (m *Manager) Sync() error {
	// If repo exists, pull latest; else clone fresh
	dir := filepath.Join(m.WorkDir, "repo")
	if _, err := os.Stat(dir); err == nil {
		if m.Verbose {
			fmt.Println("🔄 Pulling latest changes...")
		}
		if err := m.quietRun(dir, "git", "pull", "--rebase"); err != nil {
			if m.Verbose {
				fmt.Println("⚠️  Could not pull latest, recloning...")
			}
			os.RemoveAll(dir)
		}
	}

	if _, err := os.Stat(filepath.Join(m.WorkDir, "repo")); os.IsNotExist(err) {
		if m.Verbose {
			fmt.Printf("🌀 Cloning vault from %s...\n", m.RepoURL)
		} else {
			fmt.Println("🌀 Syncing vault...")
		}

		if err := m.quietRun(m.WorkDir, "git", "clone", m.RepoURL, "repo"); err != nil {
			return fmt.Errorf("failed to clone vault: %w", err)
		}
	}

	return nil
}

func (m *Manager) CommitAndPush(fileName, msg string, createPR bool) error {
	repoDir := filepath.Join(m.WorkDir, "repo")

	if createPR {
		// Create branch first, before making changes
		branch := "yoink-update-" + time.Now().Format("20060102150405")

		// Make sure we're on main first
		m.quietRun(repoDir, "git", "checkout", "main")
		m.quietRun(repoDir, "git", "checkout", "-b", "main")

		// Create and checkout new branch
		if err := m.quietRun(repoDir, "git", "checkout", "-b", branch); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Now make changes and commit
	if err := m.quietRun(repoDir, "git", "add", fileName); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if err := m.quietRun(repoDir, "git", "commit", "-m", msg); err != nil {
		// Check if it's just "nothing to commit"
		cmd := exec.Command("git", "-C", repoDir, "status", "--porcelain")
		if output, _ := cmd.Output(); len(output) == 0 {
			if m.Verbose {
				fmt.Println("ℹ️  No changes to commit")
			}
			return nil
		}
		return fmt.Errorf("failed to commit: %w", err)
	}

	if createPR {
		branch := "yoink-update-" + time.Now().Format("20060102150405")

		// Push the branch
		if m.Verbose {
			fmt.Printf("📤 Pushing branch %s...\n", branch)
		}

		if err := m.quietRun(repoDir, "git", "push", "origin", branch); err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}

		// Create PR
		if m.Verbose {
			fmt.Println("🔗 Creating pull request...")
		}

		title := fmt.Sprintf("chore(secrets): %s", msg)
		body := fmt.Sprintf("Automated secret update from Yoink.\n\n_Commit message:_ %s", msg)

		cmd := exec.Command("gh", "pr", "create",
			"--title", title,
			"--body", body,
			"--head", branch,
			"--base", "main")
		cmd.Dir = repoDir

		if !m.Verbose {
			cmd.Stdout = nil
			cmd.Stderr = nil
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		fmt.Printf("✅ Pull request created successfully\n")
	} else {
		// Direct push to main
		if m.Verbose {
			fmt.Println("📤 Pushing to main...")
		}
		if err := m.quietRun(repoDir, "git", "push"); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
	}

	return nil
}

func (m *Manager) Cleanup() {
	if util.FileExists(m.WorkDir) {
		os.RemoveAll(m.WorkDir)
	}
}
