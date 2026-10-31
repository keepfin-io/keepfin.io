package models

// MarketPrice stores unadjusted daily closes. Dividends are recorded separately.
type MarketPrice struct {
	Symbol string  `gorm:"primaryKey;type:varchar(40)" json:"symbol"`
	Date   string  `gorm:"primaryKey;type:varchar(10)" json:"date"`
	Close  float64 `gorm:"type:decimal(24,10);not null" json:"close"`
}
