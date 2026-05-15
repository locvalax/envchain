//go:build ignore
// +build ignore

package source_test

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"

	"github.com/yourorg/envchain/pkg/chain"
	"github.com/yourorg/envchain/pkg/source"
)

func Example_firebaseChain() {
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, &firebase.Config{
		DatabaseURL: "https://my-project-default-rtdb.firebaseio.com",
	})
	if err != nil {
		log.Fatalf("init firebase: %v", err)
	}

	dbClient, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("get db client: %v", err)
	}

	// Firebase source with highest priority
	fbSrc := source.NewFirebaseSource(dbClient, "/envchain/prod", "")

	// Fall back to process environment
	envSrc := source.NewEnvSource("")

	// Build chain: Firebase > Env
	c := chain.New(fbSrc, envSrc)

	val, ok, err := c.Resolve(ctx, "DB_HOST")
	if err != nil {
		log.Fatalf("resolve: %v", err)
	}
	if !ok {
		fmt.Println("DB_HOST not found in any source")
		return
	}
	fmt.Printf("DB_HOST = %s\n", val)
}
