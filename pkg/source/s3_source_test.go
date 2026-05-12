package source

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// mockS3Client implements S3Client for testing.
type mockS3Client struct {
	body string
	err  error
}

func (m *mockS3Client) GetObject(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(m.body)),
	}, nil
}

func TestS3Source_GetFound(t *testing.T) {
	client := &mockS3Client{body: `{"DB_HOST":"localhost","DB_PORT":"5432"}`}
	src, err := NewS3Source(context.Background(), client, "my-bucket", "envs/prod.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := src.Get("DB_HOST")
	if !ok || v != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q ok=%v", v, ok)
	}
}

func TestS3Source_GetMissing(t *testing.T) {
	client := &mockS3Client{body: `{"DB_HOST":"localhost"}`}
	src, err := NewS3Source(context.Background(), client, "my-bucket", "envs/prod.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := src.Get("MISSING_KEY")
	if ok {
		t.Error("expected MISSING_KEY to be absent")
	}
}

func TestS3Source_ClientError(t *testing.T) {
	client := &mockS3Client{err: errors.New("NoSuchBucket")}
	_, err := NewS3Source(context.Background(), client, "bad-bucket", "file.json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestS3Source_InvalidJSON(t *testing.T) {
	client := &mockS3Client{body: `not-json`}
	_, err := NewS3Source(context.Background(), client, "my-bucket", "bad.json")
	if err == nil {
		t.Fatal("expected JSON parse error, got nil")
	}
}

func TestS3Source_Keys(t *testing.T) {
	client := &mockS3Client{body: `{"FOO":"1","BAR":"2"}`}
	src, err := NewS3Source(context.Background(), client, "my-bucket", "envs/prod.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys := src.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestS3Source_Name(t *testing.T) {
	client := &mockS3Client{body: `{}`}
	src, err := NewS3Source(context.Background(), client, "my-bucket", "envs/prod.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Name() != "s3://my-bucket/envs/prod.json" {
		t.Errorf("unexpected name: %s", src.Name())
	}
}
