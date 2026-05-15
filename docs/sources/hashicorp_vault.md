# HashiCorp Vault Source

The HashiCorp Vault source reads secrets from a [HashiCorp Vault](https://www.vaultproject.io/) KV v2 secrets engine.

## Usage

```go
import "github.com/yourorg/envchain/pkg/source"

// Provide your own client that satisfies HashiCorpVaultClient.
vaultSrc := source.NewHashiCorpVaultSource(client, "secret", "myapp/config", "")
```

## Constructor

```go
func NewHashiCorpVaultSource(
    client     HashiCorpVaultClient,
    mountPath  string, // KV v2 mount, e.g. "secret"
    secretPath string, // path within mount, e.g. "myapp/config"
    prefix     string, // optional key prefix to strip
) Source
```

## Interface

Implement `HashiCorpVaultClient` to inject a custom Vault client:

```go
type HashiCorpVaultClient interface {
    Read(ctx context.Context, path string) (map[string]interface{}, error)
    List(ctx context.Context, path string) ([]string, error)
}
```

## Key Prefix

When `prefix` is set, the source strips it before looking up the key in Vault:

```go
// Vault secret contains: {"API_KEY": "abc"}
src := source.NewHashiCorpVaultSource(client, "secret", "app", "APP_")
val, ok, _ := src.Get(ctx, "APP_API_KEY") // looks up "API_KEY" in Vault
```

## Integration Test

Set the following environment variables and run with `-tags integration`:

| Variable | Description |
|---|---|
| `VAULT_ADDR` | Vault server address, e.g. `http://localhost:8200` |
| `VAULT_TOKEN` | Vault root or policy token |

```bash
go test -tags integration ./pkg/source/... -run TestHashiCorpVaultSource_Integration
```

## Notes

- Reads from the KV v2 `data` sub-path automatically (`<mount>/data/<secretPath>`).
- Returns `(value, false, nil)` when the secret path does not exist.
- All values are coerced to strings via `fmt.Sprintf("%v", val)`.
