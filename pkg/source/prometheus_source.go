package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// prometheusClient defines the interface for fetching Prometheus query results.
type prometheusClient interface {
	Get(url string) (*http.Response, error)
}

// PrometheusSource reads environment variable values from a Prometheus HTTP API
// by issuing instant queries. Each key maps to a PromQL expression.
type PrometheusSource struct {
	baseURL    string
	client     prometheusClient
	prefix     string
	knownKeys  []string
	queries    map[string]string // key -> PromQL expression override
}

type prometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value []json.RawMessage `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// NewPrometheusSource creates a PrometheusSource that queries the given base URL.
// opts may include WithPrometheusPrefix, WithPrometheusKnownKeys, WithPrometheusQueries.
func NewPrometheusSource(baseURL string, opts ...func(*PrometheusSource)) *PrometheusSource {
	s := &PrometheusSource{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
		queries: make(map[string]string),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithPrometheusPrefix sets an optional key prefix to strip before forming queries.
func WithPrometheusPrefix(prefix string) func(*PrometheusSource) {
	return func(s *PrometheusSource) { s.prefix = prefix }
}

// WithPrometheusKnownKeys registers keys returned by Keys().
func WithPrometheusKnownKeys(keys []string) func(*PrometheusSource) {
	return func(s *PrometheusSource) { s.knownKeys = keys }
}

// WithPrometheusQueries maps env-var keys to custom PromQL expressions.
func WithPrometheusQueries(queries map[string]string) func(*PrometheusSource) {
	return func(s *PrometheusSource) {
		for k, v := range queries {
			s.queries[k] = v
		}
	}
}

func (s *PrometheusSource) Get(_ context.Context, key string) (string, bool, error) {
	bare := strings.TrimPrefix(key, s.prefix)
	expr, ok := s.queries[bare]
	if !ok {
		expr = bare
	}
	url := fmt.Sprintf("%s/api/v1/query?query=%s", s.baseURL, expr)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", false, fmt.Errorf("prometheus: GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, fmt.Errorf("prometheus: read body: %w", err)
	}
	var pr prometheusResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return "", false, fmt.Errorf("prometheus: unmarshal: %w", err)
	}
	if pr.Status != "success" || len(pr.Data.Result) == 0 {
		return "", false, nil
	}
	result := pr.Data.Result[0]
	if len(result.Value) < 2 {
		return "", false, nil
	}
	var val string
	if err := json.Unmarshal(result.Value[1], &val); err != nil {
		return "", false, fmt.Errorf("prometheus: parse value: %w", err)
	}
	return val, true, nil
}

func (s *PrometheusSource) Keys(_ context.Context) ([]string, error) {
	return s.knownKeys, nil
}
