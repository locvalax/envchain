package source

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AzureServiceBusReceiver is an interface for receiving messages from Azure Service Bus.
type AzureServiceBusReceiver interface {
	ReceiveMessage(ctx context.Context) (string, error)
}

type azureServiceBusSource struct {
	receiver  AzureServiceBusReceiver
	prefix    string
	knownKeys []string
	cache     map[string]string
}

// NewAzureServiceBusSource creates a Source that reads a JSON-encoded environment map
// from a single Azure Service Bus message. The message body must be a flat JSON object.
// An optional prefix is stripped from keys during lookup.
func NewAzureServiceBusSource(ctx context.Context, receiver AzureServiceBusReceiver, prefix string, knownKeys []string) (Source, error) {
	s := &azureServiceBusSource{
		receiver:  receiver,
		prefix:    prefix,
		knownKeys: knownKeys,
		cache:     make(map[string]string),
	}
	if err := s.load(ctx); err != nil {
		return nil, fmt.Errorf("azure_servicebus_source: failed to load message: %w", err)
	}
	return s, nil
}

func (s *azureServiceBusSource) load(ctx context.Context) error {
	body, err := s.receiver.ReceiveMessage(ctx)
	if err != nil {
		return err
	}
	var raw map[string]string
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return fmt.Errorf("invalid JSON in service bus message: %w", err)
	}
	for k, v := range raw {
		s.cache[k] = v
	}
	return nil
}

func (s *azureServiceBusSource) Get(key string) (string, bool) {
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + key
	}
	v, ok := s.cache[lookup]
	return v, ok
}

func (s *azureServiceBusSource) Keys() []string {
	if len(s.knownKeys) > 0 {
		return s.knownKeys
	}
	keys := make([]string, 0, len(s.cache))
	for k := range s.cache {
		if s.prefix != "" {
			if strings.HasPrefix(k, s.prefix) {
				keys = append(keys, strings.TrimPrefix(k, s.prefix))
			}
		} else {
			keys = append(keys, k)
		}
	}
	return keys
}
