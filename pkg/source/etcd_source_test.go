package source

import (
	"context"
	"errors"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
)

// mockEtcdClient implements EtcdClient for testing.
type mockEtcdClient struct {
	data map[string]string
	err  error
}

func (m *mockEtcdClient) Get(_ context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	// WithPrefix: collect all matching keys
	var kvs []*mvccpb.KeyValue
	for k, v := range m.data {
		if key == k || (len(opts) > 0 && len(k) >= len(key) && k[:len(key)] == key) {
			kvs = append(kvs, &mvccpb.KeyValue{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
	}
	return &clientv3.GetResponse{Kvs: kvs}, nil
}

func TestEtcdSource_GetFound(t *testing.T) {
	client := &mockEtcdClient{data: map[string]string{"/app/DB_HOST": "localhost"}}
	s := NewEtcdSource(client, "/app/", time.Second)

	val, ok := s.Get("DB_HOST")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "localhost" {
		t.Errorf("expected 'localhost', got %q", val)
	}
}

func TestEtcdSource_GetMissing(t *testing.T) {
	client := &mockEtcdClient{data: map[string]string{}}
	s := NewEtcdSource(client, "/app/", time.Second)

	_, ok := s.Get("MISSING_KEY")
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestEtcdSource_ClientError(t *testing.T) {
	client := &mockEtcdClient{err: errors.New("connection refused")}
	s := NewEtcdSource(client, "/app/", time.Second)

	_, ok := s.Get("ANY_KEY")
	if ok {
		t.Fatal("expected false on client error")
	}
}

func TestEtcdSource_Keys(t *testing.T) {
	client := &mockEtcdClient{data: map[string]string{
		"/app/FOO": "1",
		"/app/BAR": "2",
	}}
	s := NewEtcdSource(client, "/app/", time.Second)

	keys := s.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestEtcdSource_Name(t *testing.T) {
	client := &mockEtcdClient{}
	s := NewEtcdSource(client, "/app/", time.Second)
	if s.Name() != "etcd" {
		t.Errorf("expected name 'etcd', got %q", s.Name())
	}
}
