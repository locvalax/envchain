package source

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// AzureKeyVaultClient defines the interface for Azure Key Vault operations.
type AzureKeyVaultClient interface {
	GetSecret(ctx context.Context, name, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
	NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse]
}

type azureKeyVaultSource struct {
	client AzureKeyVaultClient
	prefix string
}

// NewAzureKeyVaultSource creates a new Source backed by Azure Key Vault.
// prefix is an optional key prefix to filter secrets (e.g. "myapp-").
func NewAzureKeyVaultSource(client AzureKeyVaultClient, prefix string) Source {
	return &azureKeyVaultSource{
		client: client,
		prefix: prefix,
	}
}

func (s *azureKeyVaultSource) Get(key string) (string, bool) {
	ctx := context.Background()
	name := s.secretName(key)
	resp, err := s.client.GetSecret(ctx, name, "", nil)
	if err != nil {
		if isAzureNotFoundError(err) {
			return "", false
		}
		return "", false
	}
	if resp.Value == nil {
		return "", false
	}
	return *resp.Value, true
}

func (s *azureKeyVaultSource) Keys() []string {
	ctx := context.Background()
	var keys []string
	pager := s.client.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			break
		}
		for _, item := range page.Value {
			if item.ID == nil {
				continue
			}
			name := item.ID.Name()
			if s.prefix == "" || strings.HasPrefix(name, s.prefix) {
				keys = append(keys, s.envKey(name))
			}
		}
	}
	return keys
}

func (s *azureKeyVaultSource) secretName(key string) string {
	name := strings.ToLower(strings.ReplaceAll(key, "_", "-"))
	if s.prefix != "" {
		return fmt.Sprintf("%s%s", s.prefix, name)
	}
	return name
}

func (s *azureKeyVaultSource) envKey(name string) string {
	trimmed := strings.TrimPrefix(name, s.prefix)
	return strings.ToUpper(strings.ReplaceAll(trimmed, "-", "_"))
}

func isAzureNotFoundError(err error) bool {
	var respErr *azcore.ResponseError
	if ok := errors.As(err, &respErr); ok {
		return respErr.StatusCode == 404
	}
	return false
}
