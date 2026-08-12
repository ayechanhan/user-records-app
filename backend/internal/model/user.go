package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the relational record for an admin- or user-authenticatable account.
// PasswordHash/PasswordSalt back the HMAC-SHA256(key=serverSecret, message=salt+password)
// scheme described in plan.md — the salt is per-user so identical passwords don't
// produce identical hashes across accounts.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	PasswordSalt string    `gorm:"column:password_salt;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
