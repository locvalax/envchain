package source

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
)

// mockNATSEntry implements nats.KeyValueEntry for testing.
type mockNATSEntry struct {
	value []byte
}

func (e *mockNATSEntry) Bucket() string             { return "test" }
func (e *mockNATSEntry) Key() string                { return "" }
func (e *mockNATSEntry) Value() []byte              { return e.value }
func (e *mockNATSEntry) Revision() uint64           { return 0 }
func (e *mockNATSEntry) Delta() uint64              { return 0 }
func (e *mockNATSEntry) Created() interface{}       { return nil }
func (e *mockNATSEntry) Operation() nats.KeyValueOp { return nats.KeyValuePut }

// mockNATSKV is an in-memory NATSKVClient.
type mockNATSKV struct {
	data map[string][]byte
	err  error
}

func (m *mockNATSKV) Get(key string) (nats.KeyValueEntry, error) {
	if m.err != nil {
		return nil, m.err
	}
	v, ok := m.data[key]
	if !ok {
		return nil, nats.ErrKeyNotFound
	}
	return &mockNATSEntry{value: v}, nil
}

func (m *mockNATSKV) Keys() ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestNATSSource_GetFound(t *testing.T) {
	kv := &mockNATSKV{data: map[string][]byte{"DB_PASS": []byte("secret")}}
	s := NewNATSSourceWithClient(kv, "")
	v, ok, err := s.Get(context.Background(), "DB_PASS")
	if err != nil || !ok || v != "secret" {
		t.Fatalf("expected (secret, true, nil), got (%q, %v, %v)", v, ok, err)
	}
}

func TestNATSSource_GetMissing(t *testing.T) {
	kv := &mockNATSKV{data: map[string][]byte{}}
	s := NewNATSSourceWithClient(kv, "")
	_, ok, err := s.Get(context.Background(), "MISSING")
	if err != nil || ok {
		t.Fatalf("expected (false, nil), got (%v, %v)", ok, err)
	}
}

func TestNATSSource_PrefixedGet(t *testing.T) {
	kv := &mockNATSKV{data: map[string][]byte{"app.API_KEY": []byte("xyz")}}
	s := NewNATSSourceWithClient(kv, "app")
	v, ok, err := s.Get(context.Background(), "API_KEY")
	if err != nil || !ok || v != "xyz" {
		t.Fatalf("expected (xyz, true, nil), got (%q, %v, %v)", v, ok, err)
	}
}

func TestNATSSource_ClientError(t *testing.T) {
	kv := &mockNATSKV{err: nats.ErrConnectionClosed}
	s := NewNATSSourceWithClient(kv, "")
	_, _, err := s.Get(context.Background(), "ANY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNATSSource_Keys(t *testing.T) {
	kv := &mockNATSKV{data: map[string][]byte{"env.FOO": []byte("1"), "env.BAR": []byte("2"), "other.X": []byte("3")}}
	s := NewNATSSourceWithClient(kv, "env")
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}
