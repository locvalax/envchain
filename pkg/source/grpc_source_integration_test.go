//go:build integration
// +build integration

package source_test

import (
	"os"
	"testing"

	"github.com/yourorg/envchain/pkg/source"
)

// TestGRPCSource_Integration requires a running gRPC secrets service.
// Set GRPC_SOURCE_ADDR to the server address and GRPC_SOURCE_PREFIX
// to an optional key prefix before running.
//
//	GRPC_SOURCE_ADDR=localhost:50051 go test -tags integration ./pkg/source/...
func TestGRPCSource_Integration(t *testing.T) {
	addr := os.Getenv("GRPC_SOURCE_ADDR")
	if addr == "" {
		t.Skip("GRPC_SOURCE_ADDR not set; skipping integration test")
	}

	prefix := os.Getenv("GRPC_SOURCE_PREFIX")
	key := os.Getenv("GRPC_SOURCE_KEY")
	if key == "" {
		t.Skip("GRPC_SOURCE_KEY not set; skipping integration test")
	}

	src, err := source.NewGRPCSource(source.GRPCClientConfig{
		Address: addr,
		Prefix:  prefix,
	})
	if err != nil {
		t.Fatalf("NewGRPCSource: %v", err)
	}

	val, ok := src.Get(key)
	if !ok {
		t.Fatalf("expected key %q to be found", key)
	}
	if val == "" {
		t.Fatalf("expected non-empty value for key %q", key)
	}
	t.Logf("key=%q value length=%d", key, len(val))

	keys := src.Keys()
	t.Logf("total keys returned: %d", len(keys))
}
