package source

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretManagerClient defines the interface for GCP Secret Manager operations.
type SecretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	ListSecrets(ctx context.Context, req *secretmanagerpb.ListSecretsRequest) ([]*secretmanagerpb.Secret, error)
}

type secretManagerSource struct {
	client    SecretManagerClient
	projectID string
	prefix    string
}

// NewSecretManagerSource creates a Source backed by GCP Secret Manager.
// prefix is an optional filter applied to secret names (e.g. "myapp/").
func NewSecretManagerSource(client SecretManagerClient, projectID, prefix string) Source {
	return &secretManagerSource{
		client:    client,
		projectID: projectID,
		prefix:    prefix,
	}
}

func (s *secretManagerSource) Get(ctx context.Context, key string) (string, bool, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s%s/versions/latest", s.projectID, s.prefix, key)
	resp, err := s.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		if isNotFoundError(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("secret manager get %q: %w", key, err)
	}
	return string(resp.Payload.Data), true, nil
}

func (s *secretManagerSource) Keys(ctx context.Context) ([]string, error) {
	parent := fmt.Sprintf("projects/%s", s.projectID)
	secrets, err := s.client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{
		Parent: parent,
	})
	if err != nil {
		return nil, fmt.Errorf("secret manager list secrets: %w", err)
	}
	var keys []string
	for _, secret := range secrets {
		parts := strings.Split(secret.Name, "/")
		name := parts[len(parts)-1]
		if s.prefix == "" || strings.HasPrefix(name, s.prefix) {
			key := strings.TrimPrefix(name, s.prefix)
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "not found")
}

// Ensure compile-time interface satisfaction.
var _ secretmanager.Client = (*secretmanager.Client)(nil)
