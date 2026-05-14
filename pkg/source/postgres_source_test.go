package source

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockPostgresSource(t *testing.T) (*PostgresSource, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	src := newPostgresSourceWithClient(db, PostgresOptions{})
	return src, mock
}

func TestPostgresSource_GetFound(t *testing.T) {
	src, mock := newMockPostgresSource(t)
	rows := sqlmock.NewRows([]string{"value"}).AddRow("secret123")
	mock.ExpectQuery(`SELECT value FROM env_vars WHERE key`).WithArgs("DB_PASSWORD").WillReturnRows(rows)

	val, ok, err := src.Get(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "secret123" {
		t.Errorf("expected %q, got %q", "secret123", val)
	}
}

func TestPostgresSource_GetMissing(t *testing.T) {
	src, mock := newMockPostgresSource(t)
	mock.ExpectQuery(`SELECT value FROM env_vars WHERE key`).WithArgs("MISSING").WillReturnError(sql.ErrNoRows)

	_, ok, err := src.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestPostgresSource_ClientError(t *testing.T) {
	src, mock := newMockPostgresSource(t)
	mock.ExpectQuery(`SELECT value FROM env_vars WHERE key`).WithArgs("KEY").WillReturnError(errors.New("connection refused"))

	_, _, err := src.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPostgresSource_Keys(t *testing.T) {
	src, mock := newMockPostgresSource(t)
	rows := sqlmock.NewRows([]string{"key"}).AddRow("APP_SECRET").AddRow("DB_HOST")
	mock.ExpectQuery(`SELECT key FROM env_vars`).WillReturnRows(rows)

	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "APP_SECRET" || keys[1] != "DB_HOST" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestPostgresSource_CustomTableAndColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	src := newPostgresSourceWithClient(db, PostgresOptions{
		Table:    "config",
		KeyCol:   "name",
		ValueCol: "val",
	})
	rows := sqlmock.NewRows([]string{"val"}).AddRow("myvalue")
	mock.ExpectQuery(`SELECT val FROM config WHERE name`).WithArgs("MY_KEY").WillReturnRows(rows)

	val, ok, err := src.Get(context.Background(), "MY_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || val != "myvalue" {
		t.Errorf("expected myvalue, got %q (ok=%v)", val, ok)
	}
}
