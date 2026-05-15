package source

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

type mockAzureBlobClient struct {
	data map[string]string
	err  error
}

func (m *mockAzureBlobClient) DownloadBlob(_ context.Context, _, _ string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	b, _ := json.Marshal(m.data)
	return io.NopCloser(strings.NewReader(string(b))), nil
}

func newMockAzureBlobSource(t *testing.T, data map[string]string, prefix string) Source {
	t.Helper()
	client := &mockAzureBlobClient{data: data}
	s, err := NewAzureBlobSource(client, "my-container", "config.json", prefix)
	if err != nil {
		t.Fatalf("NewAzureBlobSource: %v", err)
	}
	return s
}

func TestAzureBlobSource_GetFound(t *testing.T) {
	s := newMockAzureBlobSource(t, map[string]string{"DB_HOST": "localhost"}, "")
	v, ok := s.Get("DB_HOST")
	if !ok || v != "localhost" {
		t.Errorf("expected (localhost, true), got (%q, %v)", v, ok)
	}
}

func TestAzureBlobSource_GetMissing(t *testing.T) {
	s := newMockAzureBlobSource(t, map[string]string{}, "")
	_, ok := s.Get("MISSING")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestAzureBlobSource_PrefixedGet(t *testing.T) {
	s := newMockAzureBlobSource(t, map[string]string{"APP_PORT": "8080"}, "APP_")
	v, ok := s.Get("PORT")
	if !ok || v != "8080" {
		t.Errorf("expected (8080, true), got (%q, %v)", v, ok)
	}
}

func TestAzureBlobSource_Keys(t *testing.T) {
	s := newMockAzureBlobSource(t, map[string]string{"A": "1", "B": "2"}, "")
	keys := s.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestAzureBlobSource_ClientError(t *testing.T) {
	client := &mockAzureBlobClient{err: errors.New("network error")}
	_, err := NewAzureBlobSource(client, "c", "b", "")
	if err == nil {
		t.Error("expected error on client failure")
	}
}

func TestAzureBlobSource_Name(t *testing.T) {
	s := newMockAzureBlobSource(t, map[string]string{}, "")
	if s.Name() != "azure-blob://my-container/config.json" {
		t.Errorf("unexpected name: %s", s.Name())
	}
}
