# DynamoDB Source

The DynamoDB source retrieves environment variables stored as items in an AWS DynamoDB table.

## Table Schema

Each item in the table should have at minimum:

| Attribute | Type   | Description                        |
|-----------|--------|------------------------------------|
| `key`     | String | The environment variable name      |
| `value`   | String | The environment variable value     |

The attribute names are configurable via `NewDynamoDBSource`.

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

cfg, err := config.LoadDefaultConfig(context.Background())
if err != nil {
    log.Fatal(err)
}

client := dynamodb.NewFromConfig(cfg)

src := source.NewDynamoDBSource(
    client,
    "my-env-table", // table name
    "key",          // partition key attribute
    "value",        // value attribute
    "prod/",        // optional key prefix
)
```

## Options

| Parameter   | Description                                                                 |
|-------------|-----------------------------------------------------------------------------|
| `client`    | An AWS DynamoDB client implementing `DynamoDBClient`                        |
| `table`     | The DynamoDB table name                                                     |
| `keyAttr`   | The attribute name used as the partition key (e.g. `"key"`)                 |
| `valueAttr` | The attribute name storing the variable value (e.g. `"value"`)             |
| `prefix`    | Optional prefix prepended to all key lookups and stripped from `Keys()`    |

## Permissions

The IAM role or user must have the following permissions:

```json
{
  "Effect": "Allow",
  "Action": [
    "dynamodb:GetItem",
    "dynamodb:Scan"
  ],
  "Resource": "arn:aws:dynamodb:*:*:table/my-env-table"
}
```

## Notes

- Only `String` (`S`) attribute values are supported for the value field.
- The `Keys()` method performs a full table `Scan` and filters by prefix; use with caution on large tables.
- For large tables, consider adding a GSI or using the SSM Parameter Store source instead.
