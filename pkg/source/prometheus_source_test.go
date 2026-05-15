package source

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func prometheusServer(t *testing.T, results []map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result":     results,
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func prometheusEmptyServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"resultType": "vector", "result": []interface{}{}},
		})
	}))
}

func TestPrometheusSource_GetFound(t *testing.T) {
	results := []map[string]interface{}{
		{"metric": map[string]string{}, "value": []interface{}{1234567890, "42"}},
	}
	srv := prometheusServer(t, results)
	defer srv.Close()

	s := NewPrometheusSource(srv.URL)
	val, ok, err := s.Get(context.Background(), "up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "42" {
		t.Errorf("expected '42', got %q", val)
	}
}

func TestPrometheusSource_GetMissing(t *testing.T) {
	srv := prometheusEmptyServer(t)
	defer srv.Close()

	s := NewPrometheusSource(srv.URL)
	_, ok, err := s.Get(context.Background(), "missing_metric")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestPrometheusSource_PrefixedGet(t *testing.T) {
	results := []map[string]interface{}{
		{"metric": map[string]string{}, "value": []interface{}{0, "99"}},
	}
	srv := prometheusServer(t, results)
	defer srv.Close()

	s := NewPrometheusSource(srv.URL, WithPrometheusPrefix("PROM_"))
	val, ok, err := s.Get(context.Background(), "PROM_cpu_usage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || val != "99" {
		t.Errorf("expected '99', got %q (ok=%v)", val, ok)
	}
}

func TestPrometheusSource_CustomQuery(t *testing.T) {
	results := []map[string]interface{}{
		{"metric": map[string]string{}, "value": []interface{}{0, "7"}},
	}
	srv := prometheusServer(t, results)
	defer srv.Close()

	s := NewPrometheusSource(srv.URL, WithPrometheusQueries(map[string]string{
		"REPLICA_COUNT": `sum(up{job="api"})`,
	}))
	val, ok, err := s.Get(context.Background(), "REPLICA_COUNT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || val != "7" {
		t.Errorf("expected '7', got %q (ok=%v)", val, ok)
	}
}

func TestPrometheusSource_Keys(t *testing.T) {
	s := NewPrometheusSource("http://localhost:9090",
		WithPrometheusKnownKeys([]string{"up", "cpu_usage"}),
	)
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestPrometheusSource_ClientError(t *testing.T) {
	s := NewPrometheusSource("http://127.0.0.1:1")
	_, ok, err := s.Get(context.Background(), "up")
	if err == nil {
		t.Fatal("expected error from unreachable server")
	}
	if ok {
		t.Fatal("expected ok=false on error")
	}
}
