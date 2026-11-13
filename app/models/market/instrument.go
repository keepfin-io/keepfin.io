package market

import "time"

// Instrument caches the currency in which a provider quotes a symbol's prices.
// An empty Currency records that the provider could not resolve the symbol;
// such rows are retried after a cool-down instead of on every calculation.
type Instrument struct {
	Symbol    string    `gorm:"primaryKey;type:varchar(40)" json:"symbol"`
	Currency  string    `gorm:"type:varchar(8);not null;default:''" json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Instrument) TableName() string {
	return "market_instruments"
}
