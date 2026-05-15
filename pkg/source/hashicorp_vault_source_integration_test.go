//go:build integration
// +build integration

package source

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"encoding/json"
)

// realHCVaultClient is a minimal Vault KV v2 client for integration tests.
type realHCVaultClient struct {
	addr  string
	token string
	hc    *http.Client
}

func (c *realHCVaultClient) Read(ctx context.Context, path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/%s", c.addr, path)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("X-Vault-Token", c.token)
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	var body struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body.Data.Data, nil
}

func (c *realHCVaultClient) List(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestHashiCorpVaultSource_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip("VAULT_ADDR or VAULT_TOKEN not set")
	}
	client := &realHCVaultClient{addr: strings.TrimRight(addr, "/"), token: token, hc: http.DefaultClient}
	src := NewHashiCorpVaultSource(client, "secret", "envchain/test", "")
	_, _, err := src.Get(context.Background(), "TEST_KEY")
	if err != nil {
		t.Fatalf("integration get failed: %v", err)
	}
}
