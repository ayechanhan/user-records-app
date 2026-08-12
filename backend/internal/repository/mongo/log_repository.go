package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

type LogRepository struct {
	coll *mongo.Collection
}

func NewLogRepository(client *mongo.Client, dbName string) *LogRepository {
	return &LogRepository{coll: client.Database(dbName).Collection(userLogsCollection)}
}

func (r *LogRepository) Create(ctx context.Context, log *model.UserLog) error {
	if _, err := r.coll.InsertOne(ctx, log); err != nil {
		return fmt.Errorf("mongo: insert user log: %w", err)
	}
	return nil
}

func (r *LogRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error) {
	filter := bson.D{{Key: "user_id", Value: userID}}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongo: count user logs: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongo: find user logs: %w", err)
	}

	var logs []model.UserLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, fmt.Errorf("mongo: decode user logs: %w", err)
	}
	return logs, total, nil
}
