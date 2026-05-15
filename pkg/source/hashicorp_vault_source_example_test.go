package source_test

import (
	"context"
	"fmt"

	"github.com/yourorg/envchain/pkg/chain"
	"github.com/yourorg/envchain/pkg/source"
)

type exampleVaultClient struct{}

func (e *exampleVaultClient) Read(_ context.Context, _ string) (map[string]interface{}, error) {
	return map[string]interface{}{"DB_PASSWORD": "vault-secret"}, nil
}

func (e *exampleVaultClient) List(_ context.Context, _ string) ([]string, error) {
	return []string{"DB_PASSWORD"}, nil
}

func Example_hashiCorpVaultChain() {
	vaultSrc := source.NewHashiCorpVaultSource(
		&exampleVaultClient{},
		"secret",
		"myapp/production",
		"",
	)

	// Vault has highest priority; fall back to process env.
	envSrc := source.NewEnvSource("")

	c := chain.New(vaultSrc, envSrc)

	val, ok, err := c.Resolve(context.Background(), "DB_PASSWORD")
	if err != nil {
		panic(err)
	}
	if ok {
		fmt.Println("resolved:", val)
	}
	// Output: resolved: vault-secret
}
