package chain

import (
	"bytes"
	"log"
	"testing"
)

func fixedResolver(m map[string]string) resolveFunc {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestWithDefault_ReturnsSourceValueWhenPresent(t *testing.T) {
	base := fixedResolver(map[string]string{"FOO": "bar"})
	wrapped := WithDefault(base, map[string]string{"FOO": "default_bar"})
	val, ok := wrapped("FOO")
	if !ok || val != "bar" {
		t.Errorf("expected 'bar', got %q (ok=%v)", val, ok)
	}
}

func TestWithDefault_FallsBackToDefault(t *testing.T) {
	base := fixedResolver(map[string]string{})
	wrapped := WithDefault(base, map[string]string{"PORT": "8080"})
	val, ok := wrapped("PORT")
	if !ok || val != "8080" {
		t.Errorf("expected '8080', got %q (ok=%v)", val, ok)
	}
}

func TestWithDefault_MissingNoDefault(t *testing.T) {
	base := fixedResolver(map[string]string{})
	wrapped := WithDefault(base, map[string]string{})
	_, ok := wrapped("MISSING")
	if ok {
		t.Error("expected key to be missing")
	}
}

func TestWithRedaction_MasksSecretKey(t *testing.T) {
	base := fixedResolver(map[string]string{"DB_PASSWORD": "supersecret"})
	wrapped := WithRedaction(base, []string{"PASSWORD", "SECRET", "TOKEN"})
	val, ok := wrapped("DB_PASSWORD")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "***REDACTED***" {
		t.Errorf("expected redacted value, got %q", val)
	}
}

func TestWithRedaction_PassesThroughNonSensitive(t *testing.T) {
	base := fixedResolver(map[string]string{"APP_ENV": "production"})
	wrapped := WithRedaction(base, []string{"PASSWORD", "SECRET"})
	val, ok := wrapped("APP_ENV")
	if !ok || val != "production" {
		t.Errorf("expected 'production', got %q (ok=%v)", val, ok)
	}
}

func TestWithLogging_LogsResolvedKey(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	base := fixedResolver(map[string]string{"API_KEY": "abc123"})
	wrapped := WithLogging(base, logger)
	wrapped("API_KEY")
	if !bytes.Contains(buf.Bytes(), []byte("API_KEY")) {
		t.Errorf("expected log to contain key name, got: %s", buf.String())
	}
}

func TestWithLogging_LogsMissingKey(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	base := fixedResolver(map[string]string{})
	wrapped := WithLogging(base, logger)
	wrapped("UNKNOWN")
	if !bytes.Contains(buf.Bytes(), []byte("missing")) {
		t.Errorf("expected 'missing' in log, got: %s", buf.String())
	}
}
