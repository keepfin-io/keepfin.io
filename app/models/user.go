package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ThemeLight = "light"
	ThemeDark  = "dark"
)

const (
	LangEnglish = "en"
	LangTurkish = "tr"
)

const (
	LocaleEnUS = "en-US"
	LocaleEnGB = "en-GB"
	LocaleTrTR = "tr-TR"
)

type User struct {
	Base
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email" validate:"required,email,lte=255"`
	PasswordHash string    `gorm:"not null" json:"password_hash,omitempty" validate:"required,lte=255"`
	UserStatus   int       `gorm:"not null;default:1" json:"user_status" validate:"required,len=1"`
	UserRole     string    `gorm:"not null;default:'user'" json:"user_role" validate:"required,lte=25"`
	Language     string    `gorm:"type:varchar(5);not null;default:'en'" json:"language" validate:"omitempty,oneof=en tr"`
	Theme        string    `gorm:"type:varchar(10);not null;default:'light'" json:"theme" validate:"omitempty,oneof=light dark"`
	BaseCurrency string    `gorm:"type:varchar(3);not null;default:'USD'" json:"base_currency" validate:"omitempty,currency"`
	Locale       string    `gorm:"type:varchar(10);not null;default:'en-US'" json:"locale" validate:"omitempty,oneof=en-US en-GB tr-TR"`
}

func (User) TableName() string {
	return "users"
}
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return u.Base.BeforeCreate(tx)
}

type UserSignUp struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}
type UserSignIn struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}

type UpdateProfileRequest struct {
	Language     string `json:"language" validate:"omitempty,oneof=en tr"`
	Theme        string `json:"theme" validate:"omitempty,oneof=light dark"`
	BaseCurrency string `json:"base_currency" validate:"omitempty,currency"`
	Locale       string `json:"locale" validate:"omitempty,oneof=en-US en-GB tr-TR"`
}

type UpdateBaseCurrencyRequest struct {
	BaseCurrency string `json:"base_currency" validate:"required,currency"`
}
