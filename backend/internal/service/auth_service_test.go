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

// mockUserRepo implements repository.UserRepository. Each method delegates
// to an optional function field so individual tests only need to set up the
// behavior they exercise; unset fields fall back to a harmless zero value.
type mockUserRepo struct {
	create     func(ctx context.Context, u *model.User) error
	getByID    func(ctx context.Context, id uuid.UUID) (*model.User, error)
	getByEmail func(ctx context.Context, email string) (*model.User, error)
	list       func(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	update     func(ctx context.Context, u *model.User) error
	delete     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error {
	if m.create == nil {
		return nil
	}
	return m.create(ctx, u)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.getByID == nil {
		return nil, nil
	}
	return m.getByID(ctx, id)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.getByEmail(ctx, email)
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	if m.list == nil {
		return nil, 0, nil
	}
	return m.list(ctx, limit, offset)
}
func (m *mockUserRepo) Update(ctx context.Context, u *model.User) error {
	if m.update == nil {
		return nil
	}
	return m.update(ctx, u)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.delete == nil {
		return nil
	}
	return m.delete(ctx, id)
}

// mockEmitter implements EventEmitter, recording every event a test can
// assert against instead of needing a real logging.Bus.
type mockEmitter struct {
	events []emittedEvent
}

type emittedEvent struct {
	userID string
	event  model.LogEvent
	data   map[string]any
}

func (m *mockEmitter) Emit(userID string, event model.LogEvent, data map[string]any) {
	m.events = append(m.events, emittedEvent{userID, event, data})
}

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
	emitter := &mockEmitter{}
	svc := NewAuthService(repo, emitter, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	result, err := svc.Login(context.Background(), "Admin@Example.com", testAdminPassword)

	require.NoError(t, err)
	assert.Equal(t, auth.RoleAdmin, result.Role)
	assert.Equal(t, "admin", result.ID)
	require.NotEmpty(t, result.Token)

	claims, err := auth.ParseToken(testJWTSecret, result.Token)
	require.NoError(t, err)
	assert.Equal(t, auth.RoleAdmin, claims.Role)
	assert.Equal(t, "admin@example.com", claims.Email)

	require.Len(t, emitter.events, 1)
	assert.Equal(t, "admin", emitter.events[0].userID)
	assert.Equal(t, model.EventUserLogin, emitter.events[0].event)
}

func TestAuthService_Login_AdminWrongPassword(t *testing.T) {
	repo := &mockUserRepo{}
	emitter := &mockEmitter{}
	svc := NewAuthService(repo, emitter, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), testAdminEmail, "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	require.Len(t, emitter.events, 1)
	assert.Equal(t, "admin", emitter.events[0].userID)
	assert.Equal(t, model.EventUserLoginFailed, emitter.events[0].event)
}

func TestAuthService_Login_UserSuccess(t *testing.T) {
	user := newTestUser(t, "correct-password")
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			require.Equal(t, "ada@example.com", email)
			return user, nil
		},
	}
	emitter := &mockEmitter{}
	svc := NewAuthService(repo, emitter, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	result, err := svc.Login(context.Background(), "Ada@Example.com", "correct-password")

	require.NoError(t, err)
	assert.Equal(t, auth.RoleUser, result.Role)
	assert.Equal(t, user.ID.String(), result.ID)
	require.NotEmpty(t, result.Token)

	require.Len(t, emitter.events, 1)
	assert.Equal(t, user.ID.String(), emitter.events[0].userID)
	assert.Equal(t, model.EventUserLogin, emitter.events[0].event)
}

func TestAuthService_Login_UserWrongPassword(t *testing.T) {
	user := newTestUser(t, "correct-password")
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			return user, nil
		},
	}
	emitter := &mockEmitter{}
	svc := NewAuthService(repo, emitter, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), user.Email, "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	require.Len(t, emitter.events, 1)
	assert.Equal(t, user.ID.String(), emitter.events[0].userID)
	assert.Equal(t, model.EventUserLoginFailed, emitter.events[0].event)
}

func TestAuthService_Login_UnknownEmail(t *testing.T) {
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	emitter := &mockEmitter{}
	svc := NewAuthService(repo, emitter, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), "nobody@example.com", "any-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	require.Len(t, emitter.events, 1)
	assert.Equal(t, "", emitter.events[0].userID)
	assert.Equal(t, model.EventUserLoginFailed, emitter.events[0].event)
}

func TestAuthService_Login_RepositoryError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*model.User, error) {
			return nil, repoErr
		},
	}
	emitter := &mockEmitter{}
	svc := NewAuthService(repo, emitter, testHMACSecret, testJWTSecret, testAdminEmail, testAdminPassword)

	_, err := svc.Login(context.Background(), "ada@example.com", "any-password")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, emitter.events, "an infra error is not a login outcome and shouldn't be logged as one")
}
