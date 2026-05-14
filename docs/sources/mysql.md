# MySQL Source

The MySQL source retrieves environment variables from a MySQL database table.

## Configuration

| Parameter  | Description                                      |
|------------|--------------------------------------------------|
| `dsn`      | MySQL Data Source Name (e.g. `user:pass@tcp(host:3306)/dbname`) |
| `table`    | Table name where secrets are stored              |
| `keyCol`   | Column name for the environment variable key     |
| `valueCol` | Column name for the environment variable value   |

## Schema Example

```sql
CREATE TABLE env_vars (
    `key`   VARCHAR(255) NOT NULL PRIMARY KEY,
    `value` TEXT         NOT NULL
);

INSERT INTO env_vars (`key`, `value`) VALUES ('API_KEY', 'abc123');
INSERT INTO env_vars (`key`, `value`) VALUES ('DB_PASS', 's3cr3t');
```

## Usage

```go
import "github.com/yourorg/envchain/pkg/source"

src, err := source.NewMySQLSource(
    "user:password@tcp(localhost:3306)/mydb",
    "env_vars",
    "key",
    "value",
)
if err != nil {
    log.Fatal(err)
}
```

## Chaining with Other Sources

```go
chain := chain.New(
    source.NewEnvSource(""),  // highest priority
    src,                      // MySQL fallback
)

val, ok, err := chain.Resolve(ctx, "API_KEY")
```

## Notes

- Requires the `github.com/go-sql-driver/mysql` driver.
- The source performs a `LIMIT 1` query per `Get` call; ensure the key column is indexed.
- `Keys()` performs a full table scan — use with care on large tables.
