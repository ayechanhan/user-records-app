package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
	"github.com/ayechanhan/user-records-app/backend/internal/repository"
)

const pgUniqueViolation = "23505"

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.Email = strings.ToLower(u.Email)

	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, translateError(err)
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", strings.ToLower(email)).Error; err != nil {
		return nil, translateError(err)
	}
	return &u, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, translateError(err)
	}
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, translateError(err)
	}
	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	u.Email = strings.ToLower(u.Email)
	if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id)
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func translateError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return repository.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return repository.ErrDuplicateEmail
	}
	return err
}
