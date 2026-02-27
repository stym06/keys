package db

import (
	"os"
	"path/filepath"
	"strings"
)

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".keys")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config"), nil
}

func GetActiveProfile() string {
	path, err := configPath()
	if err != nil {
		return "default"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "default"
	}
	profile := strings.TrimSpace(string(data))
	if profile == "" {
		return "default"
	}
	return profile
}

func SetActiveProfile(name string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(name+"\n"), 0600)
}
