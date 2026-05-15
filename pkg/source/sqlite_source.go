package source

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// sqliteQuerier abstracts the sql.DB query interface for testing.
type sqliteQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
	Close() error
}

// SQLiteSource reads environment variables from a SQLite database table.
// The table must have at least two columns: a key column and a value column.
type SQLiteSource struct {
	db         sqliteQuerier
	table      string
	keyColumn  string
	valColumn  string
}

// SQLiteOption configures a SQLiteSource.
type SQLiteOption func(*SQLiteSource)

// WithSQLiteTable sets the table name (default: "env").
func WithSQLiteTable(table string) SQLiteOption {
	return func(s *SQLiteSource) { s.table = table }
}

// WithSQLiteColumns sets the key and value column names (defaults: "key", "value").
func WithSQLiteColumns(key, value string) SQLiteOption {
	return func(s *SQLiteSource) {
		s.keyColumn = key
		s.valColumn = value
	}
}

// NewSQLiteSource opens a SQLite database at the given path and returns a Source.
func NewSQLiteSource(path string, opts ...SQLiteOption) (*SQLiteSource, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite_source: open %q: %w", path, err)
	}
	return newSQLiteSourceWithClient(db, opts...), nil
}

func newSQLiteSourceWithClient(db sqliteQuerier, opts ...SQLiteOption) *SQLiteSource {
	s := &SQLiteSource{
		db:        db,
		table:     "env",
		keyColumn: "key",
		valColumn: "value",
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Get retrieves the value for the given key from the SQLite table.
func (s *SQLiteSource) Get(key string) (string, bool) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1",
		s.valColumn, s.table, s.keyColumn)
	var val string
	err := s.db.QueryRow(query, key).Scan(&val)
	if err != nil {
		return "", false
	}
	return val, true
}

// Keys returns all keys present in the SQLite table.
func (s *SQLiteSource) Keys() []string {
	query := fmt.Sprintf("SELECT %s FROM %s", s.keyColumn, s.table)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err == nil {
			keys = append(keys, k)
		}
	}
	return keys
}
