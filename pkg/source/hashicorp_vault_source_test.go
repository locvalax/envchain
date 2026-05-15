package source

import (
	"context"
	"errors"
	"testing"
)

type mockHCVaultClient struct {
	data map[string]interface{}
	err  error
}

func (m *mockHCVaultClient) Read(_ context.Context, _ string) (map[string]interface{}, error) {
	return m.data, m.err
}

func (m *mockHCVaultClient) List(_ context.Context, _ string) ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, m.err
}

func TestHashiCorpVaultSource_GetFound(t *testing.T) {
	client := &mockHCVaultClient{data: map[string]interface{}{"DB_PASS": "s3cr3t"}}
	src := NewHashiCorpVaultSource(client, "secret", "myapp/config", "")
	val, ok, err := src.Get(context.Background(), "DB_PASS")
	if err != nil || !ok || val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %q ok=%v err=%v", val, ok, err)
	}
}

func TestHashiCorpVaultSource_GetMissing(t *testing.T) {
	client := &mockHCVaultClient{data: map[string]interface{}{}}
	src := NewHashiCorpVaultSource(client, "secret", "myapp/config", "")
	_, ok, err := src.Get(context.Background(), "MISSING")
	if err != nil || ok {
		t.Fatalf("expected miss, got ok=%v err=%v", ok, err)
	}
}

func TestHashiCorpVaultSource_ClientError(t *testing.T) {
	client := &mockHCVaultClient{err: errors.New("vault unavailable")}
	src := NewHashiCorpVaultSource(client, "secret", "myapp/config", "")
	_, _, err := src.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHashiCorpVaultSource_PrefixedGet(t *testing.T) {
	client := &mockHCVaultClient{data: map[string]interface{}{"API_KEY": "abc123"}}
	src := NewHashiCorpVaultSource(client, "secret", "myapp/config", "APP_")
	val, ok, err := src.Get(context.Background(), "APP_API_KEY")
	if err != nil || !ok || val != "abc123" {
		t.Fatalf("expected abc123, got %q ok=%v err=%v", val, ok, err)
	}
}

func TestHashiCorpVaultSource_Keys(t *testing.T) {
	client := &mockHCVaultClient{data: map[string]interface{}{"FOO": "1", "BAR": "2"}}
	src := NewHashiCorpVaultSource(client, "secret", "myapp/config", "")
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestHashiCorpVaultSource_NilDataReturnsNotFound(t *testing.T) {
	client := &mockHCVaultClient{data: nil}
	src := NewHashiCorpVaultSource(client, "secret", "myapp/config", "")
	_, ok, err := src.Get(context.Background(), "ANY")
	if err != nil || ok {
		t.Fatalf("expected miss for nil data, got ok=%v err=%v", ok, err)
	}
}
