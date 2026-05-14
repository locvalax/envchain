package source

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoRealClientAdapter wraps *mongo.Client to satisfy MongoDBClient.
type mongoRealClientAdapter struct {
	client *mongo.Client
}

func (a *mongoRealClientAdapter) Database(name string) MongoDBDatabase {
	return &mongoRealDatabaseAdapter{db: a.client.Database(name)}
}

type mongoRealDatabaseAdapter struct {
	db *mongo.Database
}

func (d *mongoRealDatabaseAdapter) Collection(name string) MongoDBCollection {
	return &mongoRealCollectionAdapter{coll: d.db.Collection(name)}
}

type mongoRealCollectionAdapter struct {
	coll *mongo.Collection
}

func (c *mongoRealCollectionAdapter) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) MongoDBSingleResult {
	return c.coll.FindOne(ctx, filter, opts...)
}

func (c *mongoRealCollectionAdapter) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return c.coll.Find(ctx, filter, opts...)
}

// NewMongoDBSourceFromClient creates a MongoDBSource from a real *mongo.Client.
func NewMongoDBSourceFromClient(client *mongo.Client, database, collection, keyField, valueField string) Source {
	return NewMongoDBSource(
		&mongoRealClientAdapter{client: client},
		database,
		collection,
		keyField,
		valueField,
	)
}
