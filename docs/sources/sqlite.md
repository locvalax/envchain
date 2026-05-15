# SQLite Source

The `SQLiteSource` reads environment variables from a SQLite database table.
This is useful for local development, embedded configurations, or offline
environments where a full database server is unnecessary.

## Table Schema

By default, the source expects a table named `env` with `key` and `value`
columns:

```sql
CREATE TABLE env (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

## Usage

```go
import "github.com/yourorg/envchain/pkg/source"

src, err := source.NewSQLiteSource("./config.db")
if err != nil {
    log.Fatal(err)
}

val, ok := src.Get("DATABASE_URL")
```

## Custom Table and Columns

Use `WithSQLiteTable` and `WithSQLiteColumns` to adapt the source to an
existing schema:

```go
src, err := source.NewSQLiteSource("./app.db",
    source.WithSQLiteTable("config"),
    source.WithSQLiteColumns("name", "val"),
)
```

## Options

| Option | Default | Description |
|---|---|---|
| `WithSQLiteTable(name)` | `"env"` | Table to query |
| `WithSQLiteColumns(key, val)` | `"key"`, `"value"` | Column names for key and value |

## Chain Example

```go
chain := chain.New(
    source.NewEnvSource(),
    sqliteSrc,
    source.NewMapSource(defaults),
)
```

The SQLite source sits below the process environment so that live overrides
always take precedence, while the database provides persisted defaults.
