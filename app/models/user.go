package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User struct to describe user object.
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" validate:"required,uuid"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email" validate:"required,email,lte=255"`
	PasswordHash string         `gorm:"not null" json:"password_hash,omitempty" validate:"required,lte=255"`
	UserStatus   int            `gorm:"not null;default:1" json:"user_status" validate:"required,len=1"`
	UserRole     string         `gorm:"not null;default:'user'" json:"user_role" validate:"required,lte=25"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}

// UserSignUp struct to describe register a new user.
type UserSignUp struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}

// UserSignIn struct to describe login user.
type UserSignIn struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}
