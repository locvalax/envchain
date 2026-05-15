package source

import (
	"context"
	"errors"
	"testing"
)

type mockServiceBusReceiver struct {
	body string
	err  error
}

func (m *mockServiceBusReceiver) ReceiveMessage(_ context.Context) (string, error) {
	return m.body, m.err
}

func newMockServiceBusSource(t *testing.T, body string, prefix string, knownKeys []string) Source {
	t.Helper()
	src, err := NewAzureServiceBusSource(context.Background(), &mockServiceBusReceiver{body: body}, prefix, knownKeys)
	if err != nil {
		t.Fatalf("unexpected error creating source: %v", err)
	}
	return src
}

func TestAzureServiceBusSource_GetFound(t *testing.T) {
	src := newMockServiceBusSource(t, `{"DB_HOST":"localhost","DB_PORT":"5432"}`, "", nil)
	v, ok := src.Get("DB_HOST")
	if !ok || v != "localhost" {
		t.Errorf("expected (localhost, true), got (%q, %v)", v, ok)
	}
}

func TestAzureServiceBusSource_GetMissing(t *testing.T) {
	src := newMockServiceBusSource(t, `{"DB_HOST":"localhost"}`, "", nil)
	_, ok := src.Get("DB_PASS")
	if ok {
		t.Error("expected key to be missing")
	}
}

func TestAzureServiceBusSource_PrefixedGet(t *testing.T) {
	src := newMockServiceBusSource(t, `{"APP_DB_HOST":"db.internal"}`, "APP_", nil)
	v, ok := src.Get("DB_HOST")
	if !ok || v != "db.internal" {
		t.Errorf("expected (db.internal, true), got (%q, %v)", v, ok)
	}
}

func TestAzureServiceBusSource_Keys(t *testing.T) {
	src := newMockServiceBusSource(t, `{"FOO":"1","BAR":"2"}`, "", nil)
	keys := src.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestAzureServiceBusSource_KnownKeys(t *testing.T) {
	src := newMockServiceBusSource(t, `{"FOO":"1","BAR":"2"}`, "", []string{"FOO"})
	keys := src.Keys()
	if len(keys) != 1 || keys[0] != "FOO" {
		t.Errorf("expected [FOO], got %v", keys)
	}
}

func TestAzureServiceBusSource_ClientError(t *testing.T) {
	_, err := NewAzureServiceBusSource(context.Background(), &mockServiceBusReceiver{err: errors.New("connection refused")}, "", nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAzureServiceBusSource_InvalidJSON(t *testing.T) {
	_, err := NewAzureServiceBusSource(context.Background(), &mockServiceBusReceiver{body: "not-json"}, "", nil)
	if err == nil {
		t.Error("expected JSON parse error, got nil")
	}
}
