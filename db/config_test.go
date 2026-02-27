package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetActiveProfileDefault(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	profile := GetActiveProfile()
	if profile != "default" {
		t.Errorf("expected 'default', got %q", profile)
	}
}

func TestSetAndGetActiveProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Create the .keys dir
	os.MkdirAll(filepath.Join(tmp, ".keys"), 0700)

	if err := SetActiveProfile("production"); err != nil {
		t.Fatalf("SetActiveProfile: %v", err)
	}

	profile := GetActiveProfile()
	if profile != "production" {
		t.Errorf("expected 'production', got %q", profile)
	}
}

func TestSetActiveProfileOverwrite(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	os.MkdirAll(filepath.Join(tmp, ".keys"), 0700)

	SetActiveProfile("first")
	SetActiveProfile("second")

	profile := GetActiveProfile()
	if profile != "second" {
		t.Errorf("expected 'second', got %q", profile)
	}
}

func TestGetActiveProfileEmptyFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".keys")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "config"), []byte(""), 0600)

	profile := GetActiveProfile()
	if profile != "default" {
		t.Errorf("expected 'default' for empty config, got %q", profile)
	}
}

func TestGetActiveProfileWhitespace(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".keys")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "config"), []byte("  staging  \n"), 0600)

	profile := GetActiveProfile()
	if profile != "staging" {
		t.Errorf("expected 'staging', got %q", profile)
	}
}
