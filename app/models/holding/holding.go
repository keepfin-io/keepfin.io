package holding

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	basemodel "keepnode/app/models/base"
)

type Holding struct {
	basemodel.Base
	ID          uuid.UUID            `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AccountID   uuid.UUID            `gorm:"type:uuid;not null;index" json:"account_id" validate:"required"`
	Symbol      string               `gorm:"type:varchar(20);not null" json:"symbol" validate:"required,lte=20"`
	Name        string               `gorm:"type:varchar(100)" json:"name"`
	Quantity    float64              `gorm:"type:decimal(18,8);not null" json:"quantity" validate:"required,gt=0"`
	BuyPrice    float64              `gorm:"type:decimal(18,4);not null" json:"buy_price" validate:"required,gte=0"`
	BuyCurrency string               `gorm:"type:varchar(3);default:'USD'" json:"buy_currency"`
	Notes       string               `gorm:"type:text" json:"notes,omitempty"`
	AssetClass  string               `gorm:"type:varchar(40)" json:"asset_class"`
	Sector      string               `gorm:"type:varchar(100)" json:"sector"`
	Country     string               `gorm:"type:varchar(100)" json:"country"`
	Account     accountmodel.Account `gorm:"foreignKey:AccountID" json:"-"`
}

func (Holding) TableName() string {
	return "holdings"
}

func (h *Holding) BeforeCreate(tx *gorm.DB) error {
	return h.Base.BeforeCreate(tx)
}

type HoldingWithLiveData struct {
	ID              uuid.UUID `json:"id"`
	AccountID       uuid.UUID `json:"account_id"`
	Symbol          string    `json:"symbol"`
	Name            string    `json:"name"`
	Quantity        float64   `json:"quantity"`
	BuyPrice        float64   `json:"buy_price"`
	BuyCurrency     string    `json:"buy_currency"`
	QuoteCurrency   string    `json:"quote_currency"`
	CurrentPrice    float64   `json:"current_price"`
	Value           float64   `json:"value"`
	Change          float64   `json:"change"`
	ChangePercent   float64   `json:"change_percent"`
	Performance     float64   `json:"performance"`
	ProfitLoss      float64   `json:"profit_loss"`
	FirstActivity   time.Time `json:"first_activity"`
	PriceAvailable  bool      `json:"price_available"`
	EntryPrice      float64   `json:"entry_price"`
	CanEditPosition bool      `json:"can_edit_position"`
	PriceDate       string    `json:"price_date"`
	AssetClass      string    `json:"asset_class"`
	Sector          string    `json:"sector"`
	Country         string    `json:"country"`
}

type HoldingsSummary struct {
	TotalValue         float64  `json:"total_value"`
	TotalCost          float64  `json:"total_cost"`
	TotalProfitLoss    float64  `json:"total_profit_loss"`
	TotalPerformance   float64  `json:"total_performance"`
	DailyChange        float64  `json:"daily_change"`
	CashBalance        float64  `json:"cash_balance"`
	RealizedProfitLoss float64  `json:"realized_profit_loss"`
	Dividend           float64  `json:"dividend"`
	Interest           float64  `json:"interest"`
	Fees               float64  `json:"fees"`
	Complete           bool     `json:"complete"`
	Warnings           []string `json:"warnings"`
}

type CreateHoldingRequest struct {
	Symbol       string  `json:"symbol" validate:"required,lte=20"`
	Name         string  `json:"name"`
	AssetClass   string  `json:"asset_class"`
	Country      string  `json:"country"`
	Quantity     float64 `json:"quantity" validate:"required,gt=0"`
	BuyPrice     float64 `json:"buy_price" validate:"required,gte=0"`
	Fee          float64 `json:"fee"`
	BuyCurrency  string  `json:"buy_currency"`
	Notes        string  `json:"notes"`
	Date         string  `json:"date"`
	ExchangeRate float64 `json:"exchange_rate"`
}

type UpdateHoldingRequest struct {
	Quantity    *float64 `json:"quantity" validate:"omitempty,gt=0"`
	BuyPrice    *float64 `json:"buy_price" validate:"omitempty,gte=0"`
	BuyCurrency *string  `json:"buy_currency"`
	Notes       *string  `json:"notes"`
	AssetClass  *string  `json:"asset_class"`
	Sector      *string  `json:"sector"`
	Country     *string  `json:"country"`
}
