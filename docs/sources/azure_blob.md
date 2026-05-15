# Azure Blob Storage Source

The `azure_blob` source reads a JSON-encoded key/value map from a blob in
Azure Blob Storage. This is useful when your configuration is stored as a
flat JSON file uploaded to a storage account.

## Format

The blob must contain a flat JSON object where every value is a string:

```json
{
  "DB_HOST": "db.example.com",
  "DB_PORT": "5432",
  "API_KEY": "secret"
}
```

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
)

// Implement AzureBlobClient or use the SDK wrapper shown in the
// integration test.
client := myAzureBlobClient

src, err := source.NewAzureBlobSource(
    client,
    "my-container",  // container name
    "config.json",   // blob name
    "APP_",          // optional key prefix (empty string = no prefix)
)
if err != nil {
    log.Fatal(err)
}
```

## Options

| Parameter   | Type     | Description                                              |
|-------------|----------|----------------------------------------------------------|
| `client`    | interface| `AzureBlobClient` implementation                         |
| `container` | string   | Azure Blob Storage container name                        |
| `blob`      | string   | Blob (file) name within the container                    |
| `prefix`    | string   | Key prefix to prepend when looking up values (optional)  |

## Integration Test

Set the following environment variables and run with `-tags integration`:

```sh
export AZURE_STORAGE_CONNECTION_STRING="DefaultEndpointsProtocol=https;..."
export AZURE_BLOB_CONTAINER="my-container"
export AZURE_BLOB_NAME="config.json"

go test ./pkg/source/... -tags integration -run TestAzureBlobSource_Integration
```

## Notes

- The blob is loaded once at construction time. To refresh, create a new source.
- When a prefix is set, only keys that start with the prefix are returned by `Keys()`.
- The `Name()` method returns `azure-blob://<container>/<blob>`.
