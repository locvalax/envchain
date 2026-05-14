package source

import (
	"context"
	"fmt"
	"strings"
)

// MemcachedClient defines the interface for interacting with Memcached.
type MemcachedClient interface {
	Get(key string) ([]byte, error)
	GetMulti(keys []string) (map[string][]byte, error)
}

type memcachedSource struct {
	client MemcachedClient
	prefix string
	keys   []string
}

// NewMemcachedSource creates a Source backed by Memcached.
// prefix is prepended to all key lookups. keys is the list of known keys
// (without prefix) that this source exposes.
func NewMemcachedSource(client MemcachedClient, prefix string, keys []string) Source {
	return &memcachedSource{
		client: client,
		prefix: prefix,
		keys:   keys,
	}
}

func (m *memcachedSource) Get(_ context.Context, key string) (string, bool, error) {
	lookup := key
	if m.prefix != "" {
		lookup = m.prefix + key
	}

	data, err := m.client.Get(lookup)
	if err != nil {
		if isMemcachedMiss(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("memcached get %q: %w", lookup, err)
	}

	return strings.TrimSpace(string(data)), true, nil
}

func (m *memcachedSource) Keys(_ context.Context) ([]string, error) {
	out := make([]string, len(m.keys))
	copy(out, m.keys)
	return out, nil
}

// isMemcachedMiss returns true for a cache-miss error.
// The canonical bradfitz/gomemcache library returns ErrCacheMiss; we match
// on the error message so we don't need to import the full client package.
func isMemcachedMiss(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "memcache: cache miss"
}
