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
	"github.com/jack-kitto/yoink/internal/vault"
)

var (
	cfg          config.Config
	projectCfg   project.ProjectConfig
	secretStore  *store.Store
	configLoaded bool
	projectMode  bool
	dryRun       bool
	verbose      bool
	version      string
	autoPR       bool
)

func Execute(v string) {
	version = v
	root := buildRoot()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// cmd/root.go (updated sections)
func buildRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "yoink",
		Short: "Yoink — a Git-native secret manager with invisible vaults",
		Long:  "Yoink manages encrypted secrets securely with SOPS and uses GitHub as a backend vault, fully automated and invisible to developers.",
	}

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show detailed output including git operations")

	rootCmd.AddCommand(
		initCmd(),
		vaultInitCmd(),
		resetCmd(),
		vaultResetCmd(),
		setCmd(),
		getCmd(),
		debugCmd(),
		deleteCmd(),
		listCmd(),
		exportCmd(),
		runCmd(),
		onboardCmd(),
		removeUserCmd(),
		versionCmd(),
		statusCmd(), // New command
		auditCmd(),  // New command
	)

	return rootCmd
}

// Update commands to use fast store for read operations
func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Retrieve a secret's decrypted value (secure temporary operation)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			// Try fast fetch first
			fs := store.NewFast(projectCfg.VaultRepo)
			val, err := fs.Get(args[0])
			if err == nil {
				fmt.Printf("%s=%s\n", args[0], val)
				return nil
			}

			if verbose {
				fmt.Printf("⚠️  Fast fetch failed (%v), falling back to git clone...\n", err)
			}

			// Fallback to traditional method
			vman, err := vault.New(projectCfg.VaultRepo)
			if err != nil {
				return err
			}
			vman.Verbose = verbose

			defer vman.Cleanup()

			if err := vman.Sync(); err != nil {
				return err
			}

			encPath := filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml")
			s := store.New(encPath)
			val, err = s.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s=%s\n", args[0], val)
			return nil
		},
	}
}

// ensureConfigLoaded loads and prepares vault if possible.
func ensureConfigLoaded() error {
	if configLoaded {
		return nil
	}

	projCfg, err := project.LoadProject()
	if err != nil {
		return fmt.Errorf("run 'yoink vault-init' first in your project: %w", err)
	}
	projectCfg = projCfg
	configLoaded = true
	projectMode = true
	return nil
}

func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Store or update a secret (creates a PR to vault automatically)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}
			key, val := args[0], args[1]
			vman, err := vault.New(projectCfg.VaultRepo)
			if err != nil {
				return err
			}

			vman.Verbose = verbose

			defer func() {
				// Only cleanup on success
				if err == nil {
					vman.Cleanup()
				}
			}()

			if err := vman.Sync(); err != nil {
				return err
			}

			encPath := filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml")
			s := store.New(encPath)
			if err := s.Set(key, val); err != nil {
				return err
			}

			if err := vman.CommitAndPush("secrets.enc.yaml", fmt.Sprintf("update secret %s", key), true); err != nil {
				return fmt.Errorf("failed to commit and create PR: %w", err)
			}

			fmt.Printf("✅ Secret '%s' updated in vault (PR created)\n", key)
			return nil
		},
	}
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a secret entry (creates a PR to remove it)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}
			key := args[0]
			vman, _ := vault.New(projectCfg.VaultRepo)
			_ = vman.Sync()
			s := store.New(filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml"))
			if err := s.Delete(key); err != nil {
				return err
			}
			if err := vman.CommitAndPush("secrets.enc.yaml", fmt.Sprintf("delete secret %s", key), true); err != nil {
				return err
			}
			vman.Cleanup()
			fmt.Printf("✅ Secret '%s' deleted (PR created)\n", key)
			return nil
		},
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize global Yoink configuration (~/.config/yoink)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would initialize global configuration")
				return nil
			}
			err := ensureGlobalConfig()
			if err != nil {
				return fmt.Errorf("failed to initialize global config: %w", err)
			}
			fmt.Println("✅ Global Yoink configuration initialized")
			return nil
		},
	}
}

func resetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset local secrets (delete .yoink/secrets and configuration)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would reset local secrets")

				return nil
			}
			fmt.Println("🧹 Resetting local secrets...")
			if err := os.RemoveAll(".yoink"); err != nil {
				return err
			}
			fmt.Println("✅ Local secrets reset")
			return nil
		},
	}
}

func vaultResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-reset",
		Short: "Force reset and re-clone the remote vault repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			vman, err := vault.New(projectCfg.VaultRepo)
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would reset and re-clone vault repository.")
				return nil
			}

			fmt.Println("🌀 Re-cloning vault repository...")

			// Clean up old repo directory ONLY, not the entire workdir parent
			repoDir := filepath.Join(vman.WorkDir, "repo")
			os.RemoveAll(repoDir)

			if err := vman.Sync(); err != nil {
				return fmt.Errorf("failed to clone vault: %w", err)
			}

			fmt.Println("✅ Vault reset complete")
			return nil
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the current version of Yoink",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Yoink version: %s\n", version)
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all secret keys from the remote vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			// Try fast fetch first
			fs := store.NewFast(projectCfg.VaultRepo)
			keys, err := fs.Keys()
			if err == nil {
				if len(keys) == 0 {
					fmt.Println("(no secrets in vault)")
					return nil
				}
				for _, k := range keys {
					fmt.Println(k)
				}
				return nil
			}

			if verbose {
				fmt.Printf("⚠️  Fast fetch failed (%v), falling back to git clone...\n", err)
			}

			// Fallback to traditional method
			vman, err := vault.New(projectCfg.VaultRepo)
			if err != nil {
				return err
			}
			vman.Verbose = verbose

			defer vman.Cleanup()

			if err := vman.Sync(); err != nil {
				return err
			}

			encPath := filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml")
			s := store.New(encPath)
			keys, err = s.Keys()
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				fmt.Println("(no secrets in vault)")
				return nil
			}
			for _, k := range keys {
				fmt.Println(k)
			}
			return nil
		},
	}
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run -- <command>",
		Short: "Run a command with secrets injected as environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			// Try fast fetch first
			fs := store.NewFast(projectCfg.VaultRepo)
			envMap, err := fs.All()
			if err != nil && verbose {
				fmt.Printf("⚠️  Fast fetch failed (%v), falling back to git clone...\n", err)
			}

			// Fallback to traditional method if fast fetch fails
			if err != nil {
				vman, err := vault.New(projectCfg.VaultRepo)
				if err != nil {
					return fmt.Errorf("failed to initialize vault: %w", err)
				}
				vman.Verbose = verbose

				defer vman.Cleanup()

				if err := vman.Sync(); err != nil {
					return fmt.Errorf("failed to sync vault: %w", err)
				}

				encPath := filepath.Join(vman.WorkDir, "repo", "secrets.enc.yaml")
				s := store.New(encPath)
				envMap, err = s.All()
				if err != nil {
					return fmt.Errorf("failed to load secrets: %w", err)
				}
			}

			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would run command with injected secrets:")
				for k := range envMap {
					fmt.Printf("  %s=***\n", k)
				}
				fmt.Printf("Command: %v\n", args)
				return nil
			}

			// Start with current environment
			env := os.Environ()

			// Add secrets as environment variables
			for k, v := range envMap {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}

			// Execute the command with inherited streams
			c := exec.Command(args[0], args[1:]...)
			c.Env = env
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin

			if !verbose {
				fmt.Printf("🚀 Running command with %d secrets...\n", len(envMap))
			} else {
				fmt.Printf("🚀 Running command with %d injected secrets: %v\n", len(envMap), args)
			}

			return c.Run()
		},
	}

	return cmd
}

func debugCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug vault and secrets information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			fmt.Printf("Project config: %+v\n", projectCfg)

			vman, err := vault.New(projectCfg.VaultRepo)
			if err != nil {
				return err
			}

			if err := vman.Sync(); err != nil {
				return err
			}

			repoDir := filepath.Join(vman.WorkDir, "repo")
			fmt.Printf("Vault repo cloned to: %s\n", repoDir)

			// List files in repo
			fmt.Println("Files in vault repo:")
			filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				relPath, _ := filepath.Rel(repoDir, path)
				fmt.Printf("  %s\n", relPath)
				return nil
			})

			encPath := filepath.Join(repoDir, "secrets.enc.yaml")
			if util.FileExists(encPath) {
				fmt.Printf("Secrets file exists: %s\n", encPath)

				// Try to get file info
				if info, err := os.Stat(encPath); err == nil {
					fmt.Printf("File size: %d bytes\n", info.Size())
				}

				// Try to read raw content (first few lines)
				if data, err := os.ReadFile(encPath); err == nil {
					lines := strings.Split(string(data), "\n")
					fmt.Printf("First few lines of encrypted file:\n")
					for i, line := range lines {
						if i >= 3 {
							break
						}
						fmt.Printf("  %s\n", line)
					}
				}
			} else {
				fmt.Printf("Secrets file does NOT exist at: %s\n", encPath)
			}

			vman.Cleanup()
			return nil
		},
	}
}
