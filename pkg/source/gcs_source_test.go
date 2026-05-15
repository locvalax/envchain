package source

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"testing"
)

type mockGCSClient struct {
	data map[string]string
	err  error
}

func (m *mockGCSClient) GetObject(_ context.Context, _, _ string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	b, err := json.Marshal(m.data)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func newMockGCSSource(data map[string]string, prefix string) (*GCSSource, error) {
	client := &mockGCSClient{data: data}
	return newGCSSourceWithClient(client, "test-bucket", "env.json", prefix)
}

func TestGCSSource_GetFound(t *testing.T) {
	src, err := newMockGCSSource(map[string]string{"DB_HOST": "localhost"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := src.Get("DB_HOST")
	if !ok || v != "localhost" {
		t.Errorf("expected (localhost, true), got (%q, %v)", v, ok)
	}
}

func TestGCSSource_GetMissing(t *testing.T) {
	src, err := newMockGCSSource(map[string]string{}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := src.Get("MISSING")
	if ok {
		t.Error("expected key to be missing")
	}
}

func TestGCSSource_PrefixedGet(t *testing.T) {
	src, err := newMockGCSSource(map[string]string{"APP_DB_HOST": "pg"}, "APP_")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := src.Get("DB_HOST")
	if !ok || v != "pg" {
		t.Errorf("expected (pg, true), got (%q, %v)", v, ok)
	}
}

func TestGCSSource_ClientError(t *testing.T) {
	client := &mockGCSClient{err: errors.New("network error")}
	_, err := newGCSSourceWithClient(client, "bucket", "key", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGCSSource_Keys(t *testing.T) {
	src, err := newMockGCSSource(map[string]string{"A": "1", "B": "2", "C": "3"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys := src.Keys()
	sort.Strings(keys)
	if len(keys) != 3 || keys[0] != "A" || keys[1] != "B" || keys[2] != "C" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestGCSSource_InvalidJSON(t *testing.T) {
	client := &mockGCSClient{}
	// Override GetObject to return invalid JSON by temporarily breaking the data
	// We do this by providing a custom client that returns bad bytes directly.
	badClient := &badJSONGCSClient{}
	_, err := newGCSSourceWithClient(badClient, "bucket", "key", "")
	if err == nil {
		t.Fatal("expected JSON parse error, got nil")
	}
	_ = client
}

type badJSONGCSClient struct{}

func (b *badJSONGCSClient) GetObject(_ context.Context, _, _ string) ([]byte, error) {
	return []byte("not-json"), nil
}
