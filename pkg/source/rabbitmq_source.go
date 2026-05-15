package source

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// RabbitMQClient defines the interface for consuming a single message from a queue.
type RabbitMQClient interface {
	Get(queue string) ([]byte, bool, error)
}

// rabbitmqSource resolves environment variables from a RabbitMQ queue.
// It consumes one message (JSON object) from the queue at construction time
// and serves keys from that snapshot.
type rabbitmqSource struct {
	mu     sync.RWMutex
	data   map[string]string
	prefix string
}

// RabbitMQOption configures a rabbitmqSource.
type RabbitMQOption func(*rabbitmqSource)

// WithRabbitMQPrefix sets a key prefix that is stripped before lookup.
func WithRabbitMQPrefix(prefix string) RabbitMQOption {
	return func(s *rabbitmqSource) {
		s.prefix = prefix
	}
}

// NewRabbitMQSource creates a Source backed by a single RabbitMQ message.
// The message must be a flat JSON object mapping string keys to string values.
// If no message is available within the deadline the source is empty.
func NewRabbitMQSource(client RabbitMQClient, queue string, opts ...RabbitMQOption) (Source, error) {
	s := &rabbitmqSource{data: make(map[string]string)}
	for _, o := range opts {
		o(s)
	}

	body, ok, err := client.Get(queue)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: get from queue %q: %w", queue, err)
	}
	if ok && len(body) > 0 {
		if err := json.Unmarshal(body, &s.data); err != nil {
			return nil, fmt.Errorf("rabbitmq: unmarshal message: %w", err)
		}
	}
	return s, nil
}

func (s *rabbitmqSource) Get(_ context.Context, key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + key
	}
	v, ok := s.data[lookup]
	return v, ok, nil
}

func (s *rabbitmqSource) Keys(_ context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	prefixLen := len(s.prefix)
	for k := range s.data {
		if s.prefix == "" {
			keys = append(keys, k)
		} else if len(k) > prefixLen && k[:prefixLen] == s.prefix {
			keys = append(keys, k[prefixLen:])
		}
	}
	return keys, nil
}

// ensure compile-time interface satisfaction
var _ Source = (*rabbitmqSource)(nil)
var _ = time.Second // imported for future timeout use
