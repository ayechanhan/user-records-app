package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
	"github.com/ayechanhan/user-records-app/backend/internal/middleware"
)

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
