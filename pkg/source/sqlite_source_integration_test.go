//go:build integration
// +build integration

package source_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/yourorg/envchain/pkg/source"
)

func TestSQLiteSource_Integration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "integration.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE env (key TEXT PRIMARY KEY, value TEXT NOT NULL)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	seeds := map[string]string{
		"INTEGRATION_KEY": "integration_value",
		"ANOTHER_KEY":     "another_value",
	}
	for k, v := range seeds {
		if _, err := db.Exec(`INSERT INTO env (key, value) VALUES (?, ?)`, k, v); err != nil {
			t.Fatalf("seed row %s: %v", k, err)
		}
	}
	db.Close()

	src, err := source.NewSQLiteSource(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteSource: %v", err)
	}

	val, ok := src.Get("INTEGRATION_KEY")
	if !ok {
		t.Fatal("expected INTEGRATION_KEY to be found")
	}
	if val != "integration_value" {
		t.Errorf("got %q, want %q", val, "integration_value")
	}

	keys := src.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d: %v", len(keys), keys)
	}

	_ = os.Remove(dbPath)
}
