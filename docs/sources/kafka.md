# Kafka Source

The Kafka source resolves environment variables from a Kafka topic. Each message in the topic is treated as a key-value pair where the message key is the environment variable name and the message value is the environment variable value.

## Use Case

Useful for dynamic configuration delivery in event-driven architectures where configuration changes are streamed via Kafka topics.

## Usage

```go
import (
    "github.com/yourorg/envchain/pkg/source"
)

// Implement the KafkaConsumer interface with your Kafka client (e.g., confluent-kafka-go, sarama)
consumer := myKafkaConsumerImpl{
    brokers: []string{"localhost:9092"},
}

src := source.NewKafkaSource(consumer, "config-topic", "APP_")
```

## Parameters

| Parameter  | Type             | Description                                                  |
|------------|------------------|--------------------------------------------------------------|
| `consumer` | `KafkaConsumer`  | Interface implementation backed by your Kafka client library |
| `topic`    | `string`         | Kafka topic name to read configuration from                  |
| `prefix`   | `string`         | Optional prefix stripped from message keys before resolution |

## KafkaConsumer Interface

You must provide an implementation of the `KafkaConsumer` interface:

```go
type KafkaConsumer interface {
    ReadMessage(ctx context.Context, topic, key string) (string, error)
    ListKeys(ctx context.Context, topic string) ([]string, error)
}
```

## Prefix Filtering

When a prefix is set, only message keys with that prefix are returned by `Keys()`. The prefix is prepended when reading a specific key:

```go
// Message key in Kafka: "APP_DB_PASSWORD"
// Resolved as:          "DB_PASSWORD"
src := source.NewKafkaSource(consumer, "config-topic", "APP_")
val, err := src.Get(ctx, "DB_PASSWORD") // reads key "APP_DB_PASSWORD"
```

## Caching

Values are cached in-memory after the first successful read to reduce Kafka broker load. The cache is held for the lifetime of the source instance.

## Notes

- This source does not maintain a persistent Kafka consumer group; it performs point-in-time reads.
- For production use, implement `KafkaConsumer` using a client such as [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go) or [sarama](https://github.com/IBM/sarama).
- Ensure your Kafka topic uses log compaction to retain the latest value for each key.
