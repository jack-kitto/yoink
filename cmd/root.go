package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/yourname/yoink/internal/config"
	"github.com/yourname/yoink/internal/store"
)

var cfg config.Config
var secretStore *store.Store

func Execute() {
	rootCmd := buildRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "yoink",
		Short: "Yoink — Yet Another No-backend Keykeeper",
		Long:  "Yoink is a lightweight, Git-friendly secret manager.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cfg = conf
			secretStore = store.New(cfg.SecretsFile)
			return nil
		},
	}

	rootCmd.AddCommand(initCmd(), setCmd(), getCmd(), listCmd())
	return rootCmd
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new yoink configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.InitConfig()
		},
	}
}

func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set or update a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			val := args[1]
			if err := secretStore.Set(key, val); err != nil {
				return err
			}
			fmt.Printf("✅ saved secret %q\n", key)
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
			val, err := secretStore.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s=%s\n", args[0], val)
			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored secret keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := secretStore.Keys()
			if err != nil {
				return err
			}
			for _, k := range keys {
				fmt.Println(k)
			}
			return nil
		},
	}
}
