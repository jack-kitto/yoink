package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jack-kitto/yoink/internal/util"
	"gopkg.in/yaml.v3"
)

type Store struct {
	Path   string
	data   map[string]string
	dryRun bool
}

func New(path string) *Store {
	return &Store{
		Path: path,
		data: make(map[string]string),
	}
}

func NewWithDryRun(path string, dryRun bool) *Store {
	return &Store{
		Path:   path,
		data:   make(map[string]string),
		dryRun: dryRun,
	}
}

func (s *Store) ensureDir() error {
	dir := filepath.Dir(s.Path)
	return util.EnsureDir(dir)
}

func (s *Store) load() error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	// Check if encrypted file exists
	if !util.FileExists(s.Path) {
		s.data = make(map[string]string)
		return nil
	}

	// Check for SOPS config before attempting to decrypt
	dir := filepath.Dir(s.Path)
	if err := CheckSOPSConfig(dir); err != nil {
		return fmt.Errorf("SOPS configuration error: %w", err)
	}

	// Create temporary file for decrypted content
	tmpFile, err := os.CreateTemp("", "yoink-decrypt-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Decrypt the file - don't silently ignore errors
	if err := DecryptWithSOPS(s.Path, tmpFile.Name()); err != nil {
		// Check if the file is actually empty/corrupted
		if fileInfo, statErr := os.Stat(s.Path); statErr == nil && fileInfo.Size() == 0 {
			fmt.Printf("⚠️  Warning: secrets file is empty\n")
			s.data = make(map[string]string)
			return nil
		}
		return fmt.Errorf("failed to decrypt secrets file %s: %w", s.Path, err)
	}

	// Read decrypted content
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return err
	}

	// Handle empty decrypted content
	if len(data) == 0 {
		s.data = make(map[string]string)
		return nil
	}

	// Parse YAML
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to parse decrypted YAML: %w", err)
	}

	// Convert to string map
	s.data = make(map[string]string)
	for k, v := range yamlData {
		s.data[k] = fmt.Sprintf("%v", v)
	}

	return nil
}

func (s *Store) save() error {
	if s.dryRun {
		fmt.Printf("🔍 [DRY RUN] Would save secrets to %s\n", s.Path)
		return nil
	}

	if err := s.ensureDir(); err != nil {
		return err
	}

	// Check for SOPS config
	dir := filepath.Dir(s.Path)
	if err := CheckSOPSConfig(dir); err != nil {
		return fmt.Errorf("SOPS configuration error: %w", err)
	}

	// Convert data to YAML
	yamlData, err := yaml.Marshal(s.data)
	if err != nil {
		return err
	}

	// Create temporary file for plaintext
	tmpFile, err := os.CreateTemp("", "yoink-encrypt-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write plaintext to temp file
	if _, err := tmpFile.Write(yamlData); err != nil {
		return err
	}
	tmpFile.Close()

	// Encrypt and save
	return EncryptWithSOPS(tmpFile.Name(), s.Path)
}

func (s *Store) Load() error {
	return s.load()
}

func (s *Store) Set(key, value string) error {
	if s.dryRun {
		fmt.Printf("🔍 [DRY RUN] Would set %s = %s\n", key, value)
		return nil
	}

	// Load current data first
	if err := s.load(); err != nil {
		return err
	}

	s.data[key] = value
	return s.save()
}

func (s *Store) Get(key string) (string, error) {
	if err := s.load(); err != nil {
		return "", err
	}

	value, exists := s.data[key]
	if !exists {
		return "", fmt.Errorf("secret '%s' not found", key)
	}
	return value, nil
}

func (s *Store) Delete(key string) error {
	if s.dryRun {
		fmt.Printf("🔍 [DRY RUN] Would delete secret: %s\n", key)
		return nil
	}

	if err := s.load(); err != nil {
		return err
	}

	if _, exists := s.data[key]; !exists {
		return fmt.Errorf("secret '%s' not found", key)
	}

	delete(s.data, key)
	return s.save()
}

func (s *Store) Keys() ([]string, error) {
	if err := s.load(); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *Store) All() (map[string]string, error) {
	if err := s.load(); err != nil {
		return nil, err
	}

	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range s.data {
		result[k] = v
	}
	return result, nil
}

func (s *Store) IsEmpty() (bool, error) {
	if err := s.load(); err != nil {
		return false, err
	}
	return len(s.data) == 0, nil
}

// ExportEnv exports secrets in environment variable format
func (s *Store) ExportEnv() (string, error) {
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
