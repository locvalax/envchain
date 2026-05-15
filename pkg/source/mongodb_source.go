package source

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBClient defines the interface for MongoDB operations used by MongoDBSource.
type MongoDBClient interface {
	Database(name string) MongoDBDatabase
}

// MongoDBDatabase defines the interface for a MongoDB database.
type MongoDBDatabase interface {
	Collection(name string) MongoDBCollection
}

// MongoDBCollection defines the interface for a MongoDB collection.
type MongoDBCollection interface {
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) MongoDBSingleResult
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
}

// MongoDBSingleResult defines the interface for a single MongoDB result.
type MongoDBSingleResult interface {
	Decode(v interface{}) error
}

type mongoDBSource struct {
	client     MongoDBClient
	database   string
	collection string
	keyField   string
	valueField string
}

// NewMongoDBSource creates a Source backed by a MongoDB collection.
// Documents are expected to have configurable key and value fields.
func NewMongoDBSource(client MongoDBClient, database, collection, keyField, valueField string) Source {
	return &mongoDBSource{
		client:     client,
		database:   database,
		collection: collection,
		keyField:   keyField,
		valueField: valueField,
	}
}

func (s *mongoDBSource) Get(ctx context.Context, key string) (string, error) {
	coll := s.client.Database(s.database).Collection(s.collection)
	filter := bson.M{s.keyField: key}

	var result bson.M
	if err := coll.FindOne(ctx, filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("mongodb get %q: %w", key, err)
	}

	val, ok := result[s.valueField]
	if !ok {
		return "", ErrKeyNotFound
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("mongodb: value for key %q is not a string", key)
	}
	return str, nil
}

func (s *mongoDBSource) Keys(ctx context.Context) ([]string, error) {
	coll := s.client.Database(s.database).Collection(s.collection)
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{s.keyField: 1}))
	if err != nil {
		return nil, fmt.Errorf("mongodb keys: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []string
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("mongodb keys: decode document: %w", err)
		}
		if k, ok := doc[s.keyField].(string); ok {
			keys = append(keys, k)
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb keys: cursor error: %w", err)
	}
	return keys, nil
}
