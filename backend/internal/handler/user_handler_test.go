package handler

import (
	"bytes"
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
	"github.com/ayechanhan/user-records-app/backend/internal/repository"
	"github.com/ayechanhan/user-records-app/backend/internal/service"
)

// mockUserService implements the userService seam so handler tests never
// touch a real database.
type mockUserService struct {
	create func(ctx context.Context, input service.CreateUserInput) (*model.User, error)
	get    func(ctx context.Context, id uuid.UUID) (*model.User, error)
	list   func(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
	update func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*model.User, error)
	delete func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserService) Create(ctx context.Context, input service.CreateUserInput) (*model.User, error) {
	return m.create(ctx, input)
}
func (m *mockUserService) Get(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return m.get(ctx, id)
}
func (m *mockUserService) List(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	return m.list(ctx, page, pageSize)
}
func (m *mockUserService) Update(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*model.User, error) {
	return m.update(ctx, id, input)
}
func (m *mockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.delete(ctx, id)
}

func newTestRouter(h *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/users", h.Create)
	r.GET("/users", h.List)
	r.GET("/users/:id", h.Get)
	r.PUT("/users/:id", h.Update)
	r.DELETE("/users/:id", h.Delete)
	return r
}

func doJSON(t *testing.T, r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUserHandler_Create_HappyPath(t *testing.T) {
	want := &model.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
	svc := &mockUserService{
		create: func(ctx context.Context, input service.CreateUserInput) (*model.User, error) {
			assert.Equal(t, "Ada", input.Name)
			return want, nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodPost, "/users", createUserRequest{Name: "Ada", Email: "ada@example.com", Password: "correct-password"})

	require.Equal(t, http.StatusCreated, w.Code)
	var got userResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, want.ID.String(), got.ID)
}

func TestUserHandler_Create_ValidationError(t *testing.T) {
	tests := []struct {
		name string
		body createUserRequest
	}{
		{"missing name", createUserRequest{Email: "ada@example.com", Password: "correct-password"}},
		{"invalid email", createUserRequest{Name: "Ada", Email: "not-an-email", Password: "correct-password"}},
		{"password too short", createUserRequest{Name: "Ada", Email: "ada@example.com", Password: "short"}},
	}
	svc := &mockUserService{
		create: func(ctx context.Context, input service.CreateUserInput) (*model.User, error) {
			t.Fatal("service should not be called when validation fails")
			return nil, nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doJSON(t, r, http.MethodPost, "/users", tt.body)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUserHandler_Create_DuplicateEmail(t *testing.T) {
	svc := &mockUserService{
		create: func(ctx context.Context, input service.CreateUserInput) (*model.User, error) {
			return nil, repository.ErrDuplicateEmail
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodPost, "/users", createUserRequest{Name: "Ada", Email: "ada@example.com", Password: "correct-password"})

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserHandler_Get_HappyPath(t *testing.T) {
	want := &model.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
	svc := &mockUserService{
		get: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			assert.Equal(t, want.ID, id)
			return want, nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodGet, "/users/"+want.ID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	var got userResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, want.Email, got.Email)
}

func TestUserHandler_Get_InvalidID(t *testing.T) {
	svc := &mockUserService{}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodGet, "/users/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	svc := &mockUserService{
		get: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodGet, "/users/"+uuid.New().String(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_List_HappyPath(t *testing.T) {
	svc := &mockUserService{
		list: func(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
			return []model.User{{ID: uuid.New(), Name: "Ada"}}, 1, nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodGet, "/users", nil)

	require.Equal(t, http.StatusOK, w.Code)
	var got listUsersResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, int64(1), got.Total)
	assert.Len(t, got.Users, 1)
}

func TestUserHandler_Update_HappyPath(t *testing.T) {
	id := uuid.New()
	want := &model.User{ID: id, Name: "Ada L.", Email: "ada@example.com"}
	svc := &mockUserService{
		update: func(ctx context.Context, gotID uuid.UUID, input service.UpdateUserInput) (*model.User, error) {
			assert.Equal(t, id, gotID)
			assert.Equal(t, "Ada L.", input.Name)
			return want, nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodPut, "/users/"+id.String(), updateUserRequest{Name: "Ada L.", Email: "ada@example.com"})

	require.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_Update_ValidationError(t *testing.T) {
	svc := &mockUserService{
		update: func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*model.User, error) {
			t.Fatal("service should not be called when validation fails")
			return nil, nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodPut, "/users/"+uuid.New().String(), updateUserRequest{Email: "not-an-email"})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	svc := &mockUserService{
		update: func(ctx context.Context, id uuid.UUID, input service.UpdateUserInput) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodPut, "/users/"+uuid.New().String(), updateUserRequest{Name: "Ada", Email: "ada@example.com"})

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_Delete_HappyPath(t *testing.T) {
	id := uuid.New()
	svc := &mockUserService{
		delete: func(ctx context.Context, gotID uuid.UUID) error {
			assert.Equal(t, id, gotID)
			return nil
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodDelete, "/users/"+id.String(), nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	svc := &mockUserService{
		delete: func(ctx context.Context, id uuid.UUID) error {
			return repository.ErrNotFound
		},
	}
	r := newTestRouter(NewUserHandler(svc))

	w := doJSON(t, r, http.MethodDelete, "/users/"+uuid.New().String(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
