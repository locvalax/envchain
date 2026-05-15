//go:build integration
// +build integration

package source

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// TestSQSSource_Integration requires:
//   - AWS credentials in environment (or ~/.aws/credentials)
//   - ENVCHAIN_TEST_SQS_QUEUE set to a queue name containing JSON messages
//
// Run with: go test -tags=integration ./pkg/source/ -run TestSQSSource_Integration
func TestSQSSource_Integration(t *testing.T) {
	queueName := os.Getenv("ENVCHAIN_TEST_SQS_QUEUE")
	if queueName == "" {
		t.Skip("ENVCHAIN_TEST_SQS_QUEUE not set")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("load aws config: %v", err)
	}

	client := sqs.NewFromConfig(cfg)
	src, err := NewSQSSource(client, queueName)
	if err != nil {
		t.Fatalf("NewSQSSource: %v", err)
	}

	keys := src.Keys()
	t.Logf("SQS source loaded %d key(s): %v", len(keys), keys)

	expectedKey := os.Getenv("ENVCHAIN_TEST_SQS_EXPECTED_KEY")
	if expectedKey == "" {
		return
	}
	v, ok := src.Get(expectedKey)
	if !ok {
		t.Errorf("expected key %q not found in SQS source", expectedKey)
	} else {
		t.Logf("key %q = %q", expectedKey, v)
	}
}
