package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ayechanhan/user-records-app/backend/internal/auth"
	"github.com/ayechanhan/user-records-app/backend/internal/model"
	"github.com/ayechanhan/user-records-app/backend/internal/repository"
)

type UserService struct {
	userRepo   repository.UserRepository
	events     EventEmitter
	hmacSecret string
}

func NewUserService(userRepo repository.UserRepository, events EventEmitter, hmacSecret string) *UserService {
	return &UserService{userRepo: userRepo, events: events, hmacSecret: hmacSecret}
}

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
}

// Create hashes the password with a fresh per-user salt, persists the user,
// and emits exactly one user.created event.
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*model.User, error) {
	salt, err := auth.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("service: generate salt: %w", err)
	}

	u := &model.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: auth.HashPassword(s.hmacSecret, salt, input.Password),
		PasswordSalt: salt,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	s.events.Emit(u.ID.String(), model.EventUserCreated, map[string]any{
		"name":  u.Name,
		"email": u.Email,
	})
	return u, nil
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.userRepo.List(ctx, pageSize, offset)
}

type UpdateUserInput struct {
	Name     string
	Email    string
	Password string // empty means "don't change"
}

// Update applies a full replace of name/email (and password, if provided),
// then emits exactly one user.updated event whose data is a diff of the
// fields that actually changed — never the password or its hash.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*model.User, error) {
	existing, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	diff := map[string]any{}
	if existing.Name != input.Name {
		diff["name"] = map[string]string{"from": existing.Name, "to": input.Name}
		existing.Name = input.Name
	}
	if existing.Email != input.Email {
		diff["email"] = map[string]string{"from": existing.Email, "to": input.Email}
		existing.Email = input.Email
	}
	if input.Password != "" {
		salt, err := auth.GenerateSalt()
		if err != nil {
			return nil, fmt.Errorf("service: generate salt: %w", err)
		}
		existing.PasswordSalt = salt
		existing.PasswordHash = auth.HashPassword(s.hmacSecret, salt, input.Password)
		diff["password"] = "changed"
	}

	if err := s.userRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.events.Emit(id.String(), model.EventUserUpdated, diff)
	return existing, nil
}

// Delete soft-deletes the user and emits exactly one user.deleted event.
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.events.Emit(id.String(), model.EventUserDeleted, map[string]any{
		"name":  existing.Name,
		"email": existing.Email,
	})
	return nil
}
