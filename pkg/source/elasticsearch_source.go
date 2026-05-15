package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ElasticsearchClient defines the interface for querying Elasticsearch.
type ElasticsearchClient interface {
	Get(ctx context.Context, index, id string) (map[string]interface{}, error)
	Keys(ctx context.Context, index string) ([]string, error)
}

type elasticsearchSource struct {
	client ElasticsearchClient
	index  string
	prefix string
}

// NewElasticsearchSource creates a Source backed by an Elasticsearch index.
// Each document ID is treated as a key; the value is read from the "value" field.
func NewElasticsearchSource(client ElasticsearchClient, index string, opts ...func(*elasticsearchSource)) Source {
	s := &elasticsearchSource{client: client, index: index}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithElasticsearchPrefix strips the given prefix from document IDs when resolving keys.
func WithElasticsearchPrefix(prefix string) func(*elasticsearchSource) {
	return func(s *elasticsearchSource) { s.prefix = prefix }
}

func (s *elasticsearchSource) Get(ctx context.Context, key string) (string, bool) {
	docID := key
	if s.prefix != "" {
		docID = s.prefix + key
	}
	doc, err := s.client.Get(ctx, s.index, docID)
	if err != nil {
		return "", false
	}
	v, ok := doc["value"]
	if !ok {
		return "", false
	}
	str, ok := v.(string)
	return str, ok
}

func (s *elasticsearchSource) Keys(ctx context.Context) ([]string, error) {
	all, err := s.client.Keys(ctx, s.index)
	if err != nil {
		return nil, err
	}
	if s.prefix == "" {
		return all, nil
	}
	var out []string
	for _, k := range all {
		if strings.HasPrefix(k, s.prefix) {
			out = append(out, strings.TrimPrefix(k, s.prefix))
		}
	}
	return out, nil
}

// httpElasticsearchClient is a minimal REST-based Elasticsearch client.
type httpElasticsearchClient struct {
	baseURL string
	hc      *http.Client
}

// NewHTTPElasticsearchClient creates an Elasticsearch client using the REST API.
func NewHTTPElasticsearchClient(baseURL string) ElasticsearchClient {
	return &httpElasticsearchClient{baseURL: strings.TrimRight(baseURL, "/"), hc: &http.Client{}}
}

func (c *httpElasticsearchClient) Get(ctx context.Context, index, id string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/_doc/%s", c.baseURL, index, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Source map[string]interface{} `json:"_source"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Source, nil
}

func (c *httpElasticsearchClient) Keys(ctx context.Context, index string) ([]string, error) {
	return nil, fmt.Errorf("Keys not supported via HTTP client; use a custom implementation")
}
