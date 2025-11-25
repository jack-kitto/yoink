package store

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jack-kitto/yoink/internal/config"
)

// setAgeKeyEnv sets the SOPS_AGE_KEY_FILE environment variable to point to yoink's age key
func setAgeKeyEnv() error {
	keyPath, err := config.GetAgeKeyPath()
	if err != nil {
		return fmt.Errorf("failed to get age key path: %w", err)
	}

	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("age key not found at %s (run 'yoink init' to generate it): %w", keyPath, err)
	}

	os.Setenv("SOPS_AGE_KEY_FILE", keyPath)
	return nil
}

// EncryptWithSOPS uses the sops CLI to encrypt a YAML or JSON file
func EncryptWithSOPS(input, output string) error {
	if err := setAgeKeyEnv(); err != nil {
		return err
	}

	cmd := exec.Command("sops", "-e", input)
	data, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("sops encryption failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("sops encryption failed: %w", err)
	}
	return os.WriteFile(output, data, 0o600)
}

// DecryptWithSOPS uses the sops CLI to decrypt a YAML or JSON file
func DecryptWithSOPS(input, output string) error {
	if err := setAgeKeyEnv(); err != nil {
		return err
	}

	cmd := exec.Command("sops", "-d", input)
	data, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("sops decryption failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("sops decryption failed: %w", err)
	}
	return os.WriteFile(output, data, 0o600)
}

// DecryptToString decrypts a SOPS file and returns the content as string
func DecryptToString(input string) (string, error) {
	if err := setAgeKeyEnv(); err != nil {
		return "", err
	}

	cmd := exec.Command("sops", "-d", input)
	data, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("sops decryption failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("sops decryption failed: %w", err)
	}
	return string(data), nil
}

// EncryptString encrypts a string using SOPS and writes to output file
func EncryptString(content string, output string) error {
	if err := setAgeKeyEnv(); err != nil {
		return err
	}

	// Create temporary file with content
	tmpFile, err := os.CreateTemp("", "yoink-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return err
	}
	tmpFile.Close()

	return EncryptWithSOPS(tmpFile.Name(), output)
}

// CheckSOPSConfig verifies that SOPS configuration is available
func CheckSOPSConfig(dir string) error {
	// Check for .sops.yaml starting from the given directory and walking up
	currentDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	for {
		sopsPath := filepath.Join(currentDir, ".sops.yaml")
		if _, err := os.Stat(sopsPath); err == nil {
			return nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root
			break
		}
		currentDir = parent
	}

	return fmt.Errorf(".sops.yaml configuration not found - run 'yoink vault-init' to set up encryption")
}

// InitSOPSForVault initializes SOPS configuration for a vault
func InitSOPSForVault(vaultPath string, ageKeys []string) error {
	sopsConfig := fmt.Sprintf(`creation_rules:
  - path_regex: .*\.(yaml|yml|json)$
    age: %s
`, strings.Join(ageKeys, ","))

	sopsPath := filepath.Join(vaultPath, ".sops.yaml")
	return os.WriteFile(sopsPath, []byte(sopsConfig), 0o644)
}

// InitSOPSForProject initializes SOPS configuration for a project (in project root)
func InitSOPSForProject(projectPath string, ageKeys []string) error {
	sopsConfig := fmt.Sprintf(`creation_rules:
  - path_regex: .*\.(yaml|yml|json)$
    age: %s
`, strings.Join(ageKeys, ","))

	sopsPath := filepath.Join(projectPath, ".sops.yaml")
	return os.WriteFile(sopsPath, []byte(sopsConfig), 0o644)
}
