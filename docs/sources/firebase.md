# Firebase Realtime Database Source

The Firebase source reads environment variables from a [Firebase Realtime Database](https://firebase.google.com/docs/database) path.

## Configuration

| Parameter  | Description                                              | Required |
|------------|----------------------------------------------------------|----------|
| `client`   | A `*db.Client` from the Firebase Admin SDK              | Yes      |
| `basePath` | The database path where key/value pairs are stored       | Yes      |
| `prefix`   | Optional key prefix to filter or namespace lookups       | No       |

## Data Format

Keys and values must be stored as flat string pairs under the configured `basePath`:

```json
{
  "/envchain/prod": {
    "DB_HOST": "db.example.com",
    "API_KEY": "supersecret"
  }
}
```

## Usage

```go
import (
    firebase "firebase.google.com/go/v4"
    "github.com/yourorg/envchain/pkg/source"
)

app, _ := firebase.NewApp(ctx, &firebase.Config{
    DatabaseURL: "https://my-project-default-rtdb.firebaseio.com",
})
client, _ := app.Database(ctx)

src := source.NewFirebaseSource(client, "/envchain/prod", "")
val, ok, err := src.Get(ctx, "DB_HOST")
```

## Prefix Filtering

When a prefix is set, it is prepended to every key lookup:

```go
src := source.NewFirebaseSource(client, "/envchain", "prod_")
// Looks up /envchain/prod_DB_HOST
val, ok, err := src.Get(ctx, "DB_HOST")
```

## Authentication

Use Application Default Credentials or a service account key file via the `GOOGLE_APPLICATION_CREDENTIALS` environment variable.

## Integration Tests

Set the following environment variables and run with the `integration` build tag:

```bash
export FIREBASE_DATABASE_URL=https://my-project-default-rtdb.firebaseio.com
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceAccount.json
go test -tags integration ./pkg/source/...
```
