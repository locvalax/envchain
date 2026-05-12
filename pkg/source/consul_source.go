package source

import (
	"context"
	"fmt"
	"strings"
)

// ConsulKVClient defines the interface for interacting with Consul KV store.
type ConsulKVClient interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Keys(ctx context.Context, prefix string) ([]string, error)
}

type consulSource struct {
	client ConsulKVClient
	prefix string
	ctx    context.Context
}

// NewConsulSource creates a new Source backed by HashiCorp Consul KV.
// prefix is an optional path prefix (e.g. "myapp/prod/").
func NewConsulSource(ctx context.Context, client ConsulKVClient, prefix string) Source {
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return &consulSource{
		client: client,
		prefix: prefix,
		ctx:    ctx,
	}
}

func (c *consulSource) Get(key string) (string, bool, error) {
	fullKey := c.prefix + key
	val, found, err := c.client.Get(c.ctx, fullKey)
	if err != nil {
		return "", false, fmt.Errorf("consul: get %q: %w", fullKey, err)
	}
	return val, found, nil
}

func (c *consulSource) Keys() ([]string, error) {
	rawKeys, err := c.client.Keys(c.ctx, c.prefix)
	if err != nil {
		return nil, fmt.Errorf("consul: keys with prefix %q: %w", c.prefix, err)
	}
	keys := make([]string, 0, len(rawKeys))
	for _, k := range rawKeys {
		keys = append(keys, strings.TrimPrefix(k, c.prefix))
	}
	return keys, nil
}
