package source

import (
	"context"
	"errors"
	"testing"
)

// mockRabbitMQClient implements RabbitMQClient for testing.
type mockRabbitMQClient struct {
	body []byte
	ok   bool
	err  error
}

func (m *mockRabbitMQClient) Get(_ string) ([]byte, bool, error) {
	return m.body, m.ok, m.err
}

func TestRabbitMQSource_GetFound(t *testing.T) {
	client := &mockRabbitMQClient{
		body: []byte(`{"DB_HOST":"localhost","DB_PORT":"5432"}`),
		ok:   true,
	}
	src, err := NewRabbitMQSource(client, "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok, err := src.Get(context.Background(), "DB_HOST")
	if err != nil || !ok || v != "localhost" {
		t.Errorf("expected (localhost, true, nil), got (%q, %v, %v)", v, ok, err)
	}
}

func TestRabbitMQSource_GetMissing(t *testing.T) {
	client := &mockRabbitMQClient{
		body: []byte(`{"DB_HOST":"localhost"}`),
		ok:   true,
	}
	src, err := NewRabbitMQSource(client, "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok, err := src.Get(context.Background(), "MISSING_KEY")
	if err != nil || ok {
		t.Errorf("expected (false, nil), got (%v, %v)", ok, err)
	}
}

func TestRabbitMQSource_ClientError(t *testing.T) {
	client := &mockRabbitMQClient{err: errors.New("connection refused")}
	_, err := NewRabbitMQSource(client, "config")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRabbitMQSource_PrefixedGet(t *testing.T) {
	client := &mockRabbitMQClient{
		body: []byte(`{"APP_SECRET":"topsecret","APP_KEY":"value"}`),
		ok:   true,
	}
	src, err := NewRabbitMQSource(client, "config", WithRabbitMQPrefix("APP_"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok, err := src.Get(context.Background(), "SECRET")
	if err != nil || !ok || v != "topsecret" {
		t.Errorf("expected (topsecret, true, nil), got (%q, %v, %v)", v, ok, err)
	}
}

func TestRabbitMQSource_Keys(t *testing.T) {
	client := &mockRabbitMQClient{
		body: []byte(`{"FOO":"1","BAR":"2"}`),
		ok:   true,
	}
	src, err := NewRabbitMQSource(client, "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestRabbitMQSource_EmptyQueue(t *testing.T) {
	client := &mockRabbitMQClient{ok: false}
	src, err := NewRabbitMQSource(client, "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys, _ := src.Keys(context.Background())
	if len(keys) != 0 {
		t.Errorf("expected 0 keys for empty queue, got %d", len(keys))
	}
}
