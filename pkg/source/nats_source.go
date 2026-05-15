package source

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSKVClient abstracts the NATS JetStream KeyValue operations.
type NATSKVClient interface {
	Get(key string) (nats.KeyValueEntry, error)
	Keys() ([]string, error)
}

type natsSource struct {
	kv     NATSKVClient
	prefix string
}

// NewNATSSource creates a Source backed by a NATS JetStream KeyValue bucket.
// prefix is an optional key prefix to strip when exposing keys.
func NewNATSSource(url, bucket, prefix string) (Source, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats: connect: %w", err)
	}

	js, err := nc.JetStream(nats.MaxWait(5 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("nats: jetstream: %w", err)
	}

	kv, err := js.KeyValue(bucket)
	if err != nil {
		return nil, fmt.Errorf("nats: kv bucket %q: %w", bucket, err)
	}

	return &natsSource{kv: kv, prefix: prefix}, nil
}

// NewNATSSourceWithClient creates a Source using a pre-built NATSKVClient (useful for testing).
func NewNATSSourceWithClient(kv NATSKVClient, prefix string) Source {
	return &natsSource{kv: kv, prefix: prefix}
}

func (s *natsSource) Get(_ context.Context, key string) (string, bool, error) {
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + "." + key
	}

	entry, err := s.kv.Get(lookup)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			return "", false, nil
		}
		return "", false, fmt.Errorf("nats: get %q: %w", lookup, err)
	}

	// Support both plain string values and JSON-encoded strings.
	raw := string(entry.Value())
	var decoded string
	if jsonErr := json.Unmarshal(entry.Value(), &decoded); jsonErr == nil {
		return decoded, true, nil
	}
	return raw, true, nil
}

func (s *natsSource) Keys(_ context.Context) ([]string, error) {
	all, err := s.kv.Keys()
	if err != nil {
		if err == nats.ErrNoKeysFound {
			return nil, nil
		}
		return nil, fmt.Errorf("nats: keys: %w", err)
	}

	result := make([]string, 0, len(all))
	for _, k := range all {
		if s.prefix != "" {
			pfx := s.prefix + "."
			if len(k) > len(pfx) && k[:len(pfx)] == pfx {
				result = append(result, k[len(pfx):])
			}
			continue
		}
		result = append(result, k)
	}
	return result, nil
}
