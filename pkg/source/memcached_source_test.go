package source

import (
	"context"
	"errors"
	"testing"
)

// mockMemcachedClient is a test double for MemcachedClient.
type mockMemcachedClient struct {
	data map[string][]byte
	err  error
}

func (m *mockMemcachedClient) Get(key string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	v, ok := m.data[key]
	if !ok {
		return nil, errors.New("memcache: cache miss")
	}
	return v, nil
}

func (m *mockMemcachedClient) GetMulti(keys []string) (map[string][]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make(map[string][]byte)
	for _, k := range keys {
		if v, ok := m.data[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

func TestMemcachedSource_GetFound(t *testing.T) {
	client := &mockMemcachedClient{data: map[string][]byte{"APP_SECRET": []byte("hunter2")}}
	s := NewMemcachedSource(client, "", []string{"APP_SECRET"})

	val, ok, err := s.Get(context.Background(), "APP_SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "hunter2" {
		t.Errorf("expected %q, got %q", "hunter2", val)
	}
}

func TestMemcachedSource_GetMissing(t *testing.T) {
	client := &mockMemcachedClient{data: map[string][]byte{}}
	s := NewMemcachedSource(client, "", nil)

	_, ok, err := s.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestMemcachedSource_PrefixedGet(t *testing.T) {
	client := &mockMemcachedClient{data: map[string][]byte{"prod/DB_PASS": []byte("secret")}}
	s := NewMemcachedSource(client, "prod/", []string{"DB_PASS"})

	val, ok, err := s.Get(context.Background(), "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected prefixed key to be found")
	}
	if val != "secret" {
		t.Errorf("expected %q, got %q", "secret", val)
	}
}

func TestMemcachedSource_ClientError(t *testing.T) {
	client := &mockMemcachedClient{err: errors.New("connection refused")}
	s := NewMemcachedSource(client, "", nil)

	_, _, err := s.Get(context.Background(), "ANY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMemcachedSource_Keys(t *testing.T) {
	client := &mockMemcachedClient{data: map[string][]byte{}}
	expected := []string{"KEY_A", "KEY_B"}
	s := NewMemcachedSource(client, "", expected)

	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("key[%d]: expected %q, got %q", i, k, keys[i])
		}
	}
}
