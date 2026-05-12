package source

import (
	"context"
	"errors"
	"testing"
)

// mockConsulClient is a test double for ConsulKVClient.
type mockConsulClient struct {
	data map[string]string
	err  error
}

func (m *mockConsulClient) Get(_ context.Context, key string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *mockConsulClient) Keys(_ context.Context, prefix string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	var keys []string
	for k := range m.data {
		if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func TestConsulSource_GetFound(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{"myapp/prod/DB_HOST": "db.example.com"}}
	src := NewConsulSource(context.Background(), client, "myapp/prod")
	val, ok, err := src.Get("DB_HOST")
	if err != nil || !ok || val != "db.example.com" {
		t.Fatalf("expected db.example.com, got %q ok=%v err=%v", val, ok, err)
	}
}

func TestConsulSource_GetMissing(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{}}
	src := NewConsulSource(context.Background(), client, "myapp/prod")
	_, ok, err := src.Get("MISSING")
	if err != nil || ok {
		t.Fatalf("expected missing key, got ok=%v err=%v", ok, err)
	}
}

func TestConsulSource_ClientError(t *testing.T) {
	client := &mockConsulClient{err: errors.New("connection refused")}
	src := NewConsulSource(context.Background(), client, "")
	_, _, err := src.Get("KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConsulSource_Keys(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{
		"app/DB_HOST": "localhost",
		"app/DB_PORT": "5432",
	}}
	src := NewConsulSource(context.Background(), client, "app")
	keys, err := src.Keys()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}

func TestConsulSource_KeysClientError(t *testing.T) {
	client := &mockConsulClient{err: errors.New("timeout")}
	src := NewConsulSource(context.Background(), client, "app")
	_, err := src.Keys()
	if err == nil {
		t.Fatal("expected error from Keys(), got nil")
	}
}

func TestConsulSource_NoPrefix(t *testing.T) {
	client := &mockConsulClient{data: map[string]string{"API_KEY": "secret"}}
	src := NewConsulSource(context.Background(), client, "")
	val, ok, err := src.Get("API_KEY")
	if err != nil || !ok || val != "secret" {
		t.Fatalf("expected secret, got %q ok=%v err=%v", val, ok, err)
	}
}
