package source

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMClient defines the subset of the AWS SSM API used by SSMSource.
type SSMClient interface {
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

// SSMSource resolves environment variables from AWS Systems Manager Parameter Store.
type SSMSource struct {
	client SSMClient
	path   string
	cache  map[string]string
}

// NewSSMSource creates a new SSMSource that fetches parameters under the given path prefix.
// Parameters are loaded lazily on first access.
func NewSSMSource(client SSMClient, path string) *SSMSource {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return &SSMSource{
		client: client,
		path:   path,
	}
}

func (s *SSMSource) load(ctx context.Context) error {
	if s.cache != nil {
		return nil
	}
	s.cache = make(map[string]string)
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(s.path),
		WithDecryption: aws.Bool(true),
		Recursive:      aws.Bool(false),
	}
	out, err := s.client.GetParametersByPath(ctx, input)
	if err != nil {
		return fmt.Errorf("ssm: failed to load parameters from %q: %w", s.path, err)
	}
	for _, p := range out.Parameters {
		if p.Name == nil || p.Value == nil {
			continue
		}
		key := strings.TrimPrefix(*p.Name, s.path)
		s.cache[key] = *p.Value
	}
	return nil
}

// Get returns the value of the parameter with the given key, or an error if not found.
func (s *SSMSource) Get(ctx context.Context, key string) (string, error) {
	if err := s.load(ctx); err != nil {
		return "", err
	}
	val, ok := s.cache[key]
	if !ok {
		return "", fmt.Errorf("ssm: key %q not found under path %q", key, s.path)
	}
	return val, nil
}

// Keys returns all parameter names available under the configured path.
func (s *SSMSource) Keys(ctx context.Context) ([]string, error) {
	if err := s.load(ctx); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(s.cache))
	for k := range s.cache {
		keys = append(keys, k)
	}
	return keys, nil
}
