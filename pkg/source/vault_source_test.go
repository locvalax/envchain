package source

import (
	"testing"
)

func TestVaultSource_GetFound(t *testing.T) {
	vs, err := NewVaultSource(map[string]string{"DB_PASSWORD": "secret123"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := vs.Get("DB_PASSWORD")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "secret123" {
		t.Errorf("expected 'secret123', got %q", val)
	}
}

func TestVaultSource_GetMissing(t *testing.T) {
	vs, _ := NewVaultSource(map[string]string{}, "")
	_, ok := vs.Get("MISSING")
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestVaultSource_PrefixedGet(t *testing.T) {
	vs, _ := NewVaultSource(map[string]string{"myapp/DB_PASSWORD": "prefixed_secret"}, "myapp/")
	val, ok := vs.Get("DB_PASSWORD")
	if !ok {
		t.Fatal("expected prefixed key to be found")
	}
	if val != "prefixed_secret" {
		t.Errorf("expected 'prefixed_secret', got %q", val)
	}
}

func TestVaultSource_Keys(t *testing.T) {
	vs, _ := NewVaultSource(map[string]string{
		"myapp/API_KEY": "key1",
		"myapp/DB_PASS": "key2",
	}, "myapp/")
	keys := vs.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "API_KEY" || keys[1] != "DB_PASS" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestVaultSource_NilSecretsError(t *testing.T) {
	_, err := NewVaultSource(nil, "")
	if err == nil {
		t.Fatal("expected error for nil secrets map")
	}
}

func TestVaultSource_Name(t *testing.T) {
	vs, _ := NewVaultSource(map[string]string{}, "prod/")
	if vs.Name() != "vault(prefix=prod/)" {
		t.Errorf("unexpected name: %s", vs.Name())
	}
	vs2, _ := NewVaultSource(map[string]string{}, "")
	if vs2.Name() != "vault" {
		t.Errorf("unexpected name: %s", vs2.Name())
	}
}
