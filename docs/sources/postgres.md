# PostgreSQL Source

The `PostgresSource` reads environment variables from a PostgreSQL table, making it easy to manage configuration centrally in an existing relational database.

## Table Schema

By default, the source expects a table named `env_vars` with `key` and `value` columns:

```sql
CREATE TABLE env_vars (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO env_vars (key, value) VALUES ('DB_PASSWORD', 's3cr3t');
```

## Usage

```go
import "github.com/yourorg/envchain/pkg/source"

src, err := source.NewPostgresSource(
    "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
    source.PostgresOptions{},
)
if err != nil {
    log.Fatal(err)
}
```

## Custom Table and Columns

If your schema differs, configure the table and column names via `PostgresOptions`:

```go
src, err := source.NewPostgresSource(dsn, source.PostgresOptions{
    Table:    "config",
    KeyCol:   "name",
    ValueCol: "val",
})
```

## Options

| Field      | Default    | Description                        |
|------------|------------|------------------------------------|
| `Table`    | `env_vars` | Table name to query                |
| `KeyCol`   | `key`      | Column name used as the key        |
| `ValueCol` | `value`    | Column name used as the value      |

## Chaining with Other Sources

```go
chain := chain.New(
    src,                          // Postgres (highest priority)
    source.NewEnvSource(""),      // OS environment (fallback)
)

val, ok, err := chain.Resolve(ctx, "DB_PASSWORD")
```

## Notes

- Requires the `github.com/lib/pq` driver to be imported in your application.
- The source performs a single `SELECT` per `Get` call; consider adding an index on the key column for production use.
