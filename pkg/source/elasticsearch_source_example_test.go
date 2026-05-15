package source_test

import (
	"context"
	"fmt"

	"github.com/yourorg/envchain/pkg/chain"
	"github.com/yourorg/envchain/pkg/source"
)

// Example_elasticsearchChain demonstrates building an envchain that falls back
// from Elasticsearch to local environment variables.
func Example_elasticsearchChain() {
	// In production replace with source.NewHTTPElasticsearchClient(url).
	mockClient := &exampleESClient{
		docs: map[string]map[string]interface{}{
			"prod/DB_HOST": {"value": "db.internal"},
		},
	}

	esSource := source.NewElasticsearchSource(
		mockClient,
		"envchain",
		source.WithElasticsearchPrefix("prod/"),
	)
	envSource := source.NewEnvSource()

	c := chain.New(esSource, envSource)

	v, ok := c.Resolve(context.Background(), "DB_HOST")
	if ok {
		fmt.Println(v)
	}
	// Output: db.internal
}

type exampleESClient struct {
	docs map[string]map[string]interface{}
}

func (e *exampleESClient) Get(_ context.Context, _, id string) (map[string]interface{}, error) {
	doc, ok := e.docs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return doc, nil
}

func (e *exampleESClient) Keys(_ context.Context, _ string) ([]string, error) {
	keys := make([]string, 0, len(e.docs))
	for k := range e.docs {
		keys = append(keys, k)
	}
	return keys, nil
}
