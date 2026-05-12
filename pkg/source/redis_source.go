package source

import (
	"context"
	"fmt"
	"strings"
)

// RedisClient defines the minimal interface needed to interact with Redis.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// redisSource resolves environment variables from a Redis instance.
type redisSource struct {
	client RedisClient
	prefix string
	ctx    context.Context
}

// NewRedisSource creates a new Source backed by Redis.
// The prefix is prepended to all key lookups (e.g. "myapp:" → "myapp:DB_HOST").
func NewRedisSource(ctx context.Context, client RedisClient, prefix string) Source {
	return &redisSource{
		client: client,
		prefix: prefix,
		ctx:    ctx,
	}
}

// Get retrieves a value from Redis by key, applying the configured prefix.
func (r *redisSource) Get(key string) (string, bool) {
	redisKey := r.buildKey(key)
	val, err := r.client.Get(r.ctx, redisKey)
	if err != nil {
		return "", false
	}
	return val, true
}

// Keys returns all keys available in Redis matching the prefix pattern,
// with the prefix stripped from the returned key names.
func (r *redisSource) Keys() []string {
	pattern := r.prefix + "*"
	redisKeys, err := r.client.Keys(r.ctx, pattern)
	if err != nil {
		return nil
	}
	keys := make([]string, 0, len(redisKeys))
	for _, rk := range redisKeys {
		keys = append(keys, strings.TrimPrefix(rk, r.prefix))
	}
	return keys
}

func (r *redisSource) buildKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s%s", r.prefix, key)
}
