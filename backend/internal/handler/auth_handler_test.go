package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
	"github.com/ayechanhan/user-records-app/backend/internal/middleware"
	"github.com/ayechanhan/user-records-app/backend/internal/service"
)

// mockAuthService implements the authService seam so Login can be tested
// without a real UserRepository or EventEmitter.
type mockAuthService struct {
	login func(ctx context.Context, email, password string) (*service.LoginResult, error)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*service.LoginResult, error) {
	return m.login(ctx, email, password)
}

func postJSON(t *testing.T, r http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, json.NewEncoder(&buf).Encode(body))
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthHandler_Login_HappyPath(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, email, password string) (*service.LoginResult, error) {
			assert.Equal(t, "ada@example.com", email)
			assert.Equal(t, "correct-password", password)
			return &service.LoginResult{Token: "signed-token", ID: "u1", Name: "Ada", Email: email, Role: auth.RoleUser}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", NewAuthHandler(svc).Login)

	w := postJSON(t, r, "/auth/login", loginRequest{Email: "ada@example.com", Password: "correct-password"})

	require.Equal(t, http.StatusOK, w.Code)
	var got identityResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "u1", got.ID)
	assert.Equal(t, "user", got.Role)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, auth.CookieName, cookies[0].Name)
	assert.Equal(t, "signed-token", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, email, password string) (*service.LoginResult, error) {
			t.Fatal("service should not be called when validation fails")
			return nil, nil
		},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", NewAuthHandler(svc).Login)

	w := postJSON(t, r, "/auth/login", map[string]string{"email": "not-an-email", "password": "x"})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, email, password string) (*service.LoginResult, error) {
			return nil, service.ErrInvalidCredentials
		},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", NewAuthHandler(svc).Login)

	w := postJSON(t, r, "/auth/login", loginRequest{Email: "ada@example.com", Password: "wrong-password"})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Empty(t, w.Result().Cookies(), "no cookie should be set on failed login")
}

func TestAuthHandler_Me_ReturnsSessionClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/me", middleware.RequireAuth("test-secret"), NewAuthHandler(nil).Me)

	token, err := auth.IssueToken("test-secret", "u1", "Ada", "ada@example.com", auth.RoleUser)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got identityResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "u1", got.ID)
	assert.Equal(t, "Ada", got.Name)
	assert.Equal(t, "user", got.Role)
}

func TestAuthHandler_Me_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/me", middleware.RequireAuth("test-secret"), NewAuthHandler(nil).Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_ClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/logout", NewAuthHandler(nil).Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, auth.CookieName, cookies[0].Name)
	assert.Empty(t, cookies[0].Value)
	assert.True(t, cookies[0].MaxAge < 0, "cookie must be expired, not just empty")
}
