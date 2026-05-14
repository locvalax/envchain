package source

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// fileSource reads environment variables from a plain text file,
// one KEY=VALUE pair per line. Lines starting with '#' are treated as comments.
type fileSource struct {
	path   string
	data   map[string]string
	loaded bool
}

// NewFileSource creates a Source that reads KEY=VALUE pairs from a plain text file.
func NewFileSource(path string) Source {
	return &fileSource{path: path, data: make(map[string]string)}
}

func (f *fileSource) load() error {
	if f.loaded {
		return nil
	}
	file, err := os.Open(f.path)
	if err != nil {
		return fmt.Errorf("file_source: open %q: %w", f.path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key != "" {
			f.data[key] = val
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("file_source: scan %q: %w", f.path, err)
	}
	f.loaded = true
	return nil
}

func (f *fileSource) Get(key string) (string, bool, error) {
	if err := f.load(); err != nil {
		return "", false, err
	}
	v, ok := f.data[key]
	return v, ok, nil
}

func (f *fileSource) Keys() ([]string, error) {
	if err := f.load(); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(f.data))
	for k := range f.data {
		keys = append(keys, k)
	}
	return keys, nil
}
