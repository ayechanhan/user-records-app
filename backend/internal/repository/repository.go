// Package repository defines the data-access interfaces the service layer
// depends on. Concrete implementations live in the postgres and mongo
// subpackages so unit tests can mock these interfaces instead of standing
// up real databases.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

var (
	ErrNotFound       = errors.New("repository: record not found")
	ErrDuplicateEmail = errors.New("repository: email already exists")
)

// UserRepository is the data-access contract for the relational Users store.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// LogRepository is the data-access contract for the UserLogs document store.
type LogRepository interface {
	Create(ctx context.Context, log *model.UserLog) error
}
