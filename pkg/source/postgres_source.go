package source

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// postgresClient defines the interface for querying a Postgres database.
type postgresClient interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// PostgresSource reads environment variables from a Postgres table.
// The table is expected to have at least two columns: key and value.
type PostgresSource struct {
	db        postgresClient
	table     string
	keyCol    string
	valueCol  string
}

// PostgresOptions configures the PostgresSource.
type PostgresOptions struct {
	Table    string
	KeyCol   string
	ValueCol string
}

// NewPostgresSource creates a PostgresSource connected to the given DSN.
func NewPostgresSource(dsn string, opts PostgresOptions) (*PostgresSource, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres source: open: %w", err)
	}
	return newPostgresSourceWithClient(db, opts), nil
}

func newPostgresSourceWithClient(db postgresClient, opts PostgresOptions) *PostgresSource {
	if opts.Table == "" {
		opts.Table = "env_vars"
	}
	if opts.KeyCol == "" {
		opts.KeyCol = "key"
	}
	if opts.ValueCol == "" {
		opts.ValueCol = "value"
	}
	return &PostgresSource{db: db, table: opts.Table, keyCol: opts.KeyCol, valueCol: opts.ValueCol}
}

// Get retrieves the value for the given key from Postgres.
func (p *PostgresSource) Get(ctx context.Context, key string) (string, bool, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 LIMIT 1", p.valueCol, p.table, p.keyCol)
	var value string
	err := p.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("postgres source: get %q: %w", key, err)
	}
	return value, true, nil
}

// Keys returns all keys stored in the configured table.
func (p *PostgresSource) Keys(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("SELECT %s FROM %s", p.keyCol, p.table)
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres source: keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("postgres source: scan key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
