package market

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	marketmodel "keepnode/app/models/market"
	wealthmodel "keepnode/app/models/wealth"
)

const (
	// unresolvedRetry limits provider lookups for symbols the provider did not list.
	unresolvedRetry = 24 * time.Hour
	// failedLookupRetry limits provider lookups while the provider is failing.
	failedLookupRetry = 10 * time.Minute
)

// minorUnits are quote units that providers use for some listings, such as
// pence on the London Stock Exchange. Users transact in the major currency.
var minorUnits = map[string]struct {
	major  string
	factor float64
}{
	"GBX": {"GBP", 0.01},
	"ZAC": {"ZAR", 0.01},
	"ILA": {"ILS", 0.01},
}

func normalizeQuoteCurrency(raw string) string {
	raw = strings.TrimSpace(raw)
	switch raw {
	case "GBp":
		return "GBX"
	case "ZAc":
		return "ZAC"
	}
	return strings.ToUpper(raw)
}

// MajorCurrency maps a minor quote unit such as GBX to the currency that
// activities use, so a pence-quoted listing still expects GBP transactions.
func MajorCurrency(currency string) string {
	if m, ok := minorUnits[currency]; ok {
		return m.major
	}
	return currency
}

// QuoteCurrency returns the currency in which symbol's market prices are
// quoted, or "" when it is unknown. Manual assets use their own currency, FX
// pairs have none, and provider answers are cached in market_instruments.
func QuoteCurrency(db *gorm.DB, market *MarketService, symbol string) string {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if db == nil || symbol == "" || strings.HasSuffix(symbol, ".FOREX") {
		return ""
	}
	if strings.HasPrefix(symbol, "MANUAL-") {
		assets := []wealthmodel.ManualAsset{}
		if db.Select("currency").Where("symbol = ?", symbol).Limit(1).Find(&assets).Error != nil || len(assets) == 0 {
			return ""
		}
		return assets[0].Currency
	}
	cached := []marketmodel.Instrument{}
	if db.Where("symbol = ?", symbol).Limit(1).Find(&cached).Error == nil && len(cached) == 1 {
		if cached[0].Currency != "" || time.Since(cached[0].UpdatedAt) < unresolvedRetry {
			return cached[0].Currency
		}
	}
	if market == nil {
		return ""
	}
	currency, ok := market.lookupQuoteCurrency(symbol)
	if !ok {
		return ""
	}
	// The table is only a cache; a failed write must not fail the calculation.
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&marketmodel.Instrument{Symbol: symbol, Currency: currency})
	return currency
}

// lookupQuoteCurrency asks the active provider which currency quotes symbol.
// ok is false when the provider could not be asked; an unlisted symbol
// returns "" with ok true so the miss can be cached.
func (s *MarketService) lookupQuoteCurrency(symbol string) (string, bool) {
	if !s.Configured(s.Provider()) {
		return "", false
	}
	s.mu.RLock()
	failedAt, failed := s.failedLookups[symbol]
	s.mu.RUnlock()
	if failed && time.Since(failedAt) < failedLookupRetry {
		return "", false
	}
	code, exchange := symbol, "US"
	if i := strings.LastIndex(symbol, "."); i > 0 {
		code, exchange = symbol[:i], symbol[i+1:]
	}
	results, err := s.SearchAssets(&marketmodel.SearchRequest{Query: code, Limit: 50})
	s.mu.Lock()
	if err != nil {
		if s.failedLookups == nil {
			s.failedLookups = map[string]time.Time{}
		}
		s.failedLookups[symbol] = time.Now()
	} else {
		delete(s.failedLookups, symbol)
	}
	s.mu.Unlock()
	if err != nil {
		return "", false
	}
	for _, r := range results {
		if strings.EqualFold(r.Code, code) && strings.EqualFold(r.Exchange, exchange) {
			return normalizeQuoteCurrency(r.Currency), true
		}
	}
	return "", true
}
