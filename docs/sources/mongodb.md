# MongoDB Source

The MongoDB source fetches environment variable values from a MongoDB collection.
Each document in the collection represents a single key-value pair, with
configurable field names for the key and value.

## Document Schema

```json
{ "key": "DB_PASSWORD", "value": "s3cr3t" }
```

The field names (`key` and `value` by default) are fully configurable.

## Usage

```go
import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/yourorg/envchain/pkg/source"
)

client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    log.Fatal(err)
}

src := source.NewMongoDBSourceFromClient(
    client,
    "myapp",      // database name
    "secrets",    // collection name
    "key",        // field used as the env var name
    "value",      // field used as the env var value
)

val, err := src.Get(ctx, "DB_PASSWORD")
```

## Integration with Chain

```go
chain := chain.New(
    source.NewEnvSource(""),
    source.NewMongoDBSourceFromClient(client, "myapp", "secrets", "key", "value"),
)
```

## Integration Tests

Set `MONGODB_URI` and run:

```bash
go test -tags integration ./pkg/source/... -run TestMongoDBSource_Integration
```

## Notes

- The source returns `ErrKeyNotFound` when no document matches the requested key.
- Values must be stored as BSON strings; non-string values produce an error.
- Use `Keys()` to enumerate all keys (performs a projection query for efficiency).
