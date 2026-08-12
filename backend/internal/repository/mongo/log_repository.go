package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"

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
