package store

import (
	"fmt"
	"os"
	"strings"

	"github.com/jack-kitto/yoink/internal/util"
	"gopkg.in/yaml.v3"
)

// FastStore provides quick access to secrets via HTTPS fetch
type FastStore struct {
	VaultRepo string
	Branch    string
	File      string
	data      map[string]string
}

// NewFast creates a store that fetches via HTTPS when possible
func NewFast(vaultRepo string) *FastStore {
	return &FastStore{
		VaultRepo: vaultRepo,
		Branch:    "main",
		File:      "secrets.enc.yaml",
		data:      make(map[string]string),
	}
}

func (s *FastStore) loadFast() error {
	// Try fast HTTPS fetch first
	content, err := util.FetchRawVaultFile(s.VaultRepo, s.Branch, s.File)
	if err != nil {
		return fmt.Errorf("fast fetch failed: %w", err)
	}

	// Create temporary file for SOPS decryption
	tmpFile, err := os.CreateTemp("", "yoink-fast-*.enc.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write encrypted content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		return err
	}
	tmpFile.Close()

	// Decrypt using existing SOPS logic
	decrypted, err := DecryptToString(tmpFile.Name())
	if err != nil {
		return err
	}

	// Parse YAML
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal([]byte(decrypted), &yamlData); err != nil {
		return fmt.Errorf("failed to parse decrypted YAML: %w", err)
	}

	// Convert to string map
	s.data = make(map[string]string)
	for k, v := range yamlData {
		s.data[k] = fmt.Sprintf("%v", v)
	}

	return nil
}

func (s *FastStore) Get(key string) (string, error) {
	if err := s.loadFast(); err != nil {
		return "", err
	}

	value, exists := s.data[key]
	if !exists {
		return "", fmt.Errorf("secret '%s' not found", key)
	}
	return value, nil
}

func (s *FastStore) All() (map[string]string, error) {
	if err := s.loadFast(); err != nil {
		return nil, err
	}

	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range s.data {
		result[k] = v
	}
	return result, nil
}

func (s *FastStore) Keys() ([]string, error) {
	if err := s.loadFast(); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *FastStore) ExportEnv() (string, error) {
	data, err := s.All()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for k, v := range data {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	return sb.String(), nil
}
