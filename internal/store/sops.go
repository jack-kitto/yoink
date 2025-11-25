package store

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EncryptWithSOPS uses the sops CLI to encrypt a YAML or JSON file
func EncryptWithSOPS(input, output string) error {
	cmd := exec.Command("sops", "-e", input)
	data, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("sops encryption failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("sops encryption failed: %w", err)
	}
	return os.WriteFile(output, data, 0600)
}

// DecryptWithSOPS uses the sops CLI to decrypt a YAML or JSON file
func DecryptWithSOPS(input, output string) error {
	cmd := exec.Command("sops", "-d", input)
	data, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("sops decryption failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("sops decryption failed: %w", err)
	}
	return os.WriteFile(output, data, 0600)
}

// DecryptToString decrypts a SOPS file and returns the content as string
func DecryptToString(input string) (string, error) {
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
	return os.WriteFile(sopsPath, []byte(sopsConfig), 0644)
}

// InitSOPSForProject initializes SOPS configuration for a project (in project root)
func InitSOPSForProject(projectPath string, ageKeys []string) error {
	sopsConfig := fmt.Sprintf(`creation_rules:
  - path_regex: .*\.(yaml|yml|json)$
    age: %s
`, strings.Join(ageKeys, ","))

	sopsPath := filepath.Join(projectPath, ".sops.yaml")
	return os.WriteFile(sopsPath, []byte(sopsConfig), 0644)
}
