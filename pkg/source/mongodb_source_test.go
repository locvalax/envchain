package source

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- mocks ---

type mockMongoResult struct {
	doc bson.M
	err error
}

func (r *mockMongoResult) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(v.(*bson.M)) = r.doc
	return nil
}

type mockMongoColl struct {
	docs []bson.M
	err  error
}

func (c *mockMongoColl) FindOne(_ context.Context, filter interface{}, _ ...*options.FindOneOptions) MongoDBSingleResult {
	if c.err != nil {
		return &mockMongoResult{err: c.err}
	}
	f := filter.(bson.M)
	for _, doc := range c.docs {
		for k, v := range f {
			if doc[k] == v {
				return &mockMongoResult{doc: doc}
			}
		}
	}
	return &mockMongoResult{err: mongo.ErrNoDocuments}
}

func (c *mockMongoColl) Find(_ context.Context, _ interface{}, _ ...*options.FindOptions) (*mongo.Cursor, error) {
	return nil, errors.New("not implemented in mock")
}

type mockMongoDatabase struct{ coll *mockMongoColl }

func (d *mockMongoDatabase) Collection(_ string) MongoDBCollection { return d.coll }

type mockMongoClient struct{ db *mockMongoDatabase }

func (c *mockMongoClient) Database(_ string) MongoDBDatabase { return c.db }

func newMockMongoSource(docs []bson.M, clientErr error) Source {
	coll := &mockMongoColl{docs: docs, err: clientErr}
	client := &mockMongoClient{db: &mockMongoDatabase{coll: coll}}
	return NewMongoDBSource(client, "testdb", "secrets", "key", "value")
}

// --- tests ---

func TestMongoDBSource_GetFound(t *testing.T) {
	src := newMockMongoSource([]bson.M{{"key": "DB_PASS", "value": "s3cr3t"}}, nil)
	val, err := src.Get(context.Background(), "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %q", val)
	}
}

func TestMongoDBSource_GetMissing(t *testing.T) {
	src := newMockMongoSource([]bson.M{}, nil)
	_, err := src.Get(context.Background(), "MISSING")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestMongoDBSource_ClientError(t *testing.T) {
	src := newMockMongoSource(nil, errors.New("connection refused"))
	_, err := src.Get(context.Background(), "ANY")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMongoDBSource_NonStringValue(t *testing.T) {
	src := newMockMongoSource([]bson.M{{"key": "NUM", "value": 42}}, nil)
	_, err := src.Get(context.Background(), "NUM")
	if err == nil {
		t.Fatal("expected type error, got nil")
	}
}
