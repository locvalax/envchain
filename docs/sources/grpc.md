# gRPC Source

The `grpc` source fetches environment variable values from a remote gRPC-based secrets service.

## Overview

This source is useful when you have an internal secrets microservice that exposes a key/value API over gRPC. It supports optional key prefixing and implements the standard `Source` interface.

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
)

src, err := source.NewGRPCSource(source.GRPCClientConfig{
    Address: "localhost:50051",
    Prefix:  "prod/",
})
if err != nil {
    log.Fatal(err)
}
```

## Custom Client

To use your own generated proto client, implement the `GRPCClient` interface:

```go
type GRPCClient interface {
    GetSecret(ctx context.Context, key string) (string, error)
    ListKeys(ctx context.Context) ([]string, error)
}
```

Then inject it directly:

```go
src := source.NewGRPCSourceWithClient(myProtoClient, "prod/")
```

## Configuration

| Field     | Type   | Description                                      |
|-----------|--------|--------------------------------------------------|
| `Address` | string | gRPC server address (e.g. `localhost:50051`)      |
| `Prefix`  | string | Optional prefix prepended to all key lookups     |

## Behaviour

- Returns `(value, true)` when the key is found.
- Returns `("", false)` when the server responds with `codes.NotFound` or any other error.
- `Keys()` returns all available keys via `ListKeys`; returns `nil` on error.
- `Name()` returns `"grpc"`.

## Notes

- The default client uses `insecure` credentials. For production use, supply a client configured with TLS.
- Prefix is prepended **before** the lookup, so `Get("API_KEY")` with prefix `prod/` queries `prod/API_KEY`.
