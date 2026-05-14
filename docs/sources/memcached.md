# Memcached Source

The Memcached source reads environment variable values from a [Memcached](https://memcached.org/) cache.

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
)

client := // your *memcache.Client (e.g. bradfitz/gomemcache)

s := source.NewMemcachedSource(client, "myapp/", []string{
    "DB_PASSWORD",
    "API_KEY",
    "JWT_SECRET",
})
```

## Parameters

| Parameter | Type              | Description                                                    |
|-----------|-------------------|----------------------------------------------------------------|
| `client`  | `MemcachedClient` | Any client implementing `Get(key) ([]byte, error)` and `GetMulti`. |
| `prefix`  | `string`          | Optional prefix prepended to every key before lookup.          |
| `keys`    | `[]string`        | Declared keys (without prefix) this source can provide.        |

## Key Prefixing

When a `prefix` is set, the source automatically prepends it during lookups:

```
prefix = "prod/"
Requested key: DB_PASSWORD
Actual Memcached key: prod/DB_PASSWORD
```

## Client Interface

The source depends on the `MemcachedClient` interface rather than a concrete
library, making it easy to plug in any Memcached client:

```go
type MemcachedClient interface {
    Get(key string) ([]byte, error)
    GetMulti(keys []string) (map[string][]byte, error)
}
```

The popular [bradfitz/gomemcache](https://github.com/bradfitz/gomemcache) client
satisfies this interface out of the box.

## Cache Miss Behaviour

A cache miss is treated as a **missing key** (returns `ok = false, err = nil`),
allowing the chain to fall through to the next source. Any other error is
propagated up as a hard failure.

## Notes

- Values are trimmed of leading/trailing whitespace before being returned.
- The `Keys()` method returns the static list provided at construction time;
  Memcached does not natively support key enumeration.
