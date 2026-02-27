package cmd

import (
	"bytes"
	"testing"

	"keys/db"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
}

func TestGetDirectLookup(t *testing.T) {
	setupTestEnv(t)
	db.AddKey("MY_KEY", "my_secret_value")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"get", "MY_KEY"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "my_secret_value\n" {
		t.Errorf("expected 'my_secret_value\\n', got %q", out)
	}
}

func TestGetNotFound(t *testing.T) {
	setupTestEnv(t)

	rootCmd.SetArgs([]string{"get", "NONEXISTENT"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestGetMultipleKeys(t *testing.T) {
	setupTestEnv(t)
	db.AddKey("FIRST", "val1")
	db.AddKey("SECOND", "val2")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"get", "SECOND"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "val2\n" {
		t.Errorf("expected 'val2\\n', got %q", out)
	}
}

func TestGetRespectsProfile(t *testing.T) {
	setupTestEnv(t)

	db.SetActiveProfile("default")
	db.AddKey("SHARED", "default_val")

	db.SetActiveProfile("dev")
	db.AddKey("SHARED", "dev_val")

	// Get from dev profile
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"get", "SHARED"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "dev_val\n" {
		t.Errorf("expected 'dev_val\\n', got %q", out)
	}

	// Switch to default and get again
	db.SetActiveProfile("default")
	buf.Reset()
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"get", "SHARED"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out = buf.String()
	if out != "default_val\n" {
		t.Errorf("expected 'default_val\\n', got %q", out)
	}
}

func TestGetTooManyArgs(t *testing.T) {
	setupTestEnv(t)

	rootCmd.SetArgs([]string{"get", "A", "B"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error with too many args")
	}
}
