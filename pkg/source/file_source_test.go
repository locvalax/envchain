package source

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "envfile")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func TestFileSource_GetFound(t *testing.T) {
	p := writeTempFile(t, "FOO=bar\nBAZ=qux\n")
	s := NewFileSource(p)
	v, ok, err := s.Get("FOO")
	if err != nil || !ok || v != "bar" {
		t.Fatalf("expected bar, got %q ok=%v err=%v", v, ok, err)
	}
}

func TestFileSource_GetMissing(t *testing.T) {
	p := writeTempFile(t, "FOO=bar\n")
	s := NewFileSource(p)
	_, ok, err := s.Get("MISSING")
	if err != nil || ok {
		t.Fatalf("expected missing, got ok=%v err=%v", ok, err)
	}
}

func TestFileSource_IgnoresComments(t *testing.T) {
	p := writeTempFile(t, "# this is a comment\nKEY=value\n")
	s := NewFileSource(p)
	v, ok, err := s.Get("KEY")
	if err != nil || !ok || v != "value" {
		t.Fatalf("expected value, got %q ok=%v err=%v", v, ok, err)
	}
	_, ok, _ = s.Get("# this is a comment")
	if ok {
		t.Fatal("comment line should not be parsed as a key")
	}
}

func TestFileSource_IgnoresBlankLines(t *testing.T) {
	p := writeTempFile(t, "\nA=1\n\nB=2\n\n")
	s := NewFileSource(p)
	keys, err := s.Keys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestFileSource_Keys(t *testing.T) {
	p := writeTempFile(t, "X=1\nY=2\nZ=3\n")
	s := NewFileSource(p)
	keys, err := s.Keys()
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(keys)
	expected := []string{"X", "Y", "Z"}
	for i, k := range expected {
		if keys[i] != k {
			t.Fatalf("expected %v, got %v", expected, keys)
		}
	}
}

func TestFileSource_FileNotFound(t *testing.T) {
	s := NewFileSource("/nonexistent/path/envfile")
	_, _, err := s.Get("KEY")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
