package source_test

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yourorg/envchain/pkg/source"
)

// mockGRPCClient implements source.GRPCClient for testing.
type mockGRPCClient struct {
	data map[string]string
	err  error
}

func (m *mockGRPCClient) GetSecret(_ context.Context, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	val, ok := m.data[key]
	if !ok {
		return "", status.Errorf(codes.NotFound, "key %q not found", key)
	}
	return val, nil
}

func (m *mockGRPCClient) ListKeys(_ context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestGRPCSource_GetFound(t *testing.T) {
	client := &mockGRPCClient{data: map[string]string{"DB_PASS": "secret123"}}
	s := source.NewGRPCSourceWithClient(client, "")
	val, ok := s.Get("DB_PASS")
	if !ok || val != "secret123" {
		t.Fatalf("expected secret123/true, got %q/%v", val, ok)
	}
}

func TestGRPCSource_GetMissing(t *testing.T) {
	client := &mockGRPCClient{data: map[string]string{}}
	s := source.NewGRPCSourceWithClient(client, "")
	_, ok := s.Get("MISSING_KEY")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestGRPCSource_PrefixedGet(t *testing.T) {
	client := &mockGRPCClient{data: map[string]string{"prod/API_KEY": "xyz"}}
	s := source.NewGRPCSourceWithClient(client, "prod/")
	val, ok := s.Get("API_KEY")
	if !ok || val != "xyz" {
		t.Fatalf("expected xyz/true, got %q/%v", val, ok)
	}
}

func TestGRPCSource_ClientError(t *testing.T) {
	client := &mockGRPCClient{err: fmt.Errorf("connection refused")}
	s := source.NewGRPCSourceWithClient(client, "")
	_, ok := s.Get("ANY_KEY")
	if ok {
		t.Fatal("expected false on client error")
	}
}

func TestGRPCSource_Keys(t *testing.T) {
	client := &mockGRPCClient{data: map[string]string{"A": "1", "B": "2"}}
	s := source.NewGRPCSourceWithClient(client, "")
	keys := s.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestGRPCSource_KeysClientError(t *testing.T) {
	client := &mockGRPCClient{err: fmt.Errorf("unavailable")}
	s := source.NewGRPCSourceWithClient(client, "")
	keys := s.Keys()
	if keys != nil {
		t.Fatalf("expected nil keys on error, got %v", keys)
	}
}
