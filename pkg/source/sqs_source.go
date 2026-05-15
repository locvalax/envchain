package source

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClient defines the interface for interacting with AWS SQS.
type SQSClient interface {
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
}

// sqsSource reads environment variables from messages in an AWS SQS queue.
// Each message body must be a JSON object mapping string keys to string values.
type sqsSource struct {
	client    SQSClient
	queueName string
	prefix    string
	cache     map[string]string
}

// NewSQSSource creates a new SQS-backed source. It fetches up to 10 messages
// from the named queue and merges their JSON bodies into a key/value store.
func NewSQSSource(client SQSClient, queueName string, opts ...func(*sqsSource)) (*sqsSource, error) {
	s := &sqsSource{
		client:    client,
		queueName: queueName,
		cache:     make(map[string]string),
	}
	for _, o := range opts {
		o(s)
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// WithSQSPrefix sets a key prefix that is stripped before storing values.
func WithSQSPrefix(prefix string) func(*sqsSource) {
	return func(s *sqsSource) { s.prefix = prefix }
}

func (s *sqsSource) load() error {
	ctx := context.Background()
	urlOut, err := s.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(s.queueName),
	})
	if err != nil {
		return fmt.Errorf("sqs: get queue url: %w", err)
	}
	out, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            urlOut.QueueUrl,
		MaxNumberOfMessages: 10,
	})
	if err != nil {
		return fmt.Errorf("sqs: receive message: %w", err)
	}
	for _, msg := range out.Messages {
		if msg.Body == nil {
			continue
		}
		var kv map[string]string
		if err := json.Unmarshal([]byte(*msg.Body), &kv); err != nil {
			continue
		}
		for k, v := range kv {
			key := k
			if s.prefix != "" {
				if len(key) > len(s.prefix) && key[:len(s.prefix)] == s.prefix {
					key = key[len(s.prefix):]
				} else {
					continue
				}
			}
			s.cache[key] = v
		}
	}
	return nil
}

func (s *sqsSource) Get(key string) (string, bool) {
	v, ok := s.cache[key]
	return v, ok
}

func (s *sqsSource) Keys() []string {
	keys := make([]string, 0, len(s.cache))
	for k := range s.cache {
		keys = append(keys, k)
	}
	return keys
}
