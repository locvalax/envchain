package source

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestSQLiteDB(t *testing.T, rows map[string]string) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`CREATE TABLE env (key TEXT PRIMARY KEY, value TEXT NOT NULL)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	for k, v := range rows {
		_, err = db.Exec(`INSERT INTO env (key, value) VALUES (?, ?)`, k, v)
		if err != nil {
			t.Fatalf("insert row: %v", err)
		}
	}
	return db
}

func TestSQLiteSource_GetFound(t *testing.T) {
	db := newTestSQLiteDB(t, map[string]string{"DB_HOST": "localhost"})
	s := newSQLiteSourceWithClient(db)

	val, ok := s.Get("DB_HOST")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "localhost" {
		t.Errorf("got %q, want %q", val, "localhost")
	}
}

func TestSQLiteSource_GetMissing(t *testing.T) {
	db := newTestSQLiteDB(t, map[string]string{})
	s := newSQLiteSourceWithClient(db)

	_, ok := s.Get("MISSING_KEY")
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestSQLiteSource_Keys(t *testing.T) {
	db := newTestSQLiteDB(t, map[string]string{
		"APP_ENV":  "production",
		"LOG_LEVEL": "info",
	})
	s := newSQLiteSourceWithClient(db)

	keys := s.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestSQLiteSource_CustomColumns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.db")
	db, _ := sql.Open("sqlite", path)
	t.Cleanup(func() { db.Close() })
	db.Exec(`CREATE TABLE config (name TEXT PRIMARY KEY, val TEXT NOT NULL)`)
	db.Exec(`INSERT INTO config (name, val) VALUES ('TIMEOUT', '30s')`)

	s := newSQLiteSourceWithClient(db,
		WithSQLiteTable("config"),
		WithSQLiteColumns("name", "val"),
	)

	v, ok := s.Get("TIMEOUT")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if v != "30s" {
		t.Errorf("got %q, want %q", v, "30s")
	}
}

func TestNewSQLiteSource_FileNotFound(t *testing.T) {
	// SQLite creates the file on open; use an invalid path to force an error.
	_, err := NewSQLiteSource("/nonexistent-dir-xyz/test.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
	os.Remove("/nonexistent-dir-xyz/test.db")
}
