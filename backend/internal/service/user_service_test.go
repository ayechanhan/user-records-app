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

func TestUserService_Create_Success(t *testing.T) {
	var created *model.User
	repo := &mockUserRepo{
		create: func(ctx context.Context, u *model.User) error {
			u.ID = uuid.New()
			created = u
			return nil
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	u, err := svc.Create(context.Background(), CreateUserInput{Name: "Ada", Email: "ada@example.com", Password: "correct-password"})

	require.NoError(t, err)
	assert.Equal(t, "Ada", u.Name)
	assert.NotEmpty(t, u.PasswordSalt)
	assert.NotEmpty(t, u.PasswordHash)
	assert.NotEqual(t, "correct-password", u.PasswordHash, "password must never be stored in plaintext")
	require.NotNil(t, created)

	require.Len(t, emitter.events, 1)
	assert.Equal(t, u.ID.String(), emitter.events[0].userID)
	assert.Equal(t, model.EventUserCreated, emitter.events[0].event)
	assert.NotContains(t, emitter.events[0].data, "password")
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{
		create: func(ctx context.Context, u *model.User) error {
			return repository.ErrDuplicateEmail
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	_, err := svc.Create(context.Background(), CreateUserInput{Name: "Ada", Email: "ada@example.com", Password: "correct-password"})

	assert.ErrorIs(t, err, repository.ErrDuplicateEmail)
	assert.Empty(t, emitter.events, "no event should be emitted when the write fails")
}

func TestUserService_Get_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	svc := NewUserService(repo, &mockEmitter{}, testHMACSecret)

	_, err := svc.Get(context.Background(), uuid.New())

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUserService_List_ClampsPagination(t *testing.T) {
	var gotLimit, gotOffset int
	repo := &mockUserRepo{
		list: func(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
			gotLimit, gotOffset = limit, offset
			return []model.User{{Name: "Ada"}}, 1, nil
		},
	}
	svc := NewUserService(repo, &mockEmitter{}, testHMACSecret)

	tests := []struct {
		name               string
		page, pageSize     int
		wantLimit, wantOff int
	}{
		{"defaults on zero", 0, 0, 20, 0},
		{"page 2", 2, 10, 10, 10},
		{"page_size above max clamps to default", 1, 1000, 20, 0},
		{"negative page clamps to 1", -5, 10, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.List(context.Background(), tt.page, tt.pageSize)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, gotLimit)
			assert.Equal(t, tt.wantOff, gotOffset)
		})
	}
}

func TestUserService_Update_EmitsDiffOfChangedFields(t *testing.T) {
	id := uuid.New()
	existing := &model.User{ID: id, Name: "Ada", Email: "ada@example.com", PasswordHash: "old-hash", PasswordSalt: "old-salt"}
	var saved *model.User
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*model.User, error) {
			require.Equal(t, id, gotID)
			return existing, nil
		},
		update: func(ctx context.Context, u *model.User) error {
			saved = u
			return nil
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	_, err := svc.Update(context.Background(), id, UpdateUserInput{Name: "Ada L.", Email: "ada@example.com", Password: "new-password"})

	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, "Ada L.", saved.Name)
	assert.NotEqual(t, "old-hash", saved.PasswordHash)

	require.Len(t, emitter.events, 1)
	assert.Equal(t, model.EventUserUpdated, emitter.events[0].event)
	diff := emitter.events[0].data
	assert.Contains(t, diff, "name")
	assert.NotContains(t, diff, "email", "email did not change, so it should not appear in the diff")
	assert.Equal(t, "changed", diff["password"], "the diff must never contain the actual password or hash")
}

func TestUserService_Update_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	_, err := svc.Update(context.Background(), uuid.New(), UpdateUserInput{Name: "Ada", Email: "ada@example.com"})

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Empty(t, emitter.events)
}

func TestUserService_Update_DuplicateEmail(t *testing.T) {
	existing := &model.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*model.User, error) { return existing, nil },
		update: func(ctx context.Context, u *model.User) error {
			return repository.ErrDuplicateEmail
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	_, err := svc.Update(context.Background(), existing.ID, UpdateUserInput{Name: "Ada", Email: "taken@example.com"})

	assert.ErrorIs(t, err, repository.ErrDuplicateEmail)
	assert.Empty(t, emitter.events, "no event should be emitted when the write fails")
}

func TestUserService_Delete_Success(t *testing.T) {
	id := uuid.New()
	existing := &model.User{ID: id, Name: "Ada", Email: "ada@example.com"}
	var deletedID uuid.UUID
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, gotID uuid.UUID) (*model.User, error) { return existing, nil },
		delete: func(ctx context.Context, gotID uuid.UUID) error {
			deletedID = gotID
			return nil
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	err := svc.Delete(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, id, deletedID)
	require.Len(t, emitter.events, 1)
	assert.Equal(t, id.String(), emitter.events[0].userID)
	assert.Equal(t, model.EventUserDeleted, emitter.events[0].event)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	err := svc.Delete(context.Background(), uuid.New())

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Empty(t, emitter.events)
}

func TestUserService_Delete_RepoErrorNotLogged(t *testing.T) {
	existing := &model.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
	repoErr := errors.New("connection refused")
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*model.User, error) { return existing, nil },
		delete: func(ctx context.Context, id uuid.UUID) error {
			return repoErr
		},
	}
	emitter := &mockEmitter{}
	svc := NewUserService(repo, emitter, testHMACSecret)

	err := svc.Delete(context.Background(), existing.ID)

	require.Error(t, err)
	assert.Empty(t, emitter.events)
}

// verify HashPassword/GenerateSalt are actually exercised through the
// service (not stubbed out), catching accidental drift from auth.VerifyPassword.
func TestUserService_Create_PasswordVerifiesAgainstStoredHash(t *testing.T) {
	var created *model.User
	repo := &mockUserRepo{
		create: func(ctx context.Context, u *model.User) error {
			created = u
			return nil
		},
	}
	svc := NewUserService(repo, &mockEmitter{}, testHMACSecret)

	_, err := svc.Create(context.Background(), CreateUserInput{Name: "Ada", Email: "ada@example.com", Password: "correct-password"})
	require.NoError(t, err)

	assert.True(t, auth.VerifyPassword(testHMACSecret, created.PasswordSalt, "correct-password", created.PasswordHash))
	assert.False(t, auth.VerifyPassword(testHMACSecret, created.PasswordSalt, "wrong-password", created.PasswordHash))
}
