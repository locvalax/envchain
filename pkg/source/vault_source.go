package source

import (
	"fmt"
	"sort"
	"strings"
)

// VaultSource is an in-memory source that simulates a secret vault.
// In a real implementation this would connect to HashiCorp Vault or similar.
type VaultSource struct {
	prefix  string
	secrets map[string]string
}

// NewVaultSource creates a VaultSource from a map of secrets.
// The prefix is stripped from keys when resolving, allowing namespaced secrets.
func NewVaultSource(secrets map[string]string, prefix string) (*VaultSource, error) {
	if secrets == nil {
		return nil, fmt.Errorf("vault: secrets map must not be nil")
	}
	normalized := make(map[string]string, len(secrets))
	for k, v := range secrets {
		normalized[strings.TrimSpace(k)] = v
	}
	return &VaultSource{
		prefix:  prefix,
		secrets: normalized,
	}, nil
}

// Get retrieves a secret by key. It tries the prefixed key first,
// then falls back to the bare key.
func (v *VaultSource) Get(key string) (string, bool) {
	if v.prefix != "" {
		if val, ok := v.secrets[v.prefix+key]; ok {
			return val, true
		}
	}
	val, ok := v.secrets[key]
	return val, ok
}

// Keys returns all keys exposed by this source, stripping the prefix.
func (v *VaultSource) Keys() []string {
	seen := make(map[string]struct{})
	for k := range v.secrets {
		bare := k
		if v.prefix != "" && strings.HasPrefix(k, v.prefix) {
			bare = strings.TrimPrefix(k, v.prefix)
		}
		seen[bare] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Name returns a human-readable identifier for this source.
func (v *VaultSource) Name() string {
	if v.prefix != "" {
		return fmt.Sprintf("vault(prefix=%s)", v.prefix)
	}
	return "vault"
}
