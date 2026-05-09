//go:build integration
// +build integration

package source_test

import (
	"context"
	"os"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/yourorg/envchain/pkg/source"
)

// TestSecretManagerSource_Integration requires a real GCP project and credentials.
// Run with: go test -tags=integration ./pkg/source/... -run TestSecretManagerSource_Integration
//
// Required env vars:
//   GCP_PROJECT_ID   — the GCP project containing your secrets
//   SM_SECRET_KEY    — name of a secret to look up (without prefix)
//   SM_PREFIX        — optional prefix applied to secret names
func TestSecretManagerSource_Integration(t *testing.T) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		t.Skip("GCP_PROJECT_ID not set; skipping integration test")
	}
	secretKey := os.Getenv("SM_SECRET_KEY")
	if secretKey == "" {
		t.Skip("SM_SECRET_KEY not set; skipping integration test")
	}
	prefix := os.Getenv("SM_PREFIX")

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		t.Fatalf("failed to create Secret Manager client: %v", err)
	}
	defer client.Close()

	src := source.NewSecretManagerSource(client, projectID, prefix)

	val, ok, err := src.Get(ctx, secretKey)
	if err != nil {
		t.Fatalf("Get(%q) error: %v", secretKey, err)
	}
	if !ok {
		t.Errorf("Get(%q): expected secret to exist", secretKey)
	}
	if val == "" {
		t.Errorf("Get(%q): expected non-empty value", secretKey)
	}
	t.Logf("secret %q resolved successfully (value redacted)", secretKey)

	keys, err := src.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys() error: %v", err)
	}
	t.Logf("found %d keys under prefix %q", len(keys), prefix)
}
