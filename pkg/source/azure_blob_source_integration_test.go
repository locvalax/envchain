//go:build integration
// +build integration

package source

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// realAzureBlobClient wraps the Azure SDK blob client.
type realAzureBlobClient struct {
	client    *azblob.Client
	accountURL string
}

func (r *realAzureBlobClient) DownloadBlob(ctx context.Context, container, blob string) (io.ReadCloser, error) {
	resp, err := r.client.DownloadStream(ctx, container, blob, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func TestAzureBlobSource_Integration(t *testing.T) {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connStr == "" {
		t.Skip("AZURE_STORAGE_CONNECTION_STRING not set")
	}
	container := os.Getenv("AZURE_BLOB_CONTAINER")
	blob := os.Getenv("AZURE_BLOB_NAME")
	if container == "" || blob == "" {
		t.Skip("AZURE_BLOB_CONTAINER or AZURE_BLOB_NAME not set")
	}

	client, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		t.Fatalf("create azure blob client: %v", err)
	}

	src, err := NewAzureBlobSource(&realAzureBlobClient{client: client}, container, blob, "")
	if err != nil {
		t.Fatalf("NewAzureBlobSource: %v", err)
	}

	keys := src.Keys()
	t.Logf("loaded %d keys from azure blob %s/%s", len(keys), container, blob)
}
