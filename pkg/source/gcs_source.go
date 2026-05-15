package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

// gcsClient abstracts the GCS operations used by GCSSource.
type gcsClient interface {
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
}

// GCSSource reads environment variables from a JSON object stored in
// a Google Cloud Storage bucket.
type GCSSource struct {
	client    gcsClient
	bucket    string
	objectKey string
	prefix    string
	data      map[string]string
}

type defaultGCSClient struct {
	client *storage.Client
}

func (c *defaultGCSClient) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	reader, err := c.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: open object %s/%s: %w", bucket, key, err)
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// NewGCSSource creates a GCSSource that reads a JSON file from GCS.
// The optional prefix is stripped from keys when looking up values.
func NewGCSSource(ctx context.Context, bucket, objectKey, prefix string) (*GCSSource, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: create client: %w", err)
	}
	return newGCSSourceWithClient(&defaultGCSClient{client: client}, bucket, objectKey, prefix)
}

func newGCSSourceWithClient(client gcsClient, bucket, objectKey, prefix string) (*GCSSource, error) {
	s := &GCSSource{
		client:    client,
		bucket:    bucket,
		objectKey: objectKey,
		prefix:    prefix,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *GCSSource) load() error {
	raw, err := s.client.GetObject(context.Background(), s.bucket, s.objectKey)
	if err != nil {
		return err
	}
	var parsed map[string]string
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("gcs: parse JSON from %s/%s: %w", s.bucket, s.objectKey, err)
	}
	s.data = parsed
	return nil
}

// Get returns the value for key, optionally stripping the configured prefix.
func (s *GCSSource) Get(key string) (string, bool) {
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + key
	}
	v, ok := s.data[lookup]
	return v, ok
}

// Keys returns all keys available in the GCS object.
func (s *GCSSource) Keys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
