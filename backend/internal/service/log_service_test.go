package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

// mockLogRepo implements repository.LogRepository for LogService tests.
type mockLogRepo struct {
	create       func(ctx context.Context, entry *model.UserLog) error
	listByUserID func(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error)
}

func (m *mockLogRepo) Create(ctx context.Context, entry *model.UserLog) error {
	if m.create == nil {
		return nil
	}
	return m.create(ctx, entry)
}

func (m *mockLogRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error) {
	return m.listByUserID(ctx, userID, limit, offset)
}

func TestLogService_ListForUser_ClampsPagination(t *testing.T) {
	var gotLimit, gotOffset int
	repo := &mockLogRepo{
		listByUserID: func(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error) {
			gotLimit, gotOffset = limit, offset
			return []model.UserLog{{UserID: userID, Event: model.EventUserLogin}}, 1, nil
		},
	}
	svc := NewLogService(repo)

	tests := []struct {
		name               string
		page, pageSize     int
		wantLimit, wantOff int
	}{
		{"defaults on zero", 0, 0, 20, 0},
		{"page 2", 2, 10, 10, 10},
		{"page_size above max clamps to default", 1, 1000, 20, 0},
		{"negative page clamps to 1", -5, 10, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.ListForUser(context.Background(), "u1", tt.page, tt.pageSize)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, gotLimit)
			assert.Equal(t, tt.wantOff, gotOffset)
		})
	}
}

func TestLogService_ListForUser_ReturnsRepoResults(t *testing.T) {
	logs := []model.UserLog{
		{UserID: "u1", Event: model.EventUserCreated},
		{UserID: "u1", Event: model.EventUserLogin},
	}
	repo := &mockLogRepo{
		listByUserID: func(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error) {
			assert.Equal(t, "u1", userID)
			return logs, 2, nil
		},
	}
	svc := NewLogService(repo)

	got, total, err := svc.ListForUser(context.Background(), "u1", 1, 20)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, logs, got)
}

func TestLogService_ListForUser_WorksForNonexistentUser(t *testing.T) {
	// ListByUserID never checks user existence — a soft-deleted (or even
	// never-existed) user's id should just return an empty page, not an error.
	repo := &mockLogRepo{
		listByUserID: func(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error) {
			return []model.UserLog{}, 0, nil
		},
	}
	svc := NewLogService(repo)

	got, total, err := svc.ListForUser(context.Background(), "does-not-exist", 1, 20)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, got)
}
