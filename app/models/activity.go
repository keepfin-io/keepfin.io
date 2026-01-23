package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityType constants
const (
	ActivityTypeBuy        = "buy"
	ActivityTypeSell       = "sell"
	ActivityTypeDividend   = "dividend"
	ActivityTypeInterest   = "interest"
	ActivityTypeDeposit    = "deposit"
	ActivityTypeWithdrawal = "withdrawal"
)

type Activity struct {
	Base
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AccountID uuid.UUID  `gorm:"type:uuid;not null;index" json:"account_id" validate:"required"`
	HoldingID *uuid.UUID `gorm:"type:uuid;index" json:"holding_id,omitempty"`
	Symbol    string     `gorm:"type:varchar(20);not null" json:"symbol" validate:"required,lte=20"`
	Name      string     `gorm:"type:varchar(100)" json:"name"`
	Type      string     `gorm:"type:varchar(20);not null" json:"type" validate:"required,oneof=buy sell dividend interest deposit withdrawal"`
	Date      time.Time  `gorm:"not null;index" json:"date" validate:"required"`
	Quantity  float64    `gorm:"type:decimal(18,8);not null" json:"quantity" validate:"required"`
	UnitPrice float64    `gorm:"type:decimal(18,4);not null" json:"unit_price" validate:"required,gte=0"`
	Fee       float64    `gorm:"type:decimal(18,4);default:0" json:"fee"`
	Currency  string     `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Notes     string     `gorm:"type:text" json:"notes,omitempty"`
	Total     float64    `gorm:"-" json:"total"`
	Account   Account    `gorm:"foreignKey:AccountID" json:"-"`
	Holding   *Holding   `gorm:"foreignKey:HoldingID" json:"-"`
}

func (Activity) TableName() string {
	return "activities"
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	return a.Base.BeforeCreate(tx)
}

func (a *Activity) AfterFind(tx *gorm.DB) error {
	a.Total = (a.Quantity * a.UnitPrice) + a.Fee
	return nil
}

type CreateActivityRequest struct {
	Symbol    string  `json:"symbol" validate:"required,lte=20"`
	Type      string  `json:"type" validate:"required,oneof=buy sell dividend interest deposit withdrawal"`
	Date      string  `json:"date" validate:"required"`
	Quantity  float64 `json:"quantity" validate:"required"`
	UnitPrice float64 `json:"unit_price" validate:"required,gte=0"`
	Fee       float64 `json:"fee"`
	Currency  string  `json:"currency"`
	Notes     string  `json:"notes"`
}

func ValidActivityTypes() []string {
	return []string{
		ActivityTypeBuy,
		ActivityTypeSell,
		ActivityTypeDividend,
		ActivityTypeInterest,
		ActivityTypeDeposit,
		ActivityTypeWithdrawal,
	}
}
