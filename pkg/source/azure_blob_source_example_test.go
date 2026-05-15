package source_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/yourorg/envchain/pkg/chain"
	"github.com/yourorg/envchain/pkg/source"
)

// staticBlobClient is a trivial in-process AzureBlobClient for examples.
type staticBlobClient struct {
	payload map[string]string
}

func (s *staticBlobClient) DownloadBlob(_ context.Context, _, _ string) (io.ReadCloser, error) {
	b, _ := json.Marshal(s.payload)
	return io.NopCloser(strings.NewReader(string(b))), nil
}

func Example_azureBlobChain() {
	blobClient := &staticBlobClient{
		payload: map[string]string{
			"APP_DB_HOST": "db.prod.example.com",
			"APP_DB_PORT": "5432",
		},
	}

	blobSrc, err := source.NewAzureBlobSource(blobClient, "config", "app.json", "APP_")
	if err != nil {
		log.Fatal(err)
	}

	envSrc := source.NewEnvSource("") // fallback to process environment

	c := chain.New(blobSrc, envSrc)

	host, ok := c.Resolve("DB_HOST")
	if !ok {
		log.Fatal("DB_HOST not found")
	}
	fmt.Println(host)
	// Output: db.prod.example.com
}
