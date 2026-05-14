package source

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// KafkaConsumer is an interface for consuming messages from a Kafka topic.
type KafkaConsumer interface {
	ReadMessage(ctx context.Context, topic, key string) (string, error)
	ListKeys(ctx context.Context, topic string) ([]string, error)
}

// kafkaSource resolves environment variables from a Kafka topic.
// Each message key maps to an env var name; the message value is the env var value.
type kafkaSource struct {
	consumer KafkaConsumer
	topic    string
	prefix   string
	mu       sync.RWMutex
	cache    map[string]string
}

// NewKafkaSource creates a new source backed by a Kafka topic.
// The topic is treated as a key-value store where message keys are env var names.
// An optional prefix is stripped from message keys when resolving variable names.
func NewKafkaSource(consumer KafkaConsumer, topic, prefix string) Source {
	return &kafkaSource{
		consumer: consumer,
		topic:    topic,
		prefix:   prefix,
		cache:    make(map[string]string),
	}
}

func (k *kafkaSource) Get(ctx context.Context, key string) (string, error) {
	k.mu.RLock()
	if val, ok := k.cache[key]; ok {
		k.mu.RUnlock()
		return val, nil
	}
	k.mu.RUnlock()

	msgKey := key
	if k.prefix != "" {
		msgKey = k.prefix + key
	}

	val, err := k.consumer.ReadMessage(ctx, k.topic, msgKey)
	if err != nil {
		return "", fmt.Errorf("kafka: failed to read key %q from topic %q: %w", msgKey, k.topic, err)
	}

	k.mu.Lock()
	k.cache[key] = val
	k.mu.Unlock()

	return val, nil
}

func (k *kafkaSource) Keys(ctx context.Context) ([]string, error) {
	rawKeys, err := k.consumer.ListKeys(ctx, k.topic)
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to list keys from topic %q: %w", k.topic, err)
	}

	keys := make([]string, 0, len(rawKeys))
	for _, rk := range rawKeys {
		if k.prefix != "" {
			if len(rk) > len(k.prefix) && rk[:len(k.prefix)] == k.prefix {
				keys = append(keys, rk[len(k.prefix):])
			}
		} else {
			keys = append(keys, rk)
		}
	}
	return keys, nil
}

// KafkaJSONConsumer is a helper that wraps a map to simulate a Kafka topic
// backed by a JSON blob — useful for testing or static configuration scenarios.
type KafkaJSONConsumer struct {
	data map[string]map[string]string // topic -> key -> value
}

func NewKafkaJSONConsumer(topicData map[string]string, topic string) *KafkaJSONConsumer {
	return &KafkaJSONConsumer{
		data: map[string]map[string]string{topic: topicData},
	}
}

func (c *KafkaJSONConsumer) ReadMessage(_ context.Context, topic, key string) (string, error) {
	topicMap, ok := c.data[topic]
	if !ok {
		return "", fmt.Errorf("topic %q not found", topic)
	}
	val, ok := topicMap[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in topic %q", key, topic)
	}
	return val, nil
}

func (c *KafkaJSONConsumer) ListKeys(_ context.Context, topic string) ([]string, error) {
	topicMap, ok := c.data[topic]
	if !ok {
		return nil, fmt.Errorf("topic %q not found", topic)
	}
	keys := make([]string, 0, len(topicMap))
	for k := range topicMap {
		keys = append(keys, k)
	}
	return keys, nil
}

// ensure KafkaJSONConsumer satisfies KafkaConsumer
var _ KafkaConsumer = (*KafkaJSONConsumer)(nil)

// suppress unused import
var _ = json.Marshal
