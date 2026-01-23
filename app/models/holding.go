package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Holding struct {
	Base
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AccountID   uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id" validate:"required"`
	Symbol      string    `gorm:"type:varchar(20);not null" json:"symbol" validate:"required,lte=20"`
	Name        string    `gorm:"type:varchar(100)" json:"name"`
	Quantity    float64   `gorm:"type:decimal(18,8);not null" json:"quantity" validate:"required,gt=0"`
	BuyPrice    float64   `gorm:"type:decimal(18,4);not null" json:"buy_price" validate:"required,gte=0"`
	BuyCurrency string    `gorm:"type:varchar(3);default:'USD'" json:"buy_currency"`
	Notes       string    `gorm:"type:text" json:"notes,omitempty"`
	Account     Account   `gorm:"foreignKey:AccountID" json:"-"`
}

func (Holding) TableName() string {
	return "holdings"
}

func (h *Holding) BeforeCreate(tx *gorm.DB) error {
	return h.Base.BeforeCreate(tx)
}

type HoldingWithLiveData struct {
	ID            uuid.UUID `json:"id"`
	AccountID     uuid.UUID `json:"account_id"`
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Quantity      float64   `json:"quantity"`
	BuyPrice      float64   `json:"buy_price"`
	BuyCurrency   string    `json:"buy_currency"`
	CurrentPrice  float64   `json:"current_price"`
	Value         float64   `json:"value"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Performance   float64   `json:"performance"`
	ProfitLoss    float64   `json:"profit_loss"`
	FirstActivity time.Time `json:"first_activity"`
}

type HoldingsSummary struct {
	TotalValue       float64 `json:"total_value"`
	TotalCost        float64 `json:"total_cost"`
	TotalProfitLoss  float64 `json:"total_profit_loss"`
	TotalPerformance float64 `json:"total_performance"`
	DailyChange      float64 `json:"daily_change"`
}

type CreateHoldingRequest struct {
	Symbol      string  `json:"symbol" validate:"required,lte=20"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	BuyPrice    float64 `json:"buy_price" validate:"required,gte=0"`
	Fee         float64 `json:"fee"`
	BuyCurrency string  `json:"buy_currency"`
	Notes       string  `json:"notes"`
}

type UpdateHoldingRequest struct {
	Quantity    *float64 `json:"quantity" validate:"omitempty,gt=0"`
	BuyPrice    *float64 `json:"buy_price" validate:"omitempty,gte=0"`
	BuyCurrency *string  `json:"buy_currency"`
	Notes       *string  `json:"notes"`
}
