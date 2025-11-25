package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jack-kitto/yoink/internal/config"
	"github.com/jack-kitto/yoink/internal/git"
	"github.com/jack-kitto/yoink/internal/project"
	"github.com/jack-kitto/yoink/internal/store"
	"github.com/jack-kitto/yoink/internal/util"
)

var (
	cfg          config.Config
	projectCfg   project.ProjectConfig
	secretStore  *store.Store
	configLoaded bool
	projectMode  bool
	dryRun       bool
	version      string
)

func Execute(v string) {
	version = v
	root := buildRoot()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "yoink",
		Short: "Yoink — a Git-native secret manager",
		Long:  "Yoink manages encrypted secrets using SOPS and GitHub without any backend service.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip dependency check for version, help, and reset commands
			if cmd.Name() == "version" || cmd.Name() == "help" || cmd.Name() == "reset" || cmd.Name() == "vault-reset" {
				return nil
			}

			// Check dependencies for all other commands except init
			if cmd.Name() != "init" {
				if err := util.CheckDependencies(); err != nil {
					return err
				}
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	rootCmd.AddCommand(
		initCmd(),
		setCmd(),
		getCmd(),
		deleteCmd(),
		listCmd(),
		runCmd(),
		exportCmd(),
		vaultInitCmd(),
		onboardCmd(),
		removeUserCmd(),
		resetCmd(),
		vaultResetCmd(),
		versionCmd(),
	)

	return rootCmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("yoink version %s\n", version)
		},
	}
}

func resetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset global yoink configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would reset global configuration")
				return nil
			}

			cfgPath, err := config.GetConfigPath()
			if err != nil {
				return err
			}

			keyPath, err := config.GetAgeKeyPath()
			if err != nil {
				return err
			}

			pubPath, err := config.GetAgePublicKeyPath()
			if err != nil {
				return err
			}

			// Remove config file
			if util.FileExists(cfgPath) {
				os.Remove(cfgPath)
				fmt.Printf("🗑️  Removed config: %s\n", cfgPath)
			}

			// Remove age keys
			if util.FileExists(keyPath) {
				os.Remove(keyPath)
				fmt.Printf("🗑️  Removed age key: %s\n", keyPath)
			}

			if util.FileExists(pubPath) {
				os.Remove(pubPath)
				fmt.Printf("🗑️  Removed public key: %s\n", pubPath)
			}

			fmt.Println("✅ Global configuration reset complete")
			fmt.Println("💡 Run 'yoink init' to set up again")
			return nil
		},
	}
}

func vaultResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-reset",
		Short: "Reset project vault configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would reset project vault configuration")
				return nil
			}

			// Remove .yoink.yaml
			if util.FileExists(".yoink.yaml") {
				os.Remove(".yoink.yaml")
				fmt.Println("🗑️  Removed .yoink.yaml")
			}

			// Remove .yoink directory
			if util.FileExists(".yoink") {
				os.RemoveAll(".yoink")
				fmt.Println("🗑️  Removed .yoink directory")
			}

			// Remove .sops.yaml if it exists
			if util.FileExists(".sops.yaml") {
				os.Remove(".sops.yaml")
				fmt.Println("🗑️  Removed .sops.yaml")
			}

			fmt.Println("✅ Project vault configuration reset complete")
			fmt.Println("💡 Run 'yoink vault-init' to set up vault again")
			return nil
		},
	}
}

// ----------------------------------------------------------------------

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize global yoink configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("🔍 [DRY RUN] Would initialize global configuration")
				return nil
			}

			if err := config.InitConfig(); err != nil {
				// If config already exists, just inform the user
				if os.IsExist(err) {
					fmt.Println("ℹ️  Global configuration already exists")
					return nil
				}
				return err
			}

			fmt.Println("✅ Global configuration initialized successfully.")
			fmt.Println("💡 Run 'yoink vault-init' in your project to set up a vault.")
			return nil
		},
	}
}

// ensureConfigLoaded loads global or project config
func ensureConfigLoaded() error {
	if configLoaded {
		return nil
	}

	// Try to load project config first
	if projCfg, err := project.LoadProject(); err == nil {
		projectCfg = projCfg
		secretStore = store.NewWithDryRun(projectCfg.SecretsPath, dryRun)
		projectMode = true
		configLoaded = true
		return nil
	}

	// Fallback to global config
	conf, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("no configuration found. Run 'yoink init' for global config or 'yoink vault-init' for project config")
	}

	cfg = conf
	secretStore = store.NewWithDryRun(cfg.SecretsFile, dryRun)
	projectMode = false
	configLoaded = true
	return nil
}

// ----------------------------------------------------------------------

func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Store or update a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			key := args[0]
			value := args[1]

			if err := secretStore.Set(key, value); err != nil {
				return err
			}

			if !dryRun {
				// Auto-commit and push if in project mode
				if projectMode {
					commitMsg := fmt.Sprintf("update secret %s", key)
					if err := git.CommitAndPush(secretStore.Path, commitMsg); err != nil {
						fmt.Printf("⚠️  Git commit failed: %v\n", err)
					} else {
						fmt.Printf("📤 Changes committed and pushed\n")
					}
				}

				fmt.Printf("✅ Secret '%s' saved\n", key)
			}
			return nil
		},
	}
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Retrieve a secret value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			value, err := secretStore.Get(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("%s=%s\n", args[0], value)
			return nil
		},
	}
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			key := args[0]
			if err := secretStore.Delete(key); err != nil {
				return err
			}

			if !dryRun {
				// Auto-commit and push if in project mode
				if projectMode {
					commitMsg := fmt.Sprintf("delete secret %s", key)
					if err := git.CommitAndPush(secretStore.Path, commitMsg); err != nil {
						fmt.Printf("⚠️  Git commit failed: %v\n", err)
					} else {
						fmt.Printf("📤 Changes committed and pushed\n")
					}
				}

				fmt.Printf("✅ Secret '%s' deleted\n", key)
			}
			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available secret keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureConfigLoaded(); err != nil {
				return err
			}

			keys, err := secretStore.Keys()
			if err != nil {
				return err
			}

			if len(keys) == 0 {
				fmt.Println("(no secrets stored yet)")
				return nil
			}

			for _, k := range keys {
				fmt.Println(k)
			}
			return nil
		},
	}
}
