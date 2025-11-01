package models

import (
	"time"

	"github.com/google/uuid"
)

// User struct to describe user object.
type User struct {
	ID           uuid.UUID `json:"id" validate:"required,uuid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email" validate:"required,email,lte=255"`
	PasswordHash string    `json:"password_hash,omitempty" validate:"required,lte=255"`
	UserStatus   int       `json:"user_status" validate:"required,len=1"`
	UserRole     string    `json:"user_role" validate:"required,lte=25"`
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
