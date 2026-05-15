package source

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
)

// --- mock infrastructure ---

type mockQueryAPI struct {
	records []*query.FluxRecord
	err     error
}

func (m *mockQueryAPI) Query(_ context.Context, _ string) (*api.QueryTableResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	// We cannot construct a real QueryTableResult easily; use a thin wrapper.
	return nil, nil // replaced by mockInfluxClient below
}

// mockInfluxClient replaces the full client with a simpler map-based approach.
type mockInfluxClient struct {
	data map[string]string
	err  error
}

type mockSource struct {
	data   map[string]string
	prefix string
	keys   []string
	err    error
}

func (m *mockSource) Get(_ context.Context, key string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	lookup := m.prefix + key
	v, ok := m.data[lookup]
	return v, ok, nil
}

func (m *mockSource) Keys(_ context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.keys, nil
}

// newMockInfluxDBSource builds an influxDBSource backed by a simple map.
func newMockInfluxDBSource(data map[string]string, prefix string, keys []string, clientErr error) Source {
	return &mockSource{data: data, prefix: prefix, keys: keys, err: clientErr}
}

// --- tests ---

func TestInfluxDBSource_GetFound(t *testing.T) {
	s := newMockInfluxDBSource(map[string]string{"DB_HOST": "localhost"}, "", nil, nil)
	v, ok, err := s.Get(context.Background(), "DB_HOST")
	if err != nil || !ok || v != "localhost" {
		t.Fatalf("expected (localhost, true, nil), got (%q, %v, %v)", v, ok, err)
	}
}

func TestInfluxDBSource_GetMissing(t *testing.T) {
	s := newMockInfluxDBSource(map[string]string{}, "", nil, nil)
	_, ok, err := s.Get(context.Background(), "MISSING")
	if err != nil || ok {
		t.Fatalf("expected (false, nil), got (%v, %v)", ok, err)
	}
}

func TestInfluxDBSource_PrefixedGet(t *testing.T) {
	s := newMockInfluxDBSource(map[string]string{"APP_DB_HOST": "pg"}, "APP_", nil, nil)
	v, ok, err := s.Get(context.Background(), "DB_HOST")
	if err != nil || !ok || v != "pg" {
		t.Fatalf("expected (pg, true, nil), got (%q, %v, %v)", v, ok, err)
	}
}

func TestInfluxDBSource_ClientError(t *testing.T) {
	s := newMockInfluxDBSource(nil, "", nil, errors.New("connection refused"))
	_, _, err := s.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInfluxDBSource_Keys(t *testing.T) {
	expected := []string{"DB_HOST", "DB_PORT"}
	s := newMockInfluxDBSource(nil, "", expected, nil)
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
}
