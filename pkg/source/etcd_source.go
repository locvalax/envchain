package source

import (
	"context"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdClient defines the interface for etcd operations used by EtcdSource.
type EtcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

type etcdSource struct {
	client EtcdClient
	prefix string
	timeout time.Duration
}

// NewEtcdSource creates a Source backed by an etcd cluster.
// prefix is prepended to all key lookups (e.g. "/myapp/").
func NewEtcdSource(client EtcdClient, prefix string, timeout time.Duration) Source {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &etcdSource{
		client:  client,
		prefix:  prefix,
		timeout: timeout,
	}
}

func (e *etcdSource) Get(key string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	fullKey := e.prefix + key
	resp, err := e.client.Get(ctx, fullKey)
	if err != nil || resp == nil || len(resp.Kvs) == 0 {
		return "", false
	}
	return string(resp.Kvs[0].Value), true
}

func (e *etcdSource) Keys() []string {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.client.Get(ctx, e.prefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil || resp == nil {
		return nil
	}

	keys := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		k := strings.TrimPrefix(string(kv.Key), e.prefix)
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func (e *etcdSource) Name() string {
	return "etcd"
}
