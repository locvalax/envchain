package source

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DotenvSource reads key=value pairs from a .env file.
type DotenvSource struct {
	path string
	data map[string]string
}

// NewDotenvSource loads a .env file from the given path.
func NewDotenvSource(path string) (*DotenvSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("dotenv: open %q: %w", path, err)
	}
	defer f.Close()

	data := make(map[string]string)
	scanner := bufio.NewScanner(f)
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
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		data[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("dotenv: scan %q: %w", path, err)
	}
	return &DotenvSource{path: path, data: data}, nil
}

func (d *DotenvSource) Name() string { return "dotenv:" + d.path }

func (d *DotenvSource) Get(key string) (string, bool) {
	v, ok := d.data[key]
	return v, ok
}

func (d *DotenvSource) Keys() []string {
	keys := make([]string, 0, len(d.data))
	for k := range d.data {
		keys = append(keys, k)
	}
	return keys
}
