package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const userLogsCollection = "user_logs"

// Connect opens a Mongo client and verifies the connection with a ping.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo: connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo: ping: %w", err)
	}
	return client, nil
}

// EnsureIndexes creates the indexes user_logs is queried by: a lookup index
// on user_id, a filter index on event, and a sort/range index on created_at.
// CreateMany is idempotent — an index that already exists with the same spec
// is a no-op.
func EnsureIndexes(ctx context.Context, client *mongo.Client, dbName string) error {
	coll := client.Database(dbName).Collection(userLogsCollection)

	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "event", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := coll.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("mongo: ensure user_logs indexes: %w", err)
	}
	return nil
}
