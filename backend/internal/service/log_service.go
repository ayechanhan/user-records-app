package service

import (
	"context"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
	"github.com/ayechanhan/user-records-app/backend/internal/repository"
)

type LogService struct {
	logRepo repository.LogRepository
}

func NewLogService(logRepo repository.LogRepository) *LogService {
	return &LogService{logRepo: logRepo}
}

// ListForUser returns a page of a user's log history, newest first. It does
// not require the user to currently exist — see LogRepository.ListByUserID.
func (s *LogService) ListForUser(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.logRepo.ListByUserID(ctx, userID, pageSize, offset)
}
