package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/jack-kitto/yoink/internal/util"
)

type ProjectConfig struct {
	VaultRepo   string `yaml:"vault"`
	SecretsPath string `yaml:"secrets_file"`
}

type SOPSConfig struct {
	CreationRules []CreationRule `yaml:"creation_rules"`
}

type CreationRule struct {
	PathRegex string   `yaml:"path_regex"`
	Age       []string `yaml:"age"`
}

func InitProject() error {
	// Check if we're in a git repository
	repoName, err := util.GetGitRepoName()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Create .yoink.yaml
	cfg := ProjectConfig{
		VaultRepo:   fmt.Sprintf("git@github.com:%s-vault.git", repoName),
		SecretsPath: ".yoink/secrets.enc.yaml",
	}

	// Ensure the vault repository exists
	if err := EnsureVaultRepo(&cfg); err != nil {
		fmt.Printf("⚠️  Warning: Could not create vault repo: %v\n", err)
		fmt.Println("You'll need to create the vault repository manually or update .yoink.yaml")
	}

	// Create .yoink.yaml
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(".yoink.yaml", data, 0644); err != nil {
		return err
	}

	// Create .yoink directory
	if err := util.EnsureDir(".yoink"); err != nil {
		return err
	}

	// Add to .gitignore
	gitignoreEntries := []string{
		".yoink/secrets.yaml",
		".yoink/secrets.json",
		"*.dec.yaml",
		"*.dec.json",
	}
	if err := util.WriteGitignore(gitignoreEntries); err != nil {
		fmt.Printf("⚠️  Warning: Could not update .gitignore: %v\n", err)
	}

	fmt.Println("✅ Project vault initialized")
	fmt.Printf("📁 Vault repository: %s\n", cfg.VaultRepo)
	fmt.Printf("🔐 Secrets file: %s\n", cfg.SecretsPath)

	return nil
}

func LoadProject() (ProjectConfig, error) {
	var c ProjectConfig
	
	// Try to find .yoink.yaml in current directory or parent directories
	configPath, err := findProjectConfig()
	if err != nil {
		return c, err
	}
	
	b, err := os.ReadFile(configPath)
	if err != nil {
		return c, fmt.Errorf("not in a yoink project (run 'yoink vault-init' first)")
	}
	
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, err
	}
	
	if c.SecretsPath == "" {
		c.SecretsPath = ".yoink/secrets.enc.yaml"
	}
	
	// Make path relative to config location
	configDir := filepath.Dir(configPath)
	if !filepath.IsAbs(c.SecretsPath) {
		c.SecretsPath = filepath.Join(configDir, c.SecretsPath)
	}
	
	return c, nil
}

func findProjectConfig() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	dir := currentDir
	for {
		configPath := filepath.Join(dir, ".yoink.yaml")
		if util.FileExists(configPath) {
			return configPath, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}
	
	return "", fmt.Errorf(".yoink.yaml not found")
}

func IsVaultRepo(path string) bool {
	return util.FileExists(filepath.Join(path, ".yoink.yaml"))
}

func EnsureVaultRepo(cfg *ProjectConfig) error {
	// Extract repo name from vault URL
	vaultName := extractRepoName(cfg.VaultRepo)
	if vaultName == "" {
		return fmt.Errorf("invalid vault repository URL: %s", cfg.VaultRepo)
	}

	// Check if repo exists
	cmd := exec.Command("gh", "repo", "view", vaultName)
	if err := cmd.Run(); err != nil {
		// Repo doesn't exist, create it
		fmt.Printf("🔨 Creating vault repository: %s\n", vaultName)
		createCmd := exec.Command("gh", "repo", "create", vaultName, "--private", "--description", "Yoink secrets vault")
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create vault repository: %w", err)
		}
	}

	return nil
}

func extractRepoName(gitURL string) string {
	// Handle both SSH and HTTPS URLs
	if strings.Contains(gitURL, "github.com:") {
		// SSH format: git@github.com:user/repo.git
		parts := strings.Split(gitURL, ":")
		if len(parts) >= 2 {
			repoPath := parts[1]
			return strings.TrimSuffix(repoPath, ".git")
		}
	} else if strings.Contains(gitURL, "github.com/") {
		// HTTPS format: https://github.com/user/repo.git
		parts := strings.Split(gitURL, "github.com/")
		if len(parts) >= 2 {
			repoPath := parts[1]
			return strings.TrimSuffix(repoPath, ".git")
		}
	}
	return ""
}

func InitSOPSConfig(vaultPath string, publicKeys []string) error {
	sopsConfig := SOPSConfig{
		CreationRules: []CreationRule{
			{
				PathRegex: ".*\\.(yaml|yml|json)$",
				Age:       publicKeys,
			},
		},
	}

	data, err := yaml.Marshal(sopsConfig)
	if err != nil {
		return err
	}

	sopsPath := filepath.Join(vaultPath, ".sops.yaml")
	return os.WriteFile(sopsPath, data, 0644)
}

func GetVaultDir() (string, error) {
	cfg, err := LoadProject()
	if err != nil {
		return "", err
	}
	
	return filepath.Dir(cfg.SecretsPath), nil
}
