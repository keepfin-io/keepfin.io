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
	LocaleDeDE = "de-DE"
	LocaleFrFR = "fr-FR"
	LocaleEsES = "es-ES"
	LocalePtBR = "pt-BR"
	LocaleJaJP = "ja-JP"
	LocaleZhCN = "zh-CN"
)

type User struct {
	Base
	ID    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email string    `gorm:"uniqueIndex;not null" json:"email" validate:"required,email,lte=255"`
	// Retained for migration compatibility; never used for authentication or exposed.
	PasswordHash string  `gorm:"not null;default:''" json:"-"`
	AuthUserID   *string `gorm:"uniqueIndex" json:"-"`
	UserStatus   int     `gorm:"not null;default:1" json:"user_status" validate:"required,len=1"`
	UserRole     string  `gorm:"not null;default:'user'" json:"user_role" validate:"required,lte=25"`
	Theme        string  `gorm:"type:varchar(10);not null;default:'light'" json:"theme" validate:"omitempty,oneof=light dark"`
	BaseCurrency string  `gorm:"type:varchar(3);not null;default:'USD'" json:"base_currency" validate:"omitempty,currency"`
	Locale       string  `gorm:"type:varchar(10);not null;default:'en-US'" json:"locale" validate:"omitempty,oneof=en-US en-GB tr-TR de-DE fr-FR es-ES pt-BR ja-JP zh-CN"`
}

func (User) TableName() string {
	return "users"
}
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return u.Base.BeforeCreate(tx)
}

type UpdateProfileRequest struct {
	Language     string `json:"language" validate:"omitempty,oneof=en tr"`
	Theme        string `json:"theme" validate:"omitempty,oneof=light dark"`
	BaseCurrency string `json:"base_currency" validate:"omitempty,currency"`
	Locale       string `json:"locale" validate:"omitempty,oneof=en-US en-GB tr-TR de-DE fr-FR es-ES pt-BR ja-JP zh-CN"`
}

type UpdateBaseCurrencyRequest struct {
	BaseCurrency string `json:"base_currency" validate:"required,currency"`
}

type UpdateThemeRequest struct {
	Theme string `json:"theme" validate:"required,oneof=light dark"`
}

type UpdateLocaleRequest struct {
	Locale string `json:"locale" validate:"required,oneof=en-US en-GB tr-TR de-DE fr-FR es-ES pt-BR ja-JP zh-CN"`
}
