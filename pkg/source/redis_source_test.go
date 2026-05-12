package source

import (
	"context"
	"errors"
	"testing"
)

// mockRedisClient is a simple in-memory mock implementing RedisClient.
type mockRedisClient struct {
	data map[string]string
	err  error
}

func (m *mockRedisClient) Get(_ context.Context, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	val, ok := m.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (m *mockRedisClient) Keys(_ context.Context, pattern string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	prefix := pattern[:len(pattern)-1] // strip trailing '*'
	var keys []string
	for k := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func TestRedisSource_GetFound(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{"app:DB_HOST": "localhost"}}
	src := NewRedisSource(context.Background(), client, "app:")
	val, ok := src.Get("DB_HOST")
	if !ok || val != "localhost" {
		t.Errorf("expected 'localhost', got '%s' (ok=%v)", val, ok)
	}
}

func TestRedisSource_GetMissing(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{}}
	src := NewRedisSource(context.Background(), client, "app:")
	_, ok := src.Get("MISSING_KEY")
	if ok {
		t.Error("expected ok=false for missing key")
	}
}

func TestRedisSource_ClientError(t *testing.T) {
	client := &mockRedisClient{err: errors.New("connection refused")}
	src := NewRedisSource(context.Background(), client, "app:")
	_, ok := src.Get("DB_HOST")
	if ok {
		t.Error("expected ok=false on client error")
	}
}

func TestRedisSource_Keys(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{
		"app:DB_HOST": "localhost",
		"app:DB_PORT": "5432",
		"other:KEY":  "ignored",
	}}
	src := NewRedisSource(context.Background(), client, "app:")
	keys := src.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d: %v", len(keys), keys)
	}
	for _, k := range keys {
		if k == "app:DB_HOST" || k == "app:DB_PORT" {
			t.Errorf("prefix should be stripped from key: %s", k)
		}
	}
}

func TestRedisSource_NoPrefix(t *testing.T) {
	client := &mockRedisClient{data: map[string]string{"API_KEY": "secret"}}
	src := NewRedisSource(context.Background(), client, "")
	val, ok := src.Get("API_KEY")
	if !ok || val != "secret" {
		t.Errorf("expected 'secret', got '%s' (ok=%v)", val, ok)
	}
}
