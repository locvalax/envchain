# Elasticsearch Source

The Elasticsearch source reads environment variables from documents stored in an Elasticsearch index.

Each document **ID** maps to a key name, and the document must contain a `"value"` field holding the string value.

## Document format

```json
{
  "value": "my-secret-value"
}
```

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
    "github.com/yourorg/envchain/pkg/chain"
)

client := source.NewHTTPElasticsearchClient("http://localhost:9200")
esSource := source.NewElasticsearchSource(client, "envchain")

c := chain.New(esSource)
val, ok := c.Resolve(ctx, "DB_HOST")
```

## Options

### `WithElasticsearchPrefix(prefix string)`

Prepends `prefix` to every document ID lookup and strips it from `Keys()` results.

```go
esSource := source.NewElasticsearchSource(
    client,
    "envchain",
    source.WithElasticsearchPrefix("prod/"),
)
// Resolving "DB_HOST" looks up document id "prod/DB_HOST"
```

## Custom client

Implement the `ElasticsearchClient` interface to use any Elasticsearch SDK:

```go
type ElasticsearchClient interface {
    Get(ctx context.Context, index, id string) (map[string]interface{}, error)
    Keys(ctx context.Context, index string) ([]string, error)
}
```

## Integration test

Run with:

```bash
ELASTICSEARCH_URL=http://localhost:9200 go test -tags integration ./pkg/source/...
```
