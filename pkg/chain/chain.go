package chain

import (
	"fmt"
	"os"
)

// Source represents a single environment variable source with a priority level.
type Source struct {
	Name     string
	Priority int
	Vars     map[string]string
}

// Chain holds an ordered list of sources and resolves variables by precedence.
type Chain struct {
	sources []*Source
}

// New creates an empty Chain.
func New() *Chain {
	return &Chain{}
}

// AddSource appends a source to the chain.
func (c *Chain) AddSource(s *Source) {
	c.sources = append(c.sources, s)
}

// Resolve returns the value for key by iterating sources in descending
// priority order. The first match wins. Falls back to the process
// environment when no source provides the key.
func (c *Chain) Resolve(key string) (string, bool) {
	best := (*Source)(nil)
	for _, s := range c.sources {
		if _, ok := s.Vars[key]; ok {
			if best == nil || s.Priority > best.Priority {
				best = s
			}
		}
	}
	if best != nil {
		return best.Vars[key], true
	}
	val, ok := os.LookupEnv(key)
	return val, ok
}

// ResolveAll merges all sources respecting priority and returns a flat map.
// Higher-priority sources override lower-priority ones for the same key.
func (c *Chain) ResolveAll() map[string]string {
	// Collect every key across all sources.
	keys := map[string]struct{}{}
	for _, s := range c.sources {
		for k := range s.Vars {
			keys[k] = struct{}{}
		}
	}

	result := make(map[string]string, len(keys))
	for k := range keys {
		if v, ok := c.Resolve(k); ok {
			result[k] = v
		}
	}
	return result
}

// Inject writes the resolved environment into the current process via os.Setenv.
// It returns the first error encountered, if any.
func (c *Chain) Inject() error {
	for k, v := range c.ResolveAll() {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("envchain: failed to set %q: %w", k, err)
		}
	}
	return nil
}
