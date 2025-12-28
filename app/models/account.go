package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	Base                     // Embedded: HashID, IsDummy, CreatedAt, UpdatedAt, DeletedAt
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	AccountName    string    `gorm:"type:varchar(100);not null" json:"account_name" validate:"required,lte=100"`
	Currency       string    `gorm:"type:varchar(3);not null" json:"currency" validate:"required,currency"`
	InitialBalance float64   `gorm:"type:decimal(18,2);default:0" json:"initial_balance"`
	Platform       string    `gorm:"type:varchar(100)" json:"platform" validate:"platform"`
	Description    string    `gorm:"type:varchar(255)" json:"description,omitempty"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	IsExcluded     bool      `gorm:"default:false" json:"is_excluded"`
}

func (Account) TableName() string {
	return "accounts"
}

// BeforeCreate hook - call Base's BeforeCreate
func (a *Account) BeforeCreate(tx *gorm.DB) error {
	return a.Base.BeforeCreate(tx)
}
