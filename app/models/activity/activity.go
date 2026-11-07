package activity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	basemodel "keepnode/app/models/base"
	holdingmodel "keepnode/app/models/holding"
)

// ActivityType constants
const (
	ActivityTypeBuy         = "buy"
	ActivityTypeSell        = "sell"
	ActivityTypeDividend    = "dividend"
	ActivityTypeInterest    = "interest"
	ActivityTypeDeposit     = "deposit"
	ActivityTypeWithdrawal  = "withdrawal"
	ActivityTypeFee         = "fee"
	ActivityTypeSplit       = "split"
	ActivityTypeOpening     = "opening"
	ActivityTypeTransferIn  = "transfer_in"
	ActivityTypeTransferOut = "transfer_out"
)

type Activity struct {
	SplitNumerator   int64 `json:"split_numerator"`
	SplitDenominator int64 `json:"split_denominator"`
	basemodel.Base
	ID           uuid.UUID             `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AccountID    uuid.UUID             `gorm:"type:uuid;not null;index;uniqueIndex:activity_import,priority:1" json:"account_id" validate:"required"`
	HoldingID    *uuid.UUID            `gorm:"type:uuid;index" json:"holding_id,omitempty"`
	Symbol       string                `gorm:"type:varchar(20);not null" json:"symbol" validate:"required,lte=20"`
	Name         string                `gorm:"type:varchar(100)" json:"name"`
	Type         string                `gorm:"type:varchar(20);not null" json:"type" validate:"required,oneof=buy sell dividend interest deposit withdrawal"`
	Date         time.Time             `gorm:"not null;index" json:"date" validate:"required"`
	Quantity     float64               `gorm:"type:decimal(18,8);not null" json:"quantity" validate:"required"`
	UnitPrice    float64               `gorm:"type:decimal(18,4);not null" json:"unit_price" validate:"required,gte=0"`
	Fee          float64               `gorm:"type:decimal(18,4);default:0" json:"fee"`
	Currency     string                `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Notes        string                `gorm:"type:text" json:"notes,omitempty"`
	Total        float64               `gorm:"-" json:"total"`
	Ledger       bool                  `gorm:"not null;default:false;index" json:"ledger"`
	ExchangeRate float64               `gorm:"type:decimal(24,10);not null;default:1" json:"exchange_rate"`
	TransferID   *uuid.UUID            `gorm:"type:uuid;index" json:"transfer_id,omitempty"`
	ImportKey    *string               `gorm:"uniqueIndex:activity_import,priority:2" json:"-"`
	Account      accountmodel.Account  `gorm:"foreignKey:AccountID" json:"-"`
	Holding      *holdingmodel.Holding `gorm:"foreignKey:HoldingID" json:"-"`
}

func (Activity) TableName() string {
	return "activities"
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	return a.Base.BeforeCreate(tx)
}

func (a *Activity) AfterFind(tx *gorm.DB) error {
	a.Total = (a.Quantity * a.UnitPrice) + a.Fee
	if a.Type == "sell" || a.Type == "dividend" || a.Type == "interest" || a.Type == "deposit" || a.Type == "transfer_in" {
		a.Total = a.Quantity*a.UnitPrice - a.Fee
	}
	return nil
}

type CreateActivityRequest struct {
	SplitNumerator   int64   `json:"split_numerator"`
	SplitDenominator int64   `json:"split_denominator"`
	Symbol           string  `json:"symbol" validate:"required,lte=20"`
	Type             string  `json:"type" validate:"required,oneof=buy sell dividend interest deposit withdrawal"`
	Date             string  `json:"date" validate:"required"`
	Quantity         float64 `json:"quantity" validate:"required"`
	UnitPrice        float64 `json:"unit_price" validate:"required,gte=0"`
	Fee              float64 `json:"fee"`
	Currency         string  `json:"currency"`
	Notes            string  `json:"notes"`
	ExchangeRate     float64 `json:"exchange_rate"`
}

func ValidActivityTypes() []string {
	return []string{
		ActivityTypeBuy,
		ActivityTypeSell,
		ActivityTypeDividend,
		ActivityTypeInterest,
		ActivityTypeDeposit,
		ActivityTypeWithdrawal,
		ActivityTypeFee,
		ActivityTypeSplit,
	}
}
