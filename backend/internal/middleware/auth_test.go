package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
)

const testSecret = "test-secret"

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", RequireAuth(testSecret), func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no claims"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID, "role": claims.Role})
	})
	r.GET("/admin-only", RequireAuth(testSecret), RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func doRequest(r http.Handler, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequireAuth_NoCookie(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, "/protected", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, "/protected", &http.Cookie{Name: auth.CookieName, Value: "not-a-jwt"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	token, err := auth.IssueToken("other-secret", "u1", "Ada", "ada@example.com", auth.RoleUser)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	r := newTestRouter()
	w := doRequest(r, "/protected", &http.Cookie{Name: auth.CookieName, Value: token})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken_SetsClaims(t *testing.T) {
	token, err := auth.IssueToken(testSecret, "u1", "Ada", "ada@example.com", auth.RoleUser)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	r := newTestRouter()
	w := doRequest(r, "/protected", &http.Cookie{Name: auth.CookieName, Value: token})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireAdmin_UserRoleForbidden(t *testing.T) {
	token, err := auth.IssueToken(testSecret, "u1", "Ada", "ada@example.com", auth.RoleUser)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	r := newTestRouter()
	w := doRequest(r, "/admin-only", &http.Cookie{Name: auth.CookieName, Value: token})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRequireAdmin_AdminRoleAllowed(t *testing.T) {
	token, err := auth.IssueToken(testSecret, "admin", "Admin", "admin@example.com", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	r := newTestRouter()
	w := doRequest(r, "/admin-only", &http.Cookie{Name: auth.CookieName, Value: token})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireAdmin_NoAuthRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin-only", RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := doRequest(r, "/admin-only", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when RequireAdmin runs without RequireAuth first, got %d", w.Code)
	}
}

func TestClaimsFromContext_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if _, ok := ClaimsFromContext(c); ok {
		t.Fatal("expected ClaimsFromContext to return false when no claims were set")
	}
}
