package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenTTL is how long an issued session token remains valid.
const TokenTTL = 24 * time.Hour

// CookieName is the httpOnly cookie the token is issued in and read back from.
const CookieName = "auth_token"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// Claims identifies the authenticated principal. UserID is the literal string
// "admin" for the config-based Admin identity, or a Users.id for a User. Name
// is carried in the token (not just looked up from the DB) so GET /auth/me
// can answer from the cookie alone.
type Claims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

// IssueToken signs a session JWT for the given identity.
func IssueToken(secret, userID, name, email string, role Role) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Name:   name,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}
	return signed, nil
}

// ParseToken validates a session JWT and returns its claims.
func ParseToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("auth: invalid token")
	}
	return claims, nil
}
