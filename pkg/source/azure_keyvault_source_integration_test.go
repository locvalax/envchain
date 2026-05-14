//go:build integration
// +build integration

package source

import (
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// TestAzureKeyVaultSource_Integration requires the following environment variables:
//   AZURE_KEYVAULT_URL    — e.g. https://my-vault.vault.azure.net/
//   AZURE_KEYVAULT_KEY    — secret name (as env key, e.g. MY_SECRET)
//   AZURE_KEYVAULT_VALUE  — expected secret value
//
// Authentication uses DefaultAzureCredential (env vars, managed identity, CLI, etc.)
func TestAzureKeyVaultSource_Integration(t *testing.T) {
	vaultURL := os.Getenv("AZURE_KEYVAULT_URL")
	if vaultURL == "" {
		t.Skip("AZURE_KEYVAULT_URL not set, skipping integration test")
	}

	envKey := os.Getenv("AZURE_KEYVAULT_KEY")
	expectedVal := os.Getenv("AZURE_KEYVAULT_VALUE")
	if envKey == "" {
		t.Skip("AZURE_KEYVAULT_KEY not set, skipping integration test")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("failed to create Azure credential: %v", err)
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		t.Fatalf("failed to create Key Vault client: %v", err)
	}

	src := NewAzureKeyVaultSource(client, "")

	val, ok := src.Get(envKey)
	if !ok {
		t.Fatalf("expected key %q to be found in Key Vault", envKey)
	}
	if expectedVal != "" && val != expectedVal {
		t.Errorf("expected value %q, got %q", expectedVal, val)
	}
	t.Logf("Successfully retrieved secret for key %q", envKey)
}
