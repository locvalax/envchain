package source

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type mockSMClient struct {
	secrets map[string]string
	listErr error
	getErr  error
}

func (m *mockSMClient) AccessSecretVersion(_ context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	val, ok := m.secrets[req.Name]
	if !ok {
		return nil, errors.New("NotFound: secret not found")
	}
	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(val)},
	}, nil
}

func (m *mockSMClient) ListSecrets(_ context.Context, req *secretmanagerpb.ListSecretsRequest) ([]*secretmanagerpb.Secret, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var out []*secretmanagerpb.Secret
	for name := range m.secrets {
		out = append(out, &secretmanagerpb.Secret{Name: "projects/proj/secrets/" + name})
	}
	return out, nil
}

func TestSecretManagerSource_GetFound(t *testing.T) {
	client := &mockSMClient{
		secrets: map[string]string{
			"projects/proj/secrets/DB_PASS/versions/latest": "s3cr3t",
		},
	}
	src := NewSecretManagerSource(client, "proj", "")
	val, ok, err := src.Get(context.Background(), "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %q", val)
	}
}

func TestSecretManagerSource_GetMissing(t *testing.T) {
	client := &mockSMClient{secrets: map[string]string{}}
	src := NewSecretManagerSource(client, "proj", "")
	_, ok, err := src.Get(context.Background(), "MISSING_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestSecretManagerSource_ClientError(t *testing.T) {
	client := &mockSMClient{getErr: errors.New("permission denied")}
	src := NewSecretManagerSource(client, "proj", "")
	_, _, err := src.Get(context.Background(), "KEY")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestSecretManagerSource_PrefixedGet(t *testing.T) {
	client := &mockSMClient{
		secrets: map[string]string{
			"projects/proj/secrets/app_TOKEN/versions/latest": "tok123",
		},
	}
	src := NewSecretManagerSource(client, "proj", "app_")
	val, ok, err := src.Get(context.Background(), "TOKEN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found with prefix")
	}
	if val != "tok123" {
		t.Errorf("expected tok123, got %q", val)
	}
}

func TestSecretManagerSource_Keys(t *testing.T) {
	client := &mockSMClient{
		secrets: map[string]string{
			"app_FOO": "v1",
			"app_BAR": "v2",
			"other":   "v3",
		},
	}
	src := NewSecretManagerSource(client, "proj", "app_")
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}
