//go:build integration
// +build integration

package source_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	source "envchain/pkg/source"
)

// TestSSMSource_Integration requires real AWS credentials and an SSM path set via
// environment variables:
//
//	SSM_TEST_PATH  — the parameter path prefix, e.g. /envchain/test
//	SSM_TEST_KEY   — a key expected to exist under that path
func TestSSMSource_Integration(t *testing.T) {
	path := os.Getenv("SSM_TEST_PATH")
	key := os.Getenv("SSM_TEST_KEY")
	if path == "" || key == "" {
		t.Skip("SSM_TEST_PATH and SSM_TEST_KEY must be set for integration tests")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("failed to load AWS config: %v", err)
	}

	client := ssm.NewFromConfig(cfg)
	s := source.NewSSMSource(client, path)

	val, err := s.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", key, err)
	}
	if val == "" {
		t.Errorf("expected non-empty value for key %q", key)
	}
	t.Logf("SSM key=%q value=%q", key, val)

	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("Keys() failed: %v", err)
	}
	if len(keys) == 0 {
		t.Error("expected at least one key from SSM path")
	}
	t.Logf("SSM keys under %q: %v", path, keys)
}
