package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// AzureBlobClient is the interface for reading blobs from Azure Blob Storage.
type AzureBlobClient interface {
	DownloadBlob(ctx context.Context, container, blob string) (io.ReadCloser, error)
}

type azureBlobSource struct {
	client    AzureBlobClient
	container string
	blob      string
	prefix    string
	data      map[string]string
}

// NewAzureBlobSource creates a Source that reads a JSON key/value blob from
// Azure Blob Storage. The optional prefix is stripped from keys when looking
// up values.
func NewAzureBlobSource(client AzureBlobClient, container, blob, prefix string) (Source, error) {
	s := &azureBlobSource{
		client:    client,
		container: container,
		blob:      blob,
		prefix:    prefix,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *azureBlobSource) load() error {
	rc, err := s.client.DownloadBlob(context.Background(), s.container, s.blob)
	if err != nil {
		return fmt.Errorf("azure_blob: download %s/%s: %w", s.container, s.blob, err)
	}
	defer rc.Close()

	var raw map[string]string
	if err := json.NewDecoder(rc).Decode(&raw); err != nil {
		return fmt.Errorf("azure_blob: decode JSON: %w", err)
	}

	s.data = make(map[string]string, len(raw))
	for k, v := range raw {
		s.data[k] = v
	}
	return nil
}

func (s *azureBlobSource) Get(key string) (string, bool) {
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + key
	}
	v, ok := s.data[lookup]
	return v, ok
}

func (s *azureBlobSource) Keys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		if s.prefix == "" || strings.HasPrefix(k, s.prefix) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (s *azureBlobSource) Name() string {
	return fmt.Sprintf("azure-blob://%s/%s", s.container, s.blob)
}
