//go:build integration
// +build integration

package source_test

import (
	"context"
	"os"
	"testing"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"github.com/yourorg/envchain/pkg/source"
)

// TestFirebaseSource_Integration requires a real Firebase project.
// Set FIREBASE_DATABASE_URL and GOOGLE_APPLICATION_CREDENTIALS before running.
func TestFirebaseSource_Integration(t *testing.T) {
	dbURL := os.Getenv("FIREBASE_DATABASE_URL")
	if dbURL == "" {
		t.Skip("FIREBASE_DATABASE_URL not set; skipping integration test")
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{DatabaseURL: dbURL},
		option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		t.Fatalf("failed to init firebase app: %v", err)
	}

	client, err := app.Database(ctx)
	if err != nil {
		t.Fatalf("failed to get database client: %v", err)
	}

	src := source.NewFirebaseSource(client, "/envchain/test", "")

	_, _, err = src.Get(ctx, "TEST_KEY")
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}

	_, err = src.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys returned unexpected error: %v", err)
	}
}
