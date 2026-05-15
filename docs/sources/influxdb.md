# InfluxDB Source

The InfluxDB source reads environment variables from an [InfluxDB v2](https://www.influxdata.com/products/influxdb/) bucket.

Each data point must use the measurement name `envchain`, a tag `key` containing the variable name, and a field `value` containing the variable value.

## Usage

```go
import (
    "github.com/your-org/envchain/pkg/source"
    "github.com/your-org/envchain/pkg/chain"
)

func main() {
    s, err := source.NewInfluxDBSource(
        "http://localhost:8086",  // InfluxDB server URL
        "my-token",               // API token
        "my-org",                 // organisation
        "my-bucket",              // bucket
        source.WithInfluxDBPrefix("APP_"),
    )
    if err != nil {
        log.Fatal(err)
    }

    c := chain.New(s)
    val, ok, err := c.Resolve(context.Background(), "DB_HOST")
}
```

## Options

| Option | Description |
|---|---|
| `WithInfluxDBPrefix(p)` | Strip prefix `p` from tag keys when looking up or listing variables. |
| `WithInfluxDBKnownKeys(keys)` | Return a static list from `Keys()` instead of querying InfluxDB. |

## Writing a value

```bash
# Using the influx CLI
influx write \
  --bucket my-bucket \
  --org my-org \
  --token my-token \
  'envchain,key=DB_HOST value="localhost"'
```

## Notes

- The source queries the **last** value written within the past hour (`range(start: -1h)`). Adjust the Flux query via a fork if a longer window is required.
- Authentication uses the standard InfluxDB v2 token mechanism.
- For production use, store the token in a higher-priority source (e.g. AWS SSM) and inject it before constructing this source.
