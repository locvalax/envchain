package source

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	statuscode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statuscode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
	}, nil
}

func jsonBody(t *testing.T, data map[string]string) string {
	t.Helper()
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("jsonBody: %v", err)
	}
	return string(b)
}

func TestHTTPSource_GetFound(t *testing.T) {
	client := &mockHTTPClient{
		statuscode: 200,
		body:       jsonBody(t, map[string]string{"API_KEY": "secret123", "ENV": "prod"}),
	}
	s, err := NewHTTPSourceWithClient("http://example.com/env", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := s.Get("API_KEY")
	if !ok || v != "secret123" {
		t.Errorf("expected secret123, got %q (ok=%v)", v, ok)
	}
}

func TestHTTPSource_GetMissing(t *testing.T) {
	client := &mockHTTPClient{
		statuscode: 200,
		body:       jsonBody(t, map[string]string{"FOO": "bar"}),
	}
	s, err := NewHTTPSourceWithClient("http://example.com/env", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := s.Get("MISSING")
	if ok {
		t.Error("expected missing key to return false")
	}
}

func TestHTTPSource_NonOKStatus(t *testing.T) {
	client := &mockHTTPClient{statuscode: 403, body: "forbidden"}
	_, err := NewHTTPSourceWithClient("http://example.com/env", client)
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestHTTPSource_InvalidJSON(t *testing.T) {
	client := &mockHTTPClient{statuscode: 200, body: "not-json"}
	_, err := NewHTTPSourceWithClient("http://example.com/env", client)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHTTPSource_Keys(t *testing.T) {
	client := &mockHTTPClient{
		statuscode: 200,
		body:       jsonBody(t, map[string]string{"A": "1", "B": "2", "C": "3"}),
	}
	s, err := NewHTTPSourceWithClient("http://example.com/env", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys := s.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}
