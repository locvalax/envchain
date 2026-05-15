# CockroachDB Source

The `CockroachDBSource` reads environment variables from a [CockroachDB](https://www.cockroachlabs.com/) table. Because CockroachDB speaks the PostgreSQL wire protocol, the source uses the standard `lib/pq` driver.

## Table Schema

```sql
CREATE TABLE env_vars (
  key   TEXT NOT NULL PRIMARY KEY,
  value TEXT NOT NULL
);
```

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
    "github.com/yourorg/envchain/pkg/chain"
)

func main() {
    crdb, err := source.NewCockroachDBSource(
        "postgresql://root@localhost:26257/defaultdb?sslmode=disable",
        "env_vars",
        "", // optional key prefix
    )
    if err != nil {
        log.Fatal(err)
    }

    c := chain.New(crdb)
    val, ok, err := c.Resolve(context.Background(), "DB_HOST")
}
```

## Options

| Parameter | Description |
|-----------|-------------|
| `dsn`     | CockroachDB connection string (PostgreSQL DSN format) |
| `table`   | Name of the table that holds key/value pairs |
| `prefix`  | Optional string prepended to every key lookup; stripped from `Keys()` results |

## Precedence

Like all envchain sources, `CockroachDBSource` can be composed with other sources via `chain.New`. Sources listed earlier have higher precedence.

```go
c := chain.New(envSource, crdbSource, dotenvSource)
```

## Notes

- The driver used is `lib/pq`; ensure it is imported (blank import `_ "github.com/lib/pq"`).
- For TLS-enabled clusters, set `sslmode=verify-full` and supply the appropriate certificates in the DSN.
