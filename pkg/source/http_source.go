package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is the interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpSource struct {
	url    string
	client HTTPClient
	cache  map[string]string
}

// NewHTTPSource creates a Source that fetches environment variables from a
// remote HTTP endpoint. The endpoint must return a JSON object with string
// key-value pairs, e.g. {"DB_HOST": "localhost", "DB_PORT": "5432"}.
func NewHTTPSource(url string, timeout time.Duration) (Source, error) {
	client := &http.Client{Timeout: timeout}
	s := &httpSource{url: url, client: client}
	if err := s.load(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

// NewHTTPSourceWithClient creates an HTTPSource with a custom HTTPClient (useful for testing).
func NewHTTPSourceWithClient(url string, client HTTPClient) (Source, error) {
	s := &httpSource{url: url, client: client}
	if err := s.load(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *httpSource) load(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return fmt.Errorf("http_source: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("http_source: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http_source: unexpected status %d from %s", resp.StatusCode, s.url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http_source: read body: %w", err)
	}

	var data map[string]string
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("http_source: parse JSON: %w", err)
	}

	s.cache = data
	return nil
}

func (s *httpSource) Get(key string) (string, bool) {
	v, ok := s.cache[key]
	return v, ok
}

func (s *httpSource) Keys() []string {
	keys := make([]string, 0, len(s.cache))
	for k := range s.cache {
		keys = append(keys, k)
	}
	return keys
}
