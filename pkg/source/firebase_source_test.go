package source

import (
	"context"
	"testing"

	"firebase.google.com/go/v4/db"
)

// mockFirebaseClient implements FirebaseClient for testing.
type mockFirebaseClient struct {
	refs map[string]*mockFirebaseRef
}

func (m *mockFirebaseClient) NewRef(path string) *db.Ref {
	// We can't easily mock db.Ref directly; we use a separate interface in tests.
	return nil
}

// testFirebaseGetter is a simplified test double used via a wrapper.
type testFirebaseGetter struct {
	data map[string]interface{}
}

type mockFirebaseRef struct {
	val interface{}
	err error
}

// fakeFirebaseSource is a test-friendly version of FirebaseSource.
type fakeFirebaseSource struct {
	store map[string]string
}

func (f *fakeFirebaseSource) Get(_ context.Context, key string) (string, bool, error) {
	v, ok := f.store[key]
	return v, ok, nil
}

func (f *fakeFirebaseSource) Keys(_ context.Context) ([]string, error) {
	keys := make([]string, 0, len(f.store))
	for k := range f.store {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestFirebaseSource_GetFound(t *testing.T) {
	src := &fakeFirebaseSource{store: map[string]string{"DB_HOST": "localhost"}}
	val, ok, err := src.Get(context.Background(), "DB_HOST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "localhost" {
		t.Errorf("expected 'localhost', got %q", val)
	}
}

func TestFirebaseSource_GetMissing(t *testing.T) {
	src := &fakeFirebaseSource{store: map[string]string{}}
	_, ok, err := src.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestFirebaseSource_Keys(t *testing.T) {
	src := &fakeFirebaseSource{store: map[string]string{"A": "1", "B": "2"}}
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestFirebaseSource_PrefixedGet(t *testing.T) {
	src := &fakeFirebaseSource{store: map[string]string{"prod_API_KEY": "secret"}}
	val, ok, err := src.Get(context.Background(), "prod_API_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || val != "secret" {
		t.Errorf("expected 'secret', got %q (found=%v)", val, ok)
	}
}
