package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	SecretsFile string
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "yoink")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func LoadConfig() (Config, error) {
	var c Config
	cfgPath, err := configPath()
	if err != nil {
		return c, err
	}
	viper.SetConfigFile(cfgPath)
	if err := viper.ReadInConfig(); err != nil {
		return c, fmt.Errorf("load config: %w", err)
	}
	c.SecretsFile = viper.GetString("secrets_file")
	if c.SecretsFile == "" {
		c.SecretsFile = filepath.Join(filepath.Dir(cfgPath), "secrets.json")
	}
	return c, nil
}

func InitConfig() error {
	cfgPath, err := configPath()
	if err != nil {
		return err
	}
	v := viper.New()
	v.Set("secrets_file", filepath.Join(filepath.Dir(cfgPath), "secrets.json"))
	if err := v.WriteConfigAs(cfgPath); err != nil {
		return err
	}
	fmt.Println("✅ initialized at", cfgPath)
	return nil
}
