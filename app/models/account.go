package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id" validate:"required"`

	AccountName string `gorm:"type:varchar(100);not null" json:"account_name" validate:"required,lte=100"`
	Currency    string `gorm:"type:varchar(3);not null" json:"currency" validate:"required,currency"`

	InitialBalance float64 `gorm:"type:decimal(18,2);default:0" json:"initial_balance"`
	Platform       string  `gorm:"type:varchar(100)" json:"platform"`
	Description    string  `gorm:"type:varchar(255)" json:"description,omitempty"`

	IsExcluded bool `gorm:"default:false" json:"is_excluded"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Account) TableName() string {
	return "accounts"
}
