package source

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type mockSQSClient struct {
	queueURL string
	messages []types.Message
	err      error
}

func (m *mockSQSClient) GetQueueUrl(_ context.Context, _ *sqs.GetQueueUrlInput, _ ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &sqs.GetQueueUrlOutput{QueueUrl: aws.String(m.queueURL)}, nil
}

func (m *mockSQSClient) ReceiveMessage(_ context.Context, _ *sqs.ReceiveMessageInput, _ ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &sqs.ReceiveMessageOutput{Messages: m.messages}, nil
}

func makeBody(t *testing.T, kv map[string]string) *string {
	t.Helper()
	b, err := json.Marshal(kv)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	return &s
}

func TestSQSSource_GetFound(t *testing.T) {
	client := &mockSQSClient{
		queueURL: "https://sqs.us-east-1.amazonaws.com/123/test",
		messages: []types.Message{
			{Body: makeBody(t, map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432"})},
		},
	}
	src, err := NewSQSSource(client, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := src.Get("DB_HOST")
	if !ok || v != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q ok=%v", v, ok)
	}
}

func TestSQSSource_GetMissing(t *testing.T) {
	client := &mockSQSClient{
		queueURL: "https://sqs.us-east-1.amazonaws.com/123/test",
		messages: []types.Message{
			{Body: makeBody(t, map[string]string{"FOO": "bar"})},
		},
	}
	src, _ := NewSQSSource(client, "test")
	_, ok := src.Get("MISSING")
	if ok {
		t.Error("expected missing key to return false")
	}
}

func TestSQSSource_ClientError(t *testing.T) {
	client := &mockSQSClient{err: errors.New("connection refused")}
	_, err := NewSQSSource(client, "test")
	if err == nil {
		t.Error("expected error from client, got nil")
	}
}

func TestSQSSource_PrefixedGet(t *testing.T) {
	client := &mockSQSClient{
		queueURL: "https://sqs.us-east-1.amazonaws.com/123/test",
		messages: []types.Message{
			{Body: makeBody(t, map[string]string{"APP_SECRET": "mysecret", "OTHER": "ignored"})},
		},
	}
	src, err := NewSQSSource(client, "test", WithSQSPrefix("APP_"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := src.Get("SECRET")
	if !ok || v != "mysecret" {
		t.Errorf("expected SECRET=mysecret, got %q ok=%v", v, ok)
	}
	_, ok = src.Get("OTHER")
	if ok {
		t.Error("expected OTHER to be filtered out by prefix")
	}
}

func TestSQSSource_Keys(t *testing.T) {
	client := &mockSQSClient{
		queueURL: "https://sqs.us-east-1.amazonaws.com/123/test",
		messages: []types.Message{
			{Body: makeBody(t, map[string]string{"A": "1", "B": "2"})},
		},
	}
	src, _ := NewSQSSource(client, "test")
	keys := src.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
