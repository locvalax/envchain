package source_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-zookeeper/zk"

	"github.com/yourorg/envchain/pkg/source"
)

// mockZookeeperClient implements ZookeeperClient for testing.
type mockZookeeperClient struct {
	data map[string]string
}

func (m *mockZookeeperClient) Get(path string) ([]byte, *zk.Stat, error) {
	val, ok := m.data[path]
	if !ok {
		return nil, nil, zk.ErrNoNode
	}
	return []byte(val), &zk.Stat{}, nil
}

func (m *mockZookeeperClient) Children(path string) ([]string, *zk.Stat, error) {
	prefix := path
	if prefix == "/" {
		prefix = ""
	}
	var keys []string
	for k := range m.data {
		if len(k) > len(prefix)+1 && k[:len(prefix)+1] == prefix+"/" {
			child := k[len(prefix)+1:]
			keys = append(keys, child)
		}
	}
	if len(keys) == 0 {
		return nil, nil, zk.ErrNoNode
	}
	return keys, &zk.Stat{}, nil
}

func newMockZookeeperSource(data map[string]string, prefix string) source.Source {
	return source.NewZookeeperSourceWithClient(&mockZookeeperClient{data: data}, prefix)
}

func TestZookeeperSource_GetFound(t *testing.T) {
	s := newMockZookeeperSource(map[string]string{"/config/DB_HOST": "localhost"}, "/config")
	val, ok, err := s.Get(context.Background(), "DB_HOST")
	if err != nil || !ok || val != "localhost" {
		t.Fatalf("expected (localhost, true, nil), got (%q, %v, %v)", val, ok, err)
	}
}

func TestZookeeperSource_GetMissing(t *testing.T) {
	s := newMockZookeeperSource(map[string]string{}, "/config")
	_, ok, err := s.Get(context.Background(), "MISSING")
	if err != nil || ok {
		t.Fatalf("expected (false, nil), got (%v, %v)", ok, err)
	}
}

func TestZookeeperSource_ClientError(t *testing.T) {
	errClient := &errZookeeperClient{err: errors.New("connection refused")}
	s := source.NewZookeeperSourceWithClient(errClient, "/config")
	_, _, err := s.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestZookeeperSource_Keys(t *testing.T) {
	s := newMockZookeeperSource(map[string]string{
		"/env/APP_ENV": "production",
		"/env/LOG_LEVEL": "info",
	}, "/env")
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}

func TestZookeeperSource_NoPrefix(t *testing.T) {
	s := newMockZookeeperSource(map[string]string{"/SECRET": "abc123"}, "")
	val, ok, err := s.Get(context.Background(), "SECRET")
	if err != nil || !ok || val != "abc123" {
		t.Fatalf("expected (abc123, true, nil), got (%q, %v, %v)", val, ok, err)
	}
}

type errZookeeperClient struct{ err error }

func (e *errZookeeperClient) Get(_ string) ([]byte, *zk.Stat, error) {
	return nil, nil, e.err
}
func (e *errZookeeperClient) Children(_ string) ([]string, *zk.Stat, error) {
	return nil, nil, e.err
}
