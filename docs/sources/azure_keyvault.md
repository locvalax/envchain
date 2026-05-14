# Azure Key Vault Source

The Azure Key Vault source allows envchain to resolve environment variables from secrets stored in [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault).

## Usage

```go
import (
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
    "github.com/yourorg/envchain/pkg/source"
    "github.com/yourorg/envchain/pkg/chain"
)

cred, _ := azidentity.NewDefaultAzureCredential(nil)
client, _ := azsecrets.NewClient("https://my-vault.vault.azure.net/", cred, nil)

// No prefix — secret named "db-password" maps to env key "DB_PASSWORD"
kvSrc := source.NewAzureKeyVaultSource(client, "")

// With prefix — secret "myapp-api-key" maps to env key "API_KEY"
kvSrc := source.NewAzureKeyVaultSource(client, "myapp-")

c := chain.New(kvSrc, source.NewEnvSource(""))
val, ok := c.Resolve("DB_PASSWORD")
```

## Key Name Mapping

Azure Key Vault secret names are mapped to/from environment variable names:

| Env Key       | Key Vault Secret Name    |
|---------------|--------------------------|
| `DB_PASSWORD` | `db-password`            |
| `API_KEY`     | `myapp-api-key` (prefix) |

- Underscores (`_`) are converted to hyphens (`-`) when looking up secrets.
- Secret names are lowercased.
- On retrieval, hyphens are converted back to underscores and uppercased.

## Authentication

Uses the [Azure SDK DefaultAzureCredential](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#NewDefaultAzureCredential), which supports:

- Environment variables (`AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`)
- Managed Identity
- Azure CLI (`az login`)
- Visual Studio Code credentials

## Integration Tests

Run integration tests with:

```bash
AZURE_KEYVAULT_URL=https://my-vault.vault.azure.net/ \
AZURE_KEYVAULT_KEY=MY_SECRET \
AZURE_KEYVAULT_VALUE=expectedvalue \
go test -tags integration ./pkg/source/...
```
