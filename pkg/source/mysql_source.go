package source

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// mysqlClient abstracts the database operations for testing.
type mysqlClient interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Close() error
}

// MySQLSource retrieves environment variables from a MySQL table.
// It expects a table with at least two columns: key and value.
type MySQLSource struct {
	client    mysqlClient
	table     string
	keyCol    string
	valueCol  string
}

// NewMySQLSource connects to MySQL using the given DSN and returns a MySQLSource.
// The table, keyCol, and valueCol parameters define where secrets are stored.
func NewMySQLSource(dsn, table, keyCol, valueCol string) (*MySQLSource, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql_source: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql_source: ping: %w", err)
	}
	return newMySQLSourceWithClient(db, table, keyCol, valueCol), nil
}

func newMySQLSourceWithClient(client mysqlClient, table, keyCol, valueCol string) *MySQLSource {
	return &MySQLSource{
		client:   client,
		table:    table,
		keyCol:   keyCol,
		valueCol: valueCol,
	}
}

// Get retrieves the value for the given key from the MySQL table.
func (s *MySQLSource) Get(ctx context.Context, key string) (string, bool, error) {
	query := fmt.Sprintf("SELECT `%s` FROM `%s` WHERE `%s` = ? LIMIT 1", s.valueCol, s.table, s.keyCol)
	var value string
	err := s.client.QueryRowContext(ctx, query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("mysql_source: get %q: %w", key, err)
	}
	return value, true, nil
}

// Keys returns all keys present in the MySQL table.
func (s *MySQLSource) Keys(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("SELECT `%s` FROM `%s`", s.keyCol, s.table)
	rows, err := s.client.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("mysql_source: keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("mysql_source: keys scan: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
