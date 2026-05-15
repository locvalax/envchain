# Google Cloud Storage (GCS) Source

The GCS source reads environment variables from a JSON object stored in a
Google Cloud Storage bucket. This is useful for sharing configuration across
multiple services or environments without managing individual secret entries.

## JSON Format

The object must contain a flat JSON map of string keys to string values:

```json
{
  "DB_HOST": "db.internal",
  "DB_PORT": "5432",
  "API_KEY": "supersecret"
}
```

## Usage

```go
import (
    "context"
    "log"

    "github.com/yourorg/envchain/pkg/chain"
    "github.com/yourorg/envchain/pkg/source"
)

func main() {
    ctx := context.Background()

    gcsSrc, err := source.NewGCSSource(ctx, "my-config-bucket", "prod/env.json", "")
    if err != nil {
        log.Fatalf("gcs source: %v", err)
    }

    c := chain.New(gcsSrc)
    if err := c.Inject(ctx); err != nil {
        log.Fatalf("inject: %v", err)
    }
}
```

## Key Prefix

If your JSON keys share a common prefix (e.g. `APP_`), you can pass it as the
`prefix` argument. The prefix will be prepended when looking up keys so callers
can use the unprefixed name:

```go
gcsSrc, err := source.NewGCSSource(ctx, "my-bucket", "env.json", "APP_")
// src.Get("DB_HOST") looks up "APP_DB_HOST" in the JSON object
```

## Authentication

The source uses Application Default Credentials (ADC). Ensure the environment
has credentials configured via one of:

- `GOOGLE_APPLICATION_CREDENTIALS` pointing to a service account key file
- Workload Identity (GKE)
- `gcloud auth application-default login` for local development

## Precedence

Like all envchain sources, GCS participates in the standard precedence chain.
Place it before lower-priority sources (e.g. `.env` files) and after
higher-priority sources (e.g. SSM or Secret Manager):

```go
c := chain.New(ssmSrc, gcsSrc, dotenvSrc)
```
