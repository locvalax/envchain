package source

import (
	"context"
	"fmt"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// influxDBQueryClient abstracts the InfluxDB query API for testing.
type influxDBQueryClient interface {
	QueryAPI(org string) api.QueryAPI
}

type influxDBSource struct {
	client influxDBQueryClient
	org    string
	bucket string
	prefix string
	keys   []string
}

// NewInfluxDBSource creates a Source that reads key/value pairs from an
// InfluxDB bucket. Each measurement named "envchain" with a tag "key" and
// field "value" is treated as one environment variable.
func NewInfluxDBSource(serverURL, token, org, bucket string, opts ...func(*influxDBSource)) (Source, error) {
	client := influxdb2.NewClient(serverURL, token)
	return newInfluxDBSourceWithClient(client, org, bucket, opts...)
}

func newInfluxDBSourceWithClient(client influxDBQueryClient, org, bucket string, opts ...func(*influxDBSource)) (Source, error) {
	s := &influxDBSource{
		client: client,
		org:    org,
		bucket: bucket,
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// WithInfluxDBPrefix filters keys by a prefix, stripping it before returning.
func WithInfluxDBPrefix(prefix string) func(*influxDBSource) {
	return func(s *influxDBSource) { s.prefix = prefix }
}

// WithInfluxDBKnownKeys declares the set of keys exposed by Keys().
func WithInfluxDBKnownKeys(keys []string) func(*influxDBSource) {
	return func(s *influxDBSource) { s.keys = keys }
}

func (s *influxDBSource) Get(ctx context.Context, key string) (string, bool, error) {
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + key
	}
	flux := fmt.Sprintf(
		`from(bucket:"%s") |> range(start: -1h) |> filter(fn:(r) => r._measurement == "envchain" and r.key == "%s") |> last()`,
		s.bucket, lookup,
	)
	result, err := s.client.QueryAPI(s.org).Query(ctx, flux)
	if err != nil {
		return "", false, fmt.Errorf("influxdb query: %w", err)
	}
	for result.Next() {
		if v, ok := result.Record().Value().(string); ok {
			return v, true, nil
		}
	}
	if err := result.Err(); err != nil {
		return "", false, fmt.Errorf("influxdb result: %w", err)
	}
	return "", false, nil
}

func (s *influxDBSource) Keys(ctx context.Context) ([]string, error) {
	if len(s.keys) > 0 {
		return s.keys, nil
	}
	flux := fmt.Sprintf(
		`from(bucket:"%s") |> range(start: -1h) |> filter(fn:(r) => r._measurement == "envchain") |> keep(columns:["key"]) |> distinct(column:"key")`,
		s.bucket,
	)
	result, err := s.client.QueryAPI(s.org).Query(ctx, flux)
	if err != nil {
		return nil, fmt.Errorf("influxdb keys query: %w", err)
	}
	var keys []string
	for result.Next() {
		if k, ok := result.Record().ValueByKey("key").(string); ok {
			if s.prefix == "" || strings.HasPrefix(k, s.prefix) {
				keys = append(keys, strings.TrimPrefix(k, s.prefix))
			}
		}
	}
	return keys, result.Err()
}
