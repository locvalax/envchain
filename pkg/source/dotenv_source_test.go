package source_test

import (
	"os"
	"testing"

	"github.com/yourorg/envchain/pkg/source"
)

func writeTempDotenv(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), ".env")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestDotenvSource_ParsesKeyValues(t *testing.T) {
	path := writeTempDotenv(t, "FOO=bar\nBAZ=\"quoted\"\n")
	s, err := source.NewDotenvSource(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := s.Get("FOO"); !ok || v != "bar" {
		t.Errorf("FOO: got %q/%v", v, ok)
	}
	if v, ok := s.Get("BAZ"); !ok || v != "quoted" {
		t.Errorf("BAZ: got %q/%v", v, ok)
	}
}

func TestDotenvSource_IgnoresComments(t *testing.T) {
	path := writeTempDotenv(t, "# this is a comment\nKEY=value\n")
	s, err := source.NewDotenvSource(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Get("# this is a comment"); ok {
		t.Error("comment line should not be parsed as key")
	}
	if v, ok := s.Get("KEY"); !ok || v != "value" {
		t.Errorf("KEY: got %q/%v", v, ok)
	}
}

func TestDotenvSource_FileNotFound(t *testing.T) {
	_, err := source.NewDotenvSource("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDotenvSource_Keys(t *testing.T) {
	path := writeTempDotenv(t, "X=1\nY=2\nZ=3\n")
	s, err := source.NewDotenvSource(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Keys()) != 3 {
		t.Errorf("expected 3 keys, got %d", len(s.Keys()))
	}
}
