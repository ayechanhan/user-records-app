package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
)

const claimsContextKey = "auth_claims"

// RequireAuth validates the session cookie and stores the claims in the
// request context for downstream handlers.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie(auth.CookieName)
		if err != nil || tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := auth.ParseToken(jwtSecret, tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

// RequireAdmin rejects the request unless RequireAuth has already run and the
// authenticated identity has the Admin role.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if claims.Role != auth.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// ClaimsFromContext returns the authenticated identity set by RequireAuth.
func ClaimsFromContext(c *gin.Context) (*auth.Claims, bool) {
	val, ok := c.Get(claimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := val.(*auth.Claims)
	return claims, ok
}
