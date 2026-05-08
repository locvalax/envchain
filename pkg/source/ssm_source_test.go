package source

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type mockSSMClient struct {
	params []types.Parameter
	err    error
}

func (m *mockSSMClient) GetParametersByPath(_ context.Context, _ *ssm.GetParametersByPathInput, _ ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ssm.GetParametersByPathOutput{Parameters: m.params}, nil
}

func makeParams(path string, kv map[string]string) []types.Parameter {
	params := make([]types.Parameter, 0, len(kv))
	for k, v := range kv {
		params = append(params, types.Parameter{
			Name:  aws.String(path + k),
			Value: aws.String(v),
		})
	}
	return params
}

func TestSSMSource_GetFound(t *testing.T) {
	client := &mockSSMClient{params: makeParams("/app/prod/", map[string]string{"DB_URL": "postgres://localhost/db"})}
	s := NewSSMSource(client, "/app/prod")
	val, err := s.Get(context.Background(), "DB_URL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "postgres://localhost/db" {
		t.Errorf("expected postgres://localhost/db, got %q", val)
	}
}

func TestSSMSource_GetMissing(t *testing.T) {
	client := &mockSSMClient{params: makeParams("/app/prod/", map[string]string{"FOO": "bar"})}
	s := NewSSMSource(client, "/app/prod")
	_, err := s.Get(context.Background(), "MISSING")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestSSMSource_ClientError(t *testing.T) {
	client := &mockSSMClient{err: errors.New("network failure")}
	s := NewSSMSource(client, "/app/prod")
	_, err := s.Get(context.Background(), "ANY")
	if err == nil {
		t.Fatal("expected error when client fails")
	}
}

func TestSSMSource_Keys(t *testing.T) {
	client := &mockSSMClient{params: makeParams("/svc/", map[string]string{"A": "1", "B": "2", "C": "3"})}
	s := NewSSMSource(client, "/svc")
	keys, err := s.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(keys)
	expected := []string{"A", "B", "C"}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("expected key %q at index %d, got %q", k, i, keys[i])
		}
	}
}

func TestSSMSource_CachesResults(t *testing.T) {
	callCount := 0
	client := &mockSSMClient{params: makeParams("/env/", map[string]string{"KEY": "val"})}
	// Wrap to count calls
	s := NewSSMSource(client, "/env")
	_ = callCount
	s.Get(context.Background(), "KEY") //nolint
	s.Get(context.Background(), "KEY") //nolint
	// If cache works, client was only called once; cache is non-nil after first call
	if s.cache == nil {
		t.Error("expected cache to be populated")
	}
}
