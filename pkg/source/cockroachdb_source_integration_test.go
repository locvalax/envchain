//go:build integration
// +build integration

package source_test

import (
	"context"
	"os"
	"testing"

	"github.com/yourorg/envchain/pkg/source"
)

// TestCockroachDBSource_Integration requires a running CockroachDB instance.
// Set COCKROACHDB_DSN to the connection string and ensure the env_vars table exists.
//
// Example:
//
//	COCKROACHDB_DSN="postgresql://root@localhost:26257/defaultdb?sslmode=disable" \
//	  go test ./pkg/source/... -tags integration -run TestCockroachDBSource_Integration
func TestCockroachDBSource_Integration(t *testing.T) {
	dsn := os.Getenv("COCKROACHDB_DSN")
	if dsn == "" {
		t.Skip("COCKROACHDB_DSN not set; skipping integration test")
	}

	src, err := source.NewCockroachDBSource(dsn, "env_vars", "")
	if err != nil {
		t.Fatalf("NewCockroachDBSource: %v", err)
	}

	ctx := context.Background()

	// Keys should return at least an empty slice without error.
	keys, err := src.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	t.Logf("found %d key(s) in env_vars", len(keys))

	// A missing key must return ok=false without error.
	_, ok, err := src.Get(ctx, "__envchain_nonexistent__")
	if err != nil {
		t.Fatalf("Get nonexistent: %v", err)
	}
	if ok {
		t.Error("expected ok=false for nonexistent key")
	}
}
