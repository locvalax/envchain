# SQS Source

The SQS source reads environment variables from messages in an [AWS Simple Queue Service (SQS)](https://aws.amazon.com/sqs/) queue. Each message body must be a JSON object mapping string keys to string values.

## Usage

```go
import (
    "context"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/yourorg/envchain/pkg/source"
    "github.com/yourorg/envchain/pkg/chain"
)

func main() {
    cfg, _ := config.LoadDefaultConfig(context.Background())
    client := sqs.NewFromConfig(cfg)

    sqsSrc, err := source.NewSQSSource(client, "my-config-queue")
    if err != nil {
        log.Fatal(err)
    }

    c := chain.New(sqsSrc)
    val, ok := c.Resolve("DB_PASSWORD")
}
```

## Message Format

Each SQS message body must be a flat JSON object:

```json
{"DB_HOST": "db.internal", "DB_PASSWORD": "s3cr3t"}
```

Up to **10 messages** are consumed per source initialisation (the SQS maximum per `ReceiveMessage` call). Messages are **not deleted** from the queue after reading — this source is intended for read-only config delivery patterns or FIFO queues with external consumers.

## Options

| Option | Description |
|---|---|
| `WithSQSPrefix(prefix)` | Only include keys that start with `prefix`; the prefix is stripped from the stored key name. |

## Prefix Filtering

```go
// Only keys starting with "APP_" are loaded; stored without the prefix.
sqsSrc, _ := source.NewSQSSource(client, "my-queue", source.WithSQSPrefix("APP_"))
// Message key "APP_SECRET" is accessible as "SECRET"
```

## IAM Permissions

The following IAM actions are required:

- `sqs:GetQueueUrl`
- `sqs:ReceiveMessage`

## Integration Test

```bash
export ENVCHAIN_TEST_SQS_QUEUE=my-config-queue
export ENVCHAIN_TEST_SQS_EXPECTED_KEY=DB_HOST
go test -tags=integration ./pkg/source/ -run TestSQSSource_Integration
```
