package source_test

import (
	"os"
	"testing"

	"github.com/yourorg/envchain/pkg/source"
)

func TestMapSource_GetFound(t *testing.T) {
	s := source.NewMapSource("test", map[string]string{"FOO": "bar"})
	v, ok := s.Get("FOO")
	if !ok || v != "bar" {
		t.Fatalf("expected bar/true, got %q/%v", v, ok)
	}
}

func TestMapSource_GetMissing(t *testing.T) {
	s := source.NewMapSource("test", map[string]string{})
	_, ok := s.Get("MISSING")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestMapSource_Keys(t *testing.T) {
	s := source.NewMapSource("test", map[string]string{"A": "1", "B": "2"})
	keys := s.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestEnvSource_Get(t *testing.T) {
	os.Setenv("ENVCHAIN_TEST_VAR", "hello")
	defer os.Unsetenv("ENVCHAIN_TEST_VAR")

	s := source.NewEnvSource("")
	v, ok := s.Get("ENVCHAIN_TEST_VAR")
	if !ok || v != "hello" {
		t.Fatalf("expected hello/true, got %q/%v", v, ok)
	}
}

func TestEnvSource_PrefixFilter(t *testing.T) {
	os.Setenv("APP_SECRET", "s3cr3t")
	defer os.Unsetenv("APP_SECRET")

	s := source.NewEnvSource("OTHER_")
	_, ok := s.Get("APP_SECRET")
	if ok {
		t.Fatal("expected key to be filtered by prefix")
	}
}

func TestEnvSource_Name(t *testing.T) {
	s := source.NewEnvSource("")
	if s.Name() != "env" {
		t.Fatalf("unexpected name: %s", s.Name())
	}
}
