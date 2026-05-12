package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client defines the interface for S3 operations used by S3Source.
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// s3Source loads environment variables from a JSON file stored in S3.
type s3Source struct {
	client S3Client
	bucket string
	key    string
	data   map[string]string
}

// NewS3Source creates a Source that reads a JSON key-value file from an S3 bucket.
// The file at the given key must be a flat JSON object: {"KEY": "value", ...}.
func NewS3Source(ctx context.Context, client S3Client, bucket, key string) (Source, error) {
	s := &s3Source{
		client: client,
		bucket: bucket,
		key:    key,
	}
	if err := s.load(ctx); err != nil {
		return nil, fmt.Errorf("s3_source: failed to load s3://%s/%s: %w", bucket, key, err)
	}
	return s, nil
}

func (s *s3Source) load(ctx context.Context) error {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	body, err := io.ReadAll(out.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("unmarshalling JSON: %w", err)
	}
	s.data = raw
	return nil
}

func (s *s3Source) Get(key string) (string, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *s3Source) Keys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *s3Source) Name() string {
	return fmt.Sprintf("s3://%s/%s", s.bucket, strings.TrimPrefix(s.key, "/"))
}
