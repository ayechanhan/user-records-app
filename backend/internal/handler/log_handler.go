package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

// logService is the seam LogHandler depends on instead of the concrete
// *service.LogService, so tests can supply a lightweight mock.
type logService interface {
	ListForUser(ctx context.Context, userID string, page, pageSize int) ([]model.UserLog, int64, error)
}

type LogHandler struct {
	logService logService
}

func NewLogHandler(s logService) *LogHandler {
	return &LogHandler{logService: s}
}

type logResponse struct {
	UserID    string         `json:"user_id"`
	Event     string         `json:"event"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

type listLogsResponse struct {
	Logs     []logResponse `json:"logs"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// ListForUser returns a paginated log history for the user in :id. The user
// does not need to currently exist — a soft-deleted user's log history
// stays queryable, which is the entire point of soft delete (see spec.md
// Assumptions).
func (h *LogHandler) ListForUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	logs, total, err := h.logService.ListForUser(c.Request.Context(), id.String(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resp := listLogsResponse{Logs: make([]logResponse, len(logs)), Total: total, Page: page, PageSize: pageSize}
	for i, l := range logs {
		resp.Logs[i] = logResponse{UserID: l.UserID, Event: string(l.Event), Data: l.Data, CreatedAt: l.CreatedAt}
	}
	c.JSON(http.StatusOK, resp)
}
