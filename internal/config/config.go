package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	SecretsFile string `yaml:"secrets_file"`
	DefaultVault string `yaml:"default_vault"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "yoink")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
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
		return c, fmt.Errorf("load config: %w (run 'yoink init' first)", err)
	}
	
	c.SecretsFile = viper.GetString("secrets_file")
	c.DefaultVault = viper.GetString("default_vault")
	
	if c.SecretsFile == "" {
		dir, _ := configDir()
		c.SecretsFile = filepath.Join(dir, "secrets.enc.yaml")
	}
	
	return c, nil
}

func InitConfig() error {
	cfgPath, err := configPath()
	if err != nil {
		return err
	}
	
	// Check if config already exists
	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("config already exists at %s", cfgPath)
	}
	
	dir, _ := configDir()
	v := viper.New()
	v.Set("secrets_file", filepath.Join(dir, "secrets.enc.yaml"))
	v.Set("default_vault", "")
	
	if err := v.WriteConfigAs(cfgPath); err != nil {
		return err
	}
	
	fmt.Printf("✅ Configuration initialized at %s\n", cfgPath)
	return nil
}

func GetAgeKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "yoink", "age.key"), nil
}

func GetAgePublicKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "yoink", "age.pub"), nil
}
