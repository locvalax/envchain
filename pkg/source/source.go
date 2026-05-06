package source

// Source represents a named provider of environment variables.
type Source interface {
	// Name returns the identifier for this source.
	Name() string
	// Get retrieves the value for the given key.
	// Returns the value and true if found, empty string and false otherwise.
	Get(key string) (string, bool)
	// Keys returns all keys available in this source.
	Keys() []string
}

// MapSource is a simple in-memory source backed by a map.
type MapSource struct {
	name string
	data map[string]string
}

// NewMapSource creates a new MapSource with the given name and data.
func NewMapSource(name string, data map[string]string) *MapSource {
	copy := make(map[string]string, len(data))
	for k, v := range data {
		copy[k] = v
	}
	return &MapSource{name: name, data: copy}
}

func (m *MapSource) Name() string { return m.name }

func (m *MapSource) Get(key string) (string, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *MapSource) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}
