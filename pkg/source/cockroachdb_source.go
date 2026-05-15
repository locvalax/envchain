package source

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// cockroachDBClient abstracts the database operations used by CockroachDBSource.
type cockroachDBClient interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// CockroachDBSource reads environment variables from a CockroachDB table.
// The table is expected to have at least two columns: key (text) and value (text).
type CockroachDBSource struct {
	client cockroachDBClient
	table  string
	prefix string
}

// NewCockroachDBSource opens a CockroachDB connection and returns a new source.
func NewCockroachDBSource(dsn, table, prefix string) (*CockroachDBSource, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("cockroachdb: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("cockroachdb: ping: %w", err)
	}
	return newCockroachDBSourceWithClient(db, table, prefix), nil
}

func newCockroachDBSourceWithClient(client cockroachDBClient, table, prefix string) *CockroachDBSource {
	return &CockroachDBSource{client: client, table: table, prefix: prefix}
}

// Get retrieves the value for the given key from the CockroachDB table.
func (s *CockroachDBSource) Get(ctx context.Context, key string) (string, bool, error) {
	lookup := s.prefix + key
	var value string
	query := fmt.Sprintf("SELECT value FROM %s WHERE key = $1 LIMIT 1", s.table)
	err := s.client.QueryRowContext(ctx, query, lookup).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("cockroachdb: get %q: %w", lookup, err)
	}
	return value, true, nil
}

// Keys returns all keys stored in the table, stripping the configured prefix.
func (s *CockroachDBSource) Keys(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("SELECT key FROM %s ORDER BY key", s.table)
	rows, err := s.client.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("cockroachdb: keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("cockroachdb: scan key: %w", err)
		}
		if s.prefix != "" && len(k) > len(s.prefix) && k[:len(s.prefix)] == s.prefix {
			k = k[len(s.prefix):]
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
