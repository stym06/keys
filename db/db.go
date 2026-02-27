package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Key struct {
	Name      string
	Value     string
	UpdatedAt int64
}

func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".keys")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "keys.db"), nil
}

func columnExists(d *sql.DB, table, column string) bool {
	rows, err := d.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

func open() (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}
	d, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Create table if not exists (original schema)
	_, err = d.Exec(`CREATE TABLE IF NOT EXISTS keys (
		name TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`)
	if err != nil {
		d.Close()
		return nil, err
	}

	// Migrate: add updated_at column
	if !columnExists(d, "keys", "updated_at") {
		now := time.Now().Unix()
		_, err = d.Exec(`ALTER TABLE keys ADD COLUMN updated_at INTEGER`)
		if err != nil {
			d.Close()
			return nil, err
		}
		_, _ = d.Exec(`UPDATE keys SET updated_at = ? WHERE updated_at IS NULL`, now)
	}

	// Migrate: add profile column and recreate table with composite PK
	if !columnExists(d, "keys", "profile") {
		tx, err := d.Begin()
		if err != nil {
			d.Close()
			return nil, err
		}
		now := time.Now().Unix()
		stmts := []string{
			`ALTER TABLE keys RENAME TO keys_old`,
			`CREATE TABLE keys (
				profile TEXT NOT NULL DEFAULT 'default',
				name TEXT NOT NULL,
				value TEXT NOT NULL,
				updated_at INTEGER,
				PRIMARY KEY (profile, name)
			)`,
			fmt.Sprintf(`INSERT INTO keys (profile, name, value, updated_at)
				SELECT 'default', name, value, COALESCE(updated_at, %d) FROM keys_old`, now),
			`DROP TABLE keys_old`,
		}
		for _, s := range stmts {
			if _, err := tx.Exec(s); err != nil {
				tx.Rollback()
				d.Close()
				return nil, err
			}
		}
		if err := tx.Commit(); err != nil {
			d.Close()
			return nil, err
		}
	}

	return d, nil
}

func AddKey(name, value string) error {
	d, err := open()
	if err != nil {
		return err
	}
	defer d.Close()

	profile := GetActiveProfile()
	now := time.Now().Unix()
	_, err = d.Exec(
		`INSERT INTO keys (profile, name, value, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(profile, name) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		profile, name, value, now,
	)
	return err
}

func GetAllKeys() ([]Key, error) {
	d, err := open()
	if err != nil {
		return nil, err
	}
	defer d.Close()

	profile := GetActiveProfile()
	rows, err := d.Query(`SELECT name, value, COALESCE(updated_at, 0) FROM keys WHERE profile = ? ORDER BY name`, profile)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []Key
	for rows.Next() {
		var k Key
		if err := rows.Scan(&k.Name, &k.Value, &k.UpdatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func GetKeysByNames(names []string) ([]Key, error) {
	if len(names) == 0 {
		return nil, nil
	}
	d, err := open()
	if err != nil {
		return nil, err
	}
	defer d.Close()

	profile := GetActiveProfile()
	query := `SELECT name, value, COALESCE(updated_at, 0) FROM keys WHERE profile = ? AND name IN (`
	args := []interface{}{profile}
	for i, n := range names {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, n)
	}
	query += `) ORDER BY name`

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []Key
	for rows.Next() {
		var k Key
		if err := rows.Scan(&k.Name, &k.Value, &k.UpdatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func KeyExists(name string) (bool, error) {
	d, err := open()
	if err != nil {
		return false, err
	}
	defer d.Close()

	profile := GetActiveProfile()
	var count int
	err = d.QueryRow(`SELECT COUNT(*) FROM keys WHERE profile = ? AND name = ?`, profile, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetKey(name string) (*Key, error) {
	d, err := open()
	if err != nil {
		return nil, err
	}
	defer d.Close()

	profile := GetActiveProfile()
	var k Key
	err = d.QueryRow(`SELECT name, value, COALESCE(updated_at, 0) FROM keys WHERE profile = ? AND name = ?`, profile, name).Scan(&k.Name, &k.Value, &k.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("key %q not found", name)
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func DeleteKey(name string) error {
	d, err := open()
	if err != nil {
		return err
	}
	defer d.Close()

	profile := GetActiveProfile()
	res, err := d.Exec(`DELETE FROM keys WHERE profile = ? AND name = ?`, profile, name)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("key %q not found", name)
	}
	return nil
}

func UpdateKey(oldName, newName, newValue string) error {
	d, err := open()
	if err != nil {
		return err
	}
	defer d.Close()

	profile := GetActiveProfile()
	now := time.Now().Unix()

	tx, err := d.Begin()
	if err != nil {
		return err
	}

	// Delete old key
	res, err := tx.Exec(`DELETE FROM keys WHERE profile = ? AND name = ?`, profile, oldName)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if n == 0 {
		tx.Rollback()
		return fmt.Errorf("key %q not found", oldName)
	}

	// Insert new key
	_, err = tx.Exec(
		`INSERT INTO keys (profile, name, value, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(profile, name) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		profile, newName, newValue, now,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func NukeKeys() (int64, error) {
	d, err := open()
	if err != nil {
		return 0, err
	}
	defer d.Close()

	profile := GetActiveProfile()
	res, err := d.Exec(`DELETE FROM keys WHERE profile = ?`, profile)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func ListProfiles() ([]string, error) {
	d, err := open()
	if err != nil {
		return nil, err
	}
	defer d.Close()

	rows, err := d.Query(`SELECT DISTINCT profile FROM keys ORDER BY profile`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}
