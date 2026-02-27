package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
}

func TestAddKeyAndGetAllKeys(t *testing.T) {
	setupTestDB(t)

	if err := AddKey("API_KEY", "abc123"); err != nil {
		t.Fatalf("AddKey: %v", err)
	}
	if err := AddKey("DB_HOST", "localhost"); err != nil {
		t.Fatalf("AddKey: %v", err)
	}

	keys, err := GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	// Should be sorted by name
	if keys[0].Name != "API_KEY" || keys[1].Name != "DB_HOST" {
		t.Errorf("unexpected order: %v, %v", keys[0].Name, keys[1].Name)
	}
	if keys[0].Value != "abc123" {
		t.Errorf("expected value abc123, got %s", keys[0].Value)
	}
}

func TestAddKeyUpsert(t *testing.T) {
	setupTestDB(t)

	if err := AddKey("KEY", "old"); err != nil {
		t.Fatalf("AddKey: %v", err)
	}
	if err := AddKey("KEY", "new"); err != nil {
		t.Fatalf("AddKey: %v", err)
	}

	keys, err := GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key after upsert, got %d", len(keys))
	}
	if keys[0].Value != "new" {
		t.Errorf("expected value 'new', got %s", keys[0].Value)
	}
}

func TestAddKeySetsUpdatedAt(t *testing.T) {
	setupTestDB(t)

	before := time.Now().Unix()
	if err := AddKey("KEY", "val"); err != nil {
		t.Fatalf("AddKey: %v", err)
	}
	after := time.Now().Unix()

	k, err := GetKey("KEY")
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if k.UpdatedAt < before || k.UpdatedAt > after {
		t.Errorf("UpdatedAt %d not in range [%d, %d]", k.UpdatedAt, before, after)
	}
}

