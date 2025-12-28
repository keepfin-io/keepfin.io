package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User struct to describe user object.
type User struct {
	Base                   // Embedded: HashID, IsDummy, CreatedAt, UpdatedAt, DeletedAt
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email" validate:"required,email,lte=255"`
	PasswordHash string    `gorm:"not null" json:"password_hash,omitempty" validate:"required,lte=255"`
	UserStatus   int       `gorm:"not null;default:1" json:"user_status" validate:"required,len=1"`
	UserRole     string    `gorm:"not null;default:'user'" json:"user_role" validate:"required,lte=25"`
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook - call Base's BeforeCreate
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return u.Base.BeforeCreate(tx)
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
