//go:build integration
// +build integration

package source_test

import (
	"context"
	"os"
	"testing"

	"github.com/your-org/envchain/pkg/source"
)

// TestInfluxDBSource_Integration requires a running InfluxDB instance.
// Set INFLUXDB_URL, INFLUXDB_TOKEN, INFLUXDB_ORG, INFLUXDB_BUCKET in the
// environment before running with -tags integration.
func TestInfluxDBSource_Integration(t *testing.T) {
	serverURL := os.Getenv("INFLUXDB_URL")
	token := os.Getenv("INFLUXDB_TOKEN")
	org := os.Getenv("INFLUXDB_ORG")
	bucket := os.Getenv("INFLUXDB_BUCKET")

	if serverURL == "" || token == "" || org == "" || bucket == "" {
		t.Skip("INFLUXDB_URL / INFLUXDB_TOKEN / INFLUXDB_ORG / INFLUXDB_BUCKET not set")
	}

	s, err := source.NewInfluxDBSource(serverURL, token, org, bucket)
	if err != nil {
		t.Fatalf("NewInfluxDBSource: %v", err)
	}

	ctx := context.Background()

	// Assumes a measurement "envchain" with tag key="TEST_KEY" and field value="hello"
	// has been written to the bucket before running this test.
	v, ok, err := s.Get(ctx, "TEST_KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected TEST_KEY to be present")
	}
	if v == "" {
		t.Fatal("expected non-empty value")
	}
	t.Logf("TEST_KEY = %q", v)

	keys, err := s.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	t.Logf("keys: %v", keys)
}
