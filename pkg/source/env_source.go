package source

import (
	"os"
	"strings"
)

// EnvSource reads environment variables from the process environment.
type EnvSource struct {
	prefix string
}

// NewEnvSource creates an EnvSource that optionally filters by prefix.
// Pass an empty string to include all environment variables.
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{prefix: prefix}
}

func (e *EnvSource) Name() string { return "env" }

func (e *EnvSource) Get(key string) (string, bool) {
	if e.prefix != "" && !strings.HasPrefix(key, e.prefix) {
		return "", false
	}
	v, ok := os.LookupEnv(key)
	return v, ok
}

func (e *EnvSource) Keys() []string {
	var keys []string
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) < 1 {
			continue
		}
		key := parts[0]
		if e.prefix == "" || strings.HasPrefix(key, e.prefix) {
			keys = append(keys, key)
		}
	}
	return keys
}
