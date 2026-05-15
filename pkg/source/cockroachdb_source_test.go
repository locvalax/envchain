package source

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockCockroachDBSource(t *testing.T) (*CockroachDBSource, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return newCockroachDBSourceWithClient(db, "env_vars", ""), mock
}

func TestCockroachDBSource_GetFound(t *testing.T) {
	src, mock := newMockCockroachDBSource(t)
	mock.ExpectQuery("SELECT value FROM env_vars").
		WithArgs("DB_HOST").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("localhost"))

	val, ok, err := src.Get(context.Background(), "DB_HOST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "localhost" {
		t.Errorf("got %q, want %q", val, "localhost")
	}
}

func TestCockroachDBSource_GetMissing(t *testing.T) {
	src, mock := newMockCockroachDBSource(t)
	mock.ExpectQuery("SELECT value FROM env_vars").
		WithArgs("MISSING").
		WillReturnError(sql.ErrNoRows)

	_, ok, err := src.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestCockroachDBSource_ClientError(t *testing.T) {
	src, mock := newMockCockroachDBSource(t)
	mock.ExpectQuery("SELECT value FROM env_vars").
		WithArgs("KEY").
		WillReturnError(errors.New("connection refused"))

	_, _, err := src.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCockroachDBSource_Keys(t *testing.T) {
	src, mock := newMockCockroachDBSource(t)
	mock.ExpectQuery("SELECT key FROM env_vars").
		WillReturnRows(sqlmock.NewRows([]string{"key"}).AddRow("DB_HOST").AddRow("DB_PORT"))

	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("got %d keys, want 2", len(keys))
	}
}

func TestCockroachDBSource_PrefixStripped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	src := newCockroachDBSourceWithClient(db, "env_vars", "APP_")

	mock.ExpectQuery("SELECT value FROM env_vars").
		WithArgs("APP_SECRET").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("s3cr3t"))

	val, ok, err := src.Get(context.Background(), "SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || val != "s3cr3t" {
		t.Errorf("got ok=%v val=%q", ok, val)
	}
}
