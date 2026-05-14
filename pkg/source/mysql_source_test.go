package source

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockMySQLSource(t *testing.T) (*MySQLSource, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	src := newMySQLSourceWithClient(db, "env_vars", "key", "value")
	return src, mock
}

func TestMySQLSource_GetFound(t *testing.T) {
	src, mock := newMockMySQLSource(t)
	rows := sqlmock.NewRows([]string{"value"}).AddRow("s3cr3t")
	mock.ExpectQuery("SELECT").WithArgs("DB_PASS").WillReturnRows(rows)

	val, ok, err := src.Get(context.Background(), "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != "s3cr3t" {
		t.Errorf("got %q, want %q", val, "s3cr3t")
	}
}

func TestMySQLSource_GetMissing(t *testing.T) {
	src, mock := newMockMySQLSource(t)
	mock.ExpectQuery("SELECT").WithArgs("MISSING").WillReturnError(sql.ErrNoRows)

	_, ok, err := src.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false")
	}
}

func TestMySQLSource_ClientError(t *testing.T) {
	src, mock := newMockMySQLSource(t)
	mock.ExpectQuery("SELECT").WithArgs("KEY").WillReturnError(errors.New("connection refused"))

	_, _, err := src.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMySQLSource_Keys(t *testing.T) {
	src, mock := newMockMySQLSource(t)
	rows := sqlmock.NewRows([]string{"key"}).AddRow("API_KEY").AddRow("DB_PASS")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("got %d keys, want 2", len(keys))
	}
}

func TestMySQLSource_KeysQueryError(t *testing.T) {
	src, mock := newMockMySQLSource(t)
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("query failed"))

	_, err := src.Keys(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
