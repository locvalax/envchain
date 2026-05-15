package source

import (
	"context"
	"testing"

	"cloud.google.com/go/bigtable"
)

// mockBigtableClient implements bigtableClient for testing.
type mockBigtableClient struct {
	rows map[string]bigtable.Row
	err  error
}

func (m *mockBigtableClient) ReadRow(_ context.Context, _, row string) (bigtable.Row, error) {
	if m.err != nil {
		return nil, m.err
	}
	r, ok := m.rows[row]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func makeRow(family, qualifier, value string) bigtable.Row {
	col := family + ":" + qualifier
	return bigtable.Row{
		family: []bigtable.ReadItem{
			{Column: col, Value: []byte(value)},
		},
	}
}

func newMockBigTableSource(rows map[string]bigtable.Row, err error, prefix string) *BigTableSource {
	return &BigTableSource{
		client:    &mockBigtableClient{rows: rows, err: err},
		table:     "config",
		family:    "cf",
		qualifier: "val",
		prefix:    prefix,
		knownKeys: []string{"DB_HOST", "DB_PORT"},
	}
}

func TestBigTableSource_GetFound(t *testing.T) {
	rows := map[string]bigtable.Row{
		"DB_HOST": makeRow("cf", "val", "localhost"),
	}
	s := newMockBigTableSource(rows, nil, "")
	val, ok, err := s.Get(context.Background(), "DB_HOST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "localhost" {
		t.Errorf("expected 'localhost', got %q", val)
	}
}

func TestBigTableSource_GetMissing(t *testing.T) {
	s := newMockBigTableSource(map[string]bigtable.Row{}, nil, "")
	_, ok, err := s.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestBigTableSource_PrefixedGet(t *testing.T) {
	rows := map[string]bigtable.Row{
		"app/API_KEY": makeRow("cf", "val", "secret"),
	}
	s := newMockBigTableSource(rows, nil, "app/")
	val, ok, err := s.Get(context.Background(), "API_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found with prefix")
	}
	if val != "secret" {
		t.Errorf("expected 'secret', got %q", val)
	}
}

func TestBigTableSource_ClientError(t *testing.T) {
	s := newMockBigTableSource(nil, errTest, "")
	_, _, err := s.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error from client")
	}
}

func TestBigTableSource_Keys(t *testing.T) {
	s := newMockBigTableSource(map[string]bigtable.Row{}, nil, "")
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
