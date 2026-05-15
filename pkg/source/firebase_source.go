package source

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/db"
)

// FirebaseClient is the interface for Firebase Realtime Database operations.
type FirebaseClient interface {
	NewRef(path string) *db.Ref
}

// firebaseRefClient wraps db.Client to satisfy FirebaseClient.
type firebaseRefGetter interface {
	Get(ctx context.Context, v interface{}) error
}

// FirebaseSource reads environment variables from a Firebase Realtime Database path.
type FirebaseSource struct {
	client FirebaseClient
	basePath string
	prefix string
}

// NewFirebaseSource creates a FirebaseSource using the provided Firebase DB client.
// basePath is the database path where key/value pairs are stored.
// prefix is an optional key prefix filter.
func NewFirebaseSource(client FirebaseClient, basePath, prefix string) *FirebaseSource {
	return &FirebaseSource{
		client:   client,
		basePath: basePath,
		prefix:   prefix,
	}
}

func (s *FirebaseSource) fetchAll(ctx context.Context) (map[string]string, error) {
	ref := s.client.NewRef(s.basePath)
	var data map[string]interface{}
	if err := ref.Get(ctx, &data); err != nil {
		return nil, fmt.Errorf("firebase: failed to get path %q: %w", s.basePath, err)
	}
	result := make(map[string]string, len(data))
	for k, v := range data {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result, nil
}

// Get retrieves the value for the given key from Firebase.
func (s *FirebaseSource) Get(ctx context.Context, key string) (string, bool, error) {
	lookup := key
	if s.prefix != "" {
		lookup = s.prefix + key
	}
	ref := s.client.NewRef(s.basePath + "/" + lookup)
	var val interface{}
	if err := ref.Get(ctx, &val); err != nil {
		return "", false, fmt.Errorf("firebase: get %q: %w", lookup, err)
	}
	if val == nil {
		return "", false, nil
	}
	str, ok := val.(string)
	if !ok {
		return "", false, fmt.Errorf("firebase: value for key %q is not a string", lookup)
	}
	return str, true, nil
}

// Keys returns all keys available under the base path.
func (s *FirebaseSource) Keys(ctx context.Context) ([]string, error) {
	data, err := s.fetchAll(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys, nil
}
