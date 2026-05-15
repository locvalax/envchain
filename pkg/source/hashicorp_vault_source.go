package source

import (
	"context"
	"fmt"
	"strings"
)

// HashiCorpVaultClient defines the interface for interacting with HashiCorp Vault.
type HashiCorpVaultClient interface {
	Read(ctx context.Context, path string) (map[string]interface{}, error)
	List(ctx context.Context, path string) ([]string, error)
}

type hashiCorpVaultSource struct {
	client HashiCorpVaultClient
	mountPath string
	secretPath string
	prefix string
}

// NewHashiCorpVaultSource creates a new source backed by HashiCorp Vault KV v2.
// mountPath is the KV mount (e.g. "secret"), secretPath is the path to the secret.
// prefix, if non-empty, is stripped from requested keys before lookup.
func NewHashiCorpVaultSource(client HashiCorpVaultClient, mountPath, secretPath, prefix string) Source {
	return &hashiCorpVaultSource{
		client:     client,
		mountPath:  strings.Trim(mountPath, "/"),
		secretPath: strings.Trim(secretPath, "/"),
		prefix:     prefix,
	}
}

func (s *hashiCorpVaultSource) fullPath() string {
	return fmt.Sprintf("%s/data/%s", s.mountPath, s.secretPath)
}

func (s *hashiCorpVaultSource) resolveKey(key string) string {
	if s.prefix != "" {
		return strings.TrimPrefix(key, s.prefix)
	}
	return key
}

func (s *hashiCorpVaultSource) Get(ctx context.Context, key string) (string, bool, error) {
	data, err := s.client.Read(ctx, s.fullPath())
	if err != nil {
		return "", false, fmt.Errorf("hashicorp vault read: %w", err)
	}
	if data == nil {
		return "", false, nil
	}
	resolvedKey := s.resolveKey(key)
	if val, ok := data[resolvedKey]; ok {
		return fmt.Sprintf("%v", val), true, nil
	}
	return "", false, nil
}

func (s *hashiCorpVaultSource) Keys(ctx context.Context) ([]string, error) {
	data, err := s.client.Read(ctx, s.fullPath())
	if err != nil {
		return nil, fmt.Errorf("hashicorp vault read keys: %w", err)
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		if s.prefix != "" {
			keys = append(keys, s.prefix+k)
		} else {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (s *hashiCorpVaultSource) Name() string {
	return fmt.Sprintf("hashicorp-vault:%s/%s", s.mountPath, s.secretPath)
}
