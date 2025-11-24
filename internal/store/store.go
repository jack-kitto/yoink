package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Store struct {
	Path string
	data map[string]string
}

func New(path string) *Store {
	s := &Store{Path: path}
	s.load()
	return s
}

func (s *Store) ensureDir() error {
	dir := filepath.Dir(s.Path)
	return os.MkdirAll(dir, 0700)
}

func (s *Store) load() {
	_ = s.ensureDir()
	f, err := os.Open(s.Path)
	if err != nil {
		s.data = map[string]string{}
		return
	}
	defer f.Close()
	_ = json.NewDecoder(f).Decode(&s.data)
	if s.data == nil {
		s.data = map[string]string{}
	}
}

func (s *Store) save() error {
	_ = s.ensureDir()
	f, err := os.Create(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s.data)
}

func (s *Store) Set(key, value string) error {
	s.data[key] = value
	return s.save()
}

func (s *Store) Get(key string) (string, error) {
	v, ok := s.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return v, nil
}

func (s *Store) Keys() ([]string, error) {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}