func TestGetKeysByNames(t *testing.T) {
	setupTestDB(t)

	AddKey("A", "1")
	AddKey("B", "2")
	AddKey("C", "3")

	keys, err := GetKeysByNames([]string{"A", "C"})
	if err != nil {
		t.Fatalf("GetKeysByNames: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Name != "A" || keys[1].Name != "C" {
		t.Errorf("unexpected keys: %v, %v", keys[0].Name, keys[1].Name)
	}
}

func TestGetKeysByNamesEmpty(t *testing.T) {
	setupTestDB(t)

	keys, err := GetKeysByNames(nil)
	if err != nil {
		t.Fatalf("GetKeysByNames: %v", err)
	}
	if keys != nil {
		t.Errorf("expected nil, got %v", keys)
	}
}

func TestKeyExists(t *testing.T) {
	setupTestDB(t)

	AddKey("EXISTS", "val")

	exists, err := KeyExists("EXISTS")
	if err != nil {
		t.Fatalf("KeyExists: %v", err)
	}
	if !exists {
		t.Error("expected key to exist")
	}

	exists, err = KeyExists("NOPE")
	if err != nil {
		t.Fatalf("KeyExists: %v", err)
	}
	if exists {
		t.Error("expected key to not exist")
	}
}

func TestGetKey(t *testing.T) {
	setupTestDB(t)

	AddKey("MY_KEY", "secret")

	k, err := GetKey("MY_KEY")
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if k.Name != "MY_KEY" || k.Value != "secret" {
		t.Errorf("unexpected key: %+v", k)
	}
}

func TestGetKeyNotFound(t *testing.T) {
	setupTestDB(t)

	_, err := GetKey("MISSING")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestDeleteKey(t *testing.T) {
	setupTestDB(t)

	AddKey("DOOMED", "val")

	if err := DeleteKey("DOOMED"); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}

	exists, _ := KeyExists("DOOMED")
	if exists {
		t.Error("key should be deleted")
	}
}

func TestDeleteKeyNotFound(t *testing.T) {
	setupTestDB(t)

	err := DeleteKey("GHOST")
	if err == nil {
		t.Fatal("expected error deleting nonexistent key")
	}
}

func TestUpdateKey(t *testing.T) {
	setupTestDB(t)

	AddKey("OLD_NAME", "old_val")

	if err := UpdateKey("OLD_NAME", "OLD_NAME", "new_val"); err != nil {
		t.Fatalf("UpdateKey: %v", err)
	}

	k, err := GetKey("OLD_NAME")
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if k.Value != "new_val" {
		t.Errorf("expected new_val, got %s", k.Value)
	}
}

func TestUpdateKeyRename(t *testing.T) {
	setupTestDB(t)

	AddKey("ORIGINAL", "val")

	if err := UpdateKey("ORIGINAL", "RENAMED", "val2"); err != nil {
		t.Fatalf("UpdateKey: %v", err)
	}

	// Old name should be gone
	exists, _ := KeyExists("ORIGINAL")
	if exists {
		t.Error("old key should not exist")
	}

	// New name should exist
	k, err := GetKey("RENAMED")
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if k.Value != "val2" {
		t.Errorf("expected val2, got %s", k.Value)
	}
}

func TestUpdateKeyNotFound(t *testing.T) {
	setupTestDB(t)

	err := UpdateKey("NOPE", "NEW", "val")
	if err == nil {
		t.Fatal("expected error updating nonexistent key")
	}
}

func TestProfileIsolation(t *testing.T) {
	setupTestDB(t)

	// Add key under default profile
	SetActiveProfile("default")
	AddKey("SHARED", "default_val")

	// Switch to dev profile
	SetActiveProfile("dev")
	AddKey("DEV_ONLY", "dev_val")

	// dev profile should only see DEV_ONLY
	keys, _ := GetAllKeys()
	if len(keys) != 1 {
		t.Fatalf("dev profile: expected 1 key, got %d", len(keys))
	}
	if keys[0].Name != "DEV_ONLY" {
		t.Errorf("expected DEV_ONLY, got %s", keys[0].Name)
	}

	// SHARED should not exist in dev
	exists, _ := KeyExists("SHARED")
	if exists {
		t.Error("SHARED should not exist in dev profile")
	}

	// Switch back to default
	SetActiveProfile("default")
	keys, _ = GetAllKeys()
	if len(keys) != 1 {
		t.Fatalf("default profile: expected 1 key, got %d", len(keys))
	}
	if keys[0].Name != "SHARED" {
		t.Errorf("expected SHARED, got %s", keys[0].Name)
	}
}

func TestListProfiles(t *testing.T) {
	setupTestDB(t)

	SetActiveProfile("default")
	AddKey("K1", "v1")

	SetActiveProfile("staging")
	AddKey("K2", "v2")

	SetActiveProfile("prod")
	AddKey("K3", "v3")

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}
	// Should be sorted
	expected := []string{"default", "prod", "staging"}
	for i, p := range profiles {
		if p != expected[i] {
			t.Errorf("profile[%d]: expected %s, got %s", i, expected[i], p)
		}
	}
}

func TestSchemaMigrationCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// DB dir should not exist yet
	dbDir := filepath.Join(tmp, ".keys")
	if _, err := os.Stat(dbDir); !os.IsNotExist(err) {
		t.Fatal("expected .keys dir to not exist before first use")
	}

	// Trigger DB creation
	AddKey("TEST", "val")

	// Dir should now exist
	if _, err := os.Stat(dbDir); err != nil {
		t.Fatalf("expected .keys dir to exist: %v", err)
	}
}

func TestGetAllKeysEmpty(t *testing.T) {
	setupTestDB(t)

	keys, err := GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestDeleteKeyInProfile(t *testing.T) {
	setupTestDB(t)

	SetActiveProfile("default")
	AddKey("KEY", "val1")

	SetActiveProfile("other")
	AddKey("KEY", "val2")

	// Delete from "other" profile
	if err := DeleteKey("KEY"); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}

	// Should still exist in default
	SetActiveProfile("default")
	exists, _ := KeyExists("KEY")
	if !exists {
		t.Error("KEY should still exist in default profile")
	}
}
