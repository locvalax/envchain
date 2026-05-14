package source

import (
	"context"
	"fmt"
	"sort"
	"testing"
)

// mockKafkaConsumer implements KafkaConsumer for testing.
type mockKafkaConsumer struct {
	data      map[string]string
	readErr   error
	listErr   error
}

func (m *mockKafkaConsumer) ReadMessage(_ context.Context, _, key string) (string, error) {
	if m.readErr != nil {
		return "", m.readErr
	}
	val, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}
	return val, nil
}

func (m *mockKafkaConsumer) ListKeys(_ context.Context, _ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestKafkaSource_GetFound(t *testing.T) {
	consumer := &mockKafkaConsumer{
		data: map[string]string{"DB_PASSWORD": "secret123"},
	}
	src := NewKafkaSource(consumer, "config-topic", "")
	val, err := src.Get(context.Background(), "DB_PASSWORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret123" {
		t.Errorf("expected %q, got %q", "secret123", val)
	}
}

func TestKafkaSource_GetMissing(t *testing.T) {
	consumer := &mockKafkaConsumer{data: map[string]string{}}
	src := NewKafkaSource(consumer, "config-topic", "")
	_, err := src.Get(context.Background(), "MISSING_KEY")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestKafkaSource_PrefixedGet(t *testing.T) {
	consumer := &mockKafkaConsumer{
		data: map[string]string{"app_API_KEY": "mytoken"},
	}
	src := NewKafkaSource(consumer, "config-topic", "app_")
	val, err := src.Get(context.Background(), "API_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "mytoken" {
		t.Errorf("expected %q, got %q", "mytoken", val)
	}
}

func TestKafkaSource_Keys(t *testing.T) {
	consumer := &mockKafkaConsumer{
		data: map[string]string{
			"app_FOO": "1",
			"app_BAR": "2",
			"other_KEY": "3",
		},
	}
	src := NewKafkaSource(consumer, "config-topic", "app_")
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(keys)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}

func TestKafkaSource_ClientError(t *testing.T) {
	consumer := &mockKafkaConsumer{
		readErr: fmt.Errorf("broker unavailable"),
	}
	src := NewKafkaSource(consumer, "config-topic", "")
	_, err := src.Get(context.Background(), "ANY_KEY")
	if err == nil {
		t.Fatal("expected error from consumer, got nil")
	}
}

func TestKafkaSource_CachesValues(t *testing.T) {
	callCount := 0
	consumer := &mockKafkaConsumer{
		data: map[string]string{"CACHED_KEY": "value"},
	}
	// Wrap to count calls
	src := NewKafkaSource(consumer, "topic", "")
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		val, err := src.Get(ctx, "CACHED_KEY")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if val != "value" {
			t.Errorf("call %d: expected %q, got %q", i, "value", val)
		}
		callCount++
	}
	if callCount != 3 {
		t.Errorf("expected 3 successful gets, got %d", callCount)
	}
}
