// Package service holds business logic that sits between HTTP handlers and
// the repository layer.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
	"github.com/ayechanhan/user-records-app/backend/internal/repository"
)

// ErrInvalidCredentials is returned for any login failure — unknown email,
// wrong password, whether Admin or User — so handlers can return one generic
// response without leaking which case occurred.
var ErrInvalidCredentials = errors.New("service: invalid credentials")

type AuthService struct {
	userRepo      repository.UserRepository
	hmacSecret    string
	jwtSecret     string
	adminEmail    string
	adminPassword string
}

func NewAuthService(userRepo repository.UserRepository, hmacSecret, jwtSecret, adminEmail, adminPassword string) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		hmacSecret:    hmacSecret,
		jwtSecret:     jwtSecret,
		adminEmail:    adminEmail,
		adminPassword: adminPassword,
	}
}

// LoginResult is the authenticated identity returned on a successful login.
type LoginResult struct {
	Token string
	ID    string
	Name  string
	Email string
	Role  auth.Role
}

// Login authenticates the Admin (config-based identity) or a User (Users row)
// through the same flow and issues a session JWT on success.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == strings.ToLower(s.adminEmail) {
		if !auth.VerifyAdminPassword(s.hmacSecret, password, s.adminPassword) {
			return nil, ErrInvalidCredentials
		}
		token, err := auth.IssueToken(s.jwtSecret, "admin", email, auth.RoleAdmin)
		if err != nil {
			return nil, fmt.Errorf("service: issue admin token: %w", err)
		}
		return &LoginResult{Token: token, ID: "admin", Name: "Admin", Email: email, Role: auth.RoleAdmin}, nil
	}

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("service: lookup user: %w", err)
	}

	if !auth.VerifyPassword(s.hmacSecret, u.PasswordSalt, password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := auth.IssueToken(s.jwtSecret, u.ID.String(), u.Email, auth.RoleUser)
	if err != nil {
		return nil, fmt.Errorf("service: issue user token: %w", err)
	}
	return &LoginResult{Token: token, ID: u.ID.String(), Name: u.Name, Email: u.Email, Role: auth.RoleUser}, nil
}
