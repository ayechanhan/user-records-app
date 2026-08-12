// Package handler holds the Gin HTTP handlers.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
	"github.com/ayechanhan/user-records-app/backend/internal/middleware"
	"github.com/ayechanhan/user-records-app/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: s}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type identityResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Login authenticates the Admin or a User and, on success, sets the session
// JWT as an httpOnly cookie rather than returning it in the response body.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	secure := gin.Mode() == gin.ReleaseMode
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.CookieName, result.Token, int(auth.TokenTTL.Seconds()), "/", "", secure, true)

	c.JSON(http.StatusOK, identityResponse{
		ID:    result.ID,
		Name:  result.Name,
		Email: result.Email,
		Role:  string(result.Role),
	})
}

// Me returns the identity carried in the session cookie's JWT claims, so the
// frontend can determine auth state server-side without ever touching the
// token itself. Requires RequireAuth to have run first.
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, identityResponse{
		ID:    claims.UserID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  string(claims.Role),
	})
}

// Logout clears the session cookie. JWTs are stateless, so there is nothing
// to invalidate server-side — clearing the cookie is sufficient.
func (h *AuthHandler) Logout(c *gin.Context) {
	secure := gin.Mode() == gin.ReleaseMode
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.CookieName, "", -1, "/", "", secure, true)
	c.Status(http.StatusNoContent)
}
