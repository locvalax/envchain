package source

import (
	"context"
	"errors"
	"testing"
)

type mockElasticsearchClient struct {
	docs map[string]map[string]interface{}
	err  error
}

func (m *mockElasticsearchClient) Get(_ context.Context, index, id string) (map[string]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	doc, ok := m.docs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return doc, nil
}

func (m *mockElasticsearchClient) Keys(_ context.Context, index string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	keys := make([]string, 0, len(m.docs))
	for k := range m.docs {
		keys = append(keys, k)
	}
	return keys, nil
}

func newMockElasticsearchSource(docs map[string]map[string]interface{}, opts ...func(*elasticsearchSource)) Source {
	return NewElasticsearchSource(&mockElasticsearchClient{docs: docs}, "envchain", opts...)
}

func TestElasticsearchSource_GetFound(t *testing.T) {
	src := newMockElasticsearchSource(map[string]map[string]interface{}{
		"DB_HOST": {"value": "localhost"},
	})
	v, ok := src.Get(context.Background(), "DB_HOST")
	if !ok || v != "localhost" {
		t.Fatalf("expected localhost, got %q ok=%v", v, ok)
	}
}

func TestElasticsearchSource_GetMissing(t *testing.T) {
	src := newMockElasticsearchSource(map[string]map[string]interface{}{})
	_, ok := src.Get(context.Background(), "MISSING")
	if ok {
		t.Fatal("expected miss")
	}
}

func TestElasticsearchSource_ClientError(t *testing.T) {
	src := NewElasticsearchSource(&mockElasticsearchClient{err: errors.New("conn refused")}, "envchain")
	_, ok := src.Get(context.Background(), "ANY")
	if ok {
		t.Fatal("expected miss on error")
	}
}

func TestElasticsearchSource_PrefixedGet(t *testing.T) {
	src := newMockElasticsearchSource(map[string]map[string]interface{}{
		"app/DB_PASS": {"value": "secret"},
	}, WithElasticsearchPrefix("app/"))
	v, ok := src.Get(context.Background(), "DB_PASS")
	if !ok || v != "secret" {
		t.Fatalf("expected secret, got %q ok=%v", v, ok)
	}
}

func TestElasticsearchSource_Keys(t *testing.T) {
	src := newMockElasticsearchSource(map[string]map[string]interface{}{
		"env/FOO": {"value": "1"},
		"env/BAR": {"value": "2"},
		"other/X": {"value": "3"},
	}, WithElasticsearchPrefix("env/"))
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}

func TestElasticsearchSource_KeysError(t *testing.T) {
	src := NewElasticsearchSource(&mockElasticsearchClient{err: errors.New("timeout")}, "envchain")
	_, err := src.Keys(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
