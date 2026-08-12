package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

type mockLogService struct {
	listForUser func(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error)
}

func (m *mockLogService) ListForUser(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error) {
	return m.listForUser(ctx, userID, page, pageSize)
}

func newLogTestRouter(h *LogHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/:id/logs", h.ListForUser)
	return r
}

func TestLogHandler_ListForUser_HappyPath(t *testing.T) {
	id := uuid.New()
	svc := &mockLogService{
		listForUser: func(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error) {
			assert.Equal(t, id.String(), userID)
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, pageSize)
			return []model.UserLog{
				{UserID: userID, Event: model.EventUserCreated, Data: map[string]any{"name": "Ada"}},
			}, 1, nil
		},
	}
	r := newLogTestRouter(NewLogHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+id.String()+"/logs", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got listLogsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, int64(1), got.Total)
	require.Len(t, got.Logs, 1)
	assert.Equal(t, "user.created", got.Logs[0].Event)
}

func TestLogHandler_ListForUser_InvalidID(t *testing.T) {
	svc := &mockLogService{}
	r := newLogTestRouter(NewLogHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid/logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogHandler_ListForUser_PassesQueryParams(t *testing.T) {
	id := uuid.New()
	svc := &mockLogService{
		listForUser: func(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error) {
			assert.Equal(t, 3, page)
			assert.Equal(t, 5, pageSize)
			return nil, 0, nil
		},
	}
	r := newLogTestRouter(NewLogHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+id.String()+"/logs?page=3&page_size=5", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestLogHandler_ListForUser_ServiceError(t *testing.T) {
	id := uuid.New()
	svc := &mockLogService{
		listForUser: func(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error) {
			return nil, 0, assert.AnError
		},
	}
	r := newLogTestRouter(NewLogHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+id.String()+"/logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
