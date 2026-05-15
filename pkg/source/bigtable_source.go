package source

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigtable"
)

// bigtableClient defines the interface for Bigtable row reading.
type bigtableClient interface {
	ReadRow(ctx context.Context, table, row string) (bigtable.Row, error)
}

type realBigtableClient struct {
	tables map[string]*bigtable.Table
	client *bigtable.Client
}

func (r *realBigtableClient) ReadRow(ctx context.Context, tableName, row string) (bigtable.Row, error) {
	tbl, ok := r.tables[tableName]
	if !ok {
		tbl = r.client.Open(tableName)
		r.tables[tableName] = tbl
	}
	return tbl.ReadRow(ctx, row)
}

// BigTableSource reads environment variables from a Google Cloud Bigtable table.
// Each row key is a variable name; the value is read from the specified column family and qualifier.
type BigTableSource struct {
	client     bigtableClient
	table      string
	family     string
	qualifier  string
	prefix     string
	knownKeys  []string
}

// NewBigTableSource creates a BigTableSource backed by a real Bigtable client.
func NewBigTableSource(ctx context.Context, project, instance, table, family, qualifier string, opts ...option) (*BigTableSource, error) {
	client, err := bigtable.NewClient(ctx, project, instance)
	if err != nil {
		return nil, fmt.Errorf("bigtable: failed to create client: %w", err)
	}
	s := &BigTableSource{
		client:    &realBigtableClient{client: client, tables: make(map[string]*bigtable.Table)},
		table:     table,
		family:    family,
		qualifier: qualifier,
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

type option func(*BigTableSource)

// WithBigTablePrefix filters keys by prefix when listing.
func WithBigTablePrefix(prefix string) option {
	return func(s *BigTableSource) { s.prefix = prefix }
}

// WithBigTableKnownKeys sets a static list of known keys for Keys().
func WithBigTableKnownKeys(keys []string) option {
	return func(s *BigTableSource) { s.knownKeys = keys }
}

// Get retrieves the value for the given key from Bigtable.
func (s *BigTableSource) Get(ctx context.Context, key string) (string, bool, error) {
	rowKey := key
	if s.prefix != "" {
		rowKey = s.prefix + key
	}
	row, err := s.client.ReadRow(ctx, s.table, rowKey)
	if err != nil {
		return "", false, fmt.Errorf("bigtable: read row %q: %w", rowKey, err)
	}
	if row == nil {
		return "", false, nil
	}
	col := s.family + ":" + s.qualifier
	for _, item := range row[s.family] {
		if item.Column == col {
			return string(item.Value), true, nil
		}
	}
	return "", false, nil
}

// Keys returns the known keys for this source.
func (s *BigTableSource) Keys(ctx context.Context) ([]string, error) {
	return s.knownKeys, nil
}
