package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
	"github.com/ayechanhan/user-records-app/backend/internal/model"
	"github.com/ayechanhan/user-records-app/backend/internal/repository"
)

// mockUserRepo implements repository.UserRepository for tests that only
// exercise the login path, i.e. GetByEmail.
type mockUserRepo struct {
	getByEmail func(ctx context.Context, email string) (*model.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.getByEmail(ctx, email)
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) Update(ctx context.Context, u *model.User) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error  { return nil }

const (
	testHMACSecret    = "test-hmac-secret"
	testJWTSecret     = "test-jwt-secret"
	testAdminEmail    = "admin@example.com"
	testAdminPassword = "admin-password"
)

func newTestUser(t *testing.T, password string) *model.User {
	t.Helper()
	salt, err := auth.GenerateSalt()
	require.NoError(t, err)
	return &model.User{
		ID:           uuid.New(),
		Name:         "Ada Lovelace",
		Email:        "ada@example.com",
		PasswordHash: auth.HashPassword(testHMACSecret, salt, password),
		PasswordSalt: salt,
	}
}

func TestAuthService_Login_AdminSuccess(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewAuthService(repo, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	result, err := svc.Login(context.Background(), "Admin@Example.com", testAdminPassword)

	require.NoError(t, err)
	assert.Equal(t, auth.RoleAdmin, result.Role)
	assert.Equal(t, "admin", result.ID)
	require.NotEmpty(t, result.Token)

	claims, err := auth.ParseToken(testJWTSecret, result.Token)
	require.NoError(t, err)
	assert.Equal(t, auth.RoleAdmin, claims.Role)
	assert.Equal(t, "admin@example.com", claims.Email)
}

func TestAuthService_Login_AdminWrongPassword(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewAuthService(repo, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), testAdminEmail, "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_Login_UserSuccess(t *testing.T) {
	user := newTestUser(t, "correct-password")
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			require.Equal(t, "ada@example.com", email)
			return user, nil
		},
	}
	svc := NewAuthService(repo, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	result, err := svc.Login(context.Background(), "Ada@Example.com", "correct-password")

	require.NoError(t, err)
	assert.Equal(t, auth.RoleUser, result.Role)
	assert.Equal(t, user.ID.String(), result.ID)
	require.NotEmpty(t, result.Token)
}

func TestAuthService_Login_UserWrongPassword(t *testing.T) {
	user := newTestUser(t, "correct-password")
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			return user, nil
		},
	}
	svc := NewAuthService(repo, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), user.Email, "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_Login_UnknownEmail(t *testing.T) {
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	svc := NewAuthService(repo, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), "nobody@example.com", "any-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_Login_RepositoryError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			return nil, repoErr
		},
	}
	svc := NewAuthService(repo, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), "ada@example.com", "any-password")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidCredentials)
}
