//go:build integration
// +build integration

package source

import (
	"context"
	"os"
	"testing"
)

// TestElasticsearchSource_Integration requires a running Elasticsearch instance.
// Set ELASTICSEARCH_URL to the base URL (default: http://localhost:9200).
// The test index "envchain_test" must contain a document with id "INTEGRATION_KEY"
// and body {"value": "integration_value"}.
func TestElasticsearchSource_Integration(t *testing.T) {
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200"
	}

	client := NewHTTPElasticsearchClient(url)
	src := NewElasticsearchSource(client, "envchain_test")

	ctx := context.Background()

	v, ok := src.Get(ctx, "INTEGRATION_KEY")
	if !ok {
		t.Fatal("expected INTEGRATION_KEY to be present in Elasticsearch")
	}
	if v != "integration_value" {
		t.Fatalf("expected integration_value, got %q", v)
	}

	v, ok = src.Get(ctx, "NONEXISTENT_KEY")
	if ok {
		t.Fatalf("expected miss for NONEXISTENT_KEY, got %q", v)
	}
}
