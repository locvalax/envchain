package source

import (
	"context"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

// ZookeeperClient defines the interface for interacting with ZooKeeper.
type ZookeeperClient interface {
	Get(path string) ([]byte, *zk.Stat, error)
	Children(path string) ([]string, *zk.Stat, error)
}

type zookeeperSource struct {
	client ZookeeperClient
	prefix string
}

// NewZookeeperSource creates a new Source backed by Apache ZooKeeper.
// The prefix is prepended to all key lookups as a ZooKeeper path segment.
// For example, prefix "/config" and key "DB_HOST" resolves to "/config/DB_HOST".
func NewZookeeperSource(servers []string, prefix string, timeout time.Duration) (Source, error) {
	conn, _, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, err
	}
	return &zookeeperSource{
		client: conn,
		prefix: strings.TrimRight(prefix, "/"),
	}, nil
}

// NewZookeeperSourceWithClient creates a ZooKeeper source with an injected client (useful for testing).
func NewZookeeperSourceWithClient(client ZookeeperClient, prefix string) Source {
	return &zookeeperSource{
		client: client,
		prefix: strings.TrimRight(prefix, "/"),
	}
}

func (z *zookeeperSource) path(key string) string {
	if z.prefix == "" {
		return "/" + key
	}
	return z.prefix + "/" + key
}

func (z *zookeeperSource) Get(_ context.Context, key string) (string, bool, error) {
	data, _, err := z.client.Get(z.path(key))
	if err != nil {
		if err == zk.ErrNoNode {
			return "", false, nil
		}
		return "", false, err
	}
	return string(data), true, nil
}

func (z *zookeeperSource) Keys(_ context.Context) ([]string, error) {
	basePath := z.prefix
	if basePath == "" {
		basePath = "/"
	}
	children, _, err := z.client.Children(basePath)
	if err != nil {
		if err == zk.ErrNoNode {
			return []string{}, nil
		}
		return nil, err
	}
	return children, nil
}
