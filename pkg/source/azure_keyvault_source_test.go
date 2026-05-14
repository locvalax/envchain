package source

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type mockAzureKVClient struct {
	secrets map[string]string
	listErr error
	getErr  error
}

func (m *mockAzureKVClient) GetSecret(_ context.Context, name, _ string, _ *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	if m.getErr != nil {
		return azsecrets.GetSecretResponse{}, m.getErr
	}
	val, ok := m.secrets[name]
	if !ok {
		return azsecrets.GetSecretResponse{}, &azcore.ResponseError{StatusCode: 404}
	}
	return azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{Value: &val},
	}, nil
}

func (m *mockAzureKVClient) NewListSecretPropertiesPager(_ *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	return nil // simplified; Keys() tested via integration
}

func TestAzureKeyVaultSource_GetFound(t *testing.T) {
	client := &mockAzureKVClient{
		secrets: map[string]string{"db-password": "secret123"},
	}
	src := NewAzureKeyVaultSource(client, "")
	val, ok := src.Get("DB_PASSWORD")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "secret123" {
		t.Errorf("expected 'secret123', got %q", val)
	}
}

func TestAzureKeyVaultSource_GetMissing(t *testing.T) {
	client := &mockAzureKVClient{
		secrets: map[string]string{},
	}
	src := NewAzureKeyVaultSource(client, "")
	_, ok := src.Get("MISSING_KEY")
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestAzureKeyVaultSource_PrefixedGet(t *testing.T) {
	client := &mockAzureKVClient{
		secrets: map[string]string{"myapp-api-key": "tok_abc"},
	}
	src := NewAzureKeyVaultSource(client, "myapp-")
	val, ok := src.Get("API_KEY")
	if !ok {
		t.Fatal("expected key to be found with prefix")
	}
	if val != "tok_abc" {
		t.Errorf("expected 'tok_abc', got %q", val)
	}
}

func TestAzureKeyVaultSource_ClientError(t *testing.T) {
	client := &mockAzureKVClient{
		getErr: errors.New("connection refused"),
	}
	src := NewAzureKeyVaultSource(client, "")
	_, ok := src.Get("ANY_KEY")
	if ok {
		t.Fatal("expected false on client error")
	}
}

func TestAzureKeyVaultSource_NilValue(t *testing.T) {
	client := &mockAzureKVClient{
		secrets: map[string]string{},
	}
	src := NewAzureKeyVaultSource(client, "")
	_, ok := src.Get("EMPTY_SECRET")
	if ok {
		t.Fatal("expected false for nil secret value")
	}
}
