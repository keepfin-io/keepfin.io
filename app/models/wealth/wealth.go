package wealth

import (
	"github.com/google/uuid"
	"time"
)

// These records are always scoped by UserID at the service boundary.
type WatchItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:watch_owner_symbol" json:"-"`
	Symbol     string    `gorm:"size:20;not null;uniqueIndex:watch_owner_symbol" json:"symbol"`
	Name       string    `gorm:"size:100" json:"name"`
	Currency   string    `gorm:"size:3" json:"currency"`
	AssetClass string    `gorm:"size:40" json:"asset_class"`
	CreatedAt  time.Time `json:"created_at"`
}

type ManualAsset struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Symbol     string    `gorm:"size:20;not null;uniqueIndex" json:"symbol"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Currency   string    `gorm:"size:3;not null" json:"currency"`
	AssetClass string    `gorm:"size:40" json:"asset_class"`
	Sector     string    `gorm:"size:100" json:"sector"`
	Country    string    `gorm:"size:100" json:"country"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ManualValuation struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AssetID   uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:manual_asset_date" json:"asset_id"`
	Asset     ManualAsset `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"-"`
	Date      string      `gorm:"type:varchar(10);not null;uniqueIndex:manual_asset_date" json:"date"`
	Price     float64     `gorm:"type:numeric(24,8);not null" json:"price"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Liability struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Currency  string    `gorm:"size:3;not null" json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type LiabilityBalance struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LiabilityID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:liability_date" json:"liability_id"`
	Liability   Liability `gorm:"foreignKey:LiabilityID;constraint:OnDelete:CASCADE" json:"-"`
	Date        string    `gorm:"type:varchar(10);not null;uniqueIndex:liability_date" json:"date"`
	Amount      float64   `gorm:"type:numeric(24,8);not null" json:"amount"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:tag_owner_name" json:"-"`
	Name   string    `gorm:"size:40;not null;uniqueIndex:tag_owner_name" json:"name"`
}
type TagAssignment struct {
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	TagID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"tag_id"`
	Tag      Tag       `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE" json:"-"`
	EntityID uuid.UUID `gorm:"type:uuid;primaryKey" json:"entity_id"`
	Kind     string    `gorm:"size:12;primaryKey" json:"kind"`
}

type Preferences struct {
	UserID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	MaxAssetWeight      float64   `json:"max_asset_weight"`
	MaxAccountWeight    float64   `json:"max_account_weight"`
	MaxCurrencyWeight   float64   `json:"max_currency_weight"`
	MaxFeeRatio         float64   `json:"max_fee_ratio"`
	EmergencyMonths     float64   `json:"emergency_months"`
	EmergencyFund       float64   `json:"emergency_fund"`
	MonthlyExpenses     float64   `json:"monthly_expenses"`
	MonthlyContribution float64   `json:"monthly_contribution"`
	AnnualExpenses      float64   `json:"annual_expenses"`
	AnnualReturn        float64   `json:"annual_return"`
	WithdrawalRate      float64   `json:"withdrawal_rate"`
	HorizonYears        int       `json:"horizon_years"`
	DisabledRules       []string  `gorm:"serializer:json;type:jsonb" json:"disabled_rules"`
	Zen                 bool      `json:"zen"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Share struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"-"`
	Name        string      `gorm:"size:100" json:"name"`
	TokenHash   string      `gorm:"size:64;uniqueIndex;not null" json:"-"`
	AccountIDs  []uuid.UUID `gorm:"serializer:json;type:jsonb;not null" json:"account_ids"`
	RecipientID *uuid.UUID  `gorm:"type:uuid" json:"recipient_id"`
	ExpiresAt   time.Time   `gorm:"not null" json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	RevokedAt   *time.Time  `json:"revoked_at"`
}

type MarketJob struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Symbol      string    `gorm:"size:20;not null" json:"symbol"`
	From        string    `gorm:"type:varchar(10)" json:"from"`
	To          string    `gorm:"type:varchar(10)" json:"to"`
	Status      string    `gorm:"size:16;index;not null" json:"status"`
	Attempts    int       `json:"attempts"`
	Error       string    `gorm:"size:255" json:"error"`
	AvailableAt time.Time `gorm:"index" json:"available_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MarketConfig struct {
	ID        int       `gorm:"primaryKey" json:"-"`
	Provider  string    `gorm:"size:32;not null" json:"provider"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Explicit corrections survive provider refreshes and retain their editor/reason.
type PriceOverride struct {
	Symbol    string    `gorm:"primaryKey;size:40" json:"symbol"`
	Date      string    `gorm:"primaryKey;type:varchar(10)" json:"date"`
	Close     float64   `gorm:"type:numeric(24,10);not null" json:"close"`
	Reason    string    `gorm:"size:500;not null" json:"reason"`
	EditorID  uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	UpdatedAt time.Time `json:"updated_at"`
}
