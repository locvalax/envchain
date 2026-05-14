//go:build integration
// +build integration

package source

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestMongoDBSource_Integration requires a real MongoDB instance.
// Set MONGODB_URI to point to it, e.g. mongodb://localhost:27017
// The test seeds a document and verifies retrieval.
func TestMongoDBSource_Integration(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("envchain_test")
	coll := db.Collection("secrets")
	_, _ = coll.DeleteMany(ctx, map[string]interface{}{})
	_, err = coll.InsertOne(ctx, map[string]interface{}{"key": "INTEGRATION_KEY", "value": "integration_value"})
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	adapter := &mongoRealClientAdapter{client: client}
	src := NewMongoDBSource(adapter, "envchain_test", "secrets", "key", "value")

	val, err := src.Get(ctx, "INTEGRATION_KEY")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "integration_value" {
		t.Errorf("expected integration_value, got %q", val)
	}
}
