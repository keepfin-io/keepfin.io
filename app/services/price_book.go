package services

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"keepnode/app/models"
	"keepnode/pkg/portfolio"
	"sort"
	"time"
)

type priceBook struct {
	db       *gorm.DB
	market   *MarketService
	from, to time.Time
	series   map[string][]models.MarketPrice
	failures map[string]error
}

func newPriceBook(db *gorm.DB, market *MarketService, from, to time.Time) *priceBook {
	return &priceBook{db: db, market: market, from: portfolio.Day(from).AddDate(0, 0, -7), to: portfolio.Day(to), series: map[string][]models.MarketPrice{}, failures: map[string]error{}}
}
func (b *priceBook) load(symbol string) {
	if _, ok := b.series[symbol]; ok {
		return
	}
	rows := []models.MarketPrice{}
	if err := b.db.Where("symbol = ? AND date >= ? AND date <= ?", symbol, b.from.Format("2006-01-02"), b.to.Format("2006-01-02")).Order("date").Find(&rows).Error; err != nil {
		b.failures[symbol] = err
		b.series[symbol] = rows
		return
	}
	// Reuse persisted history; request only the missing tail. No requests per day.
	start := b.from
	if len(rows) > 0 {
		d, _ := time.Parse("2006-01-02", rows[len(rows)-1].Date)
		start = d.AddDate(0, 0, 1)
		first, _ := time.Parse("2006-01-02", rows[0].Date)
		if first.Sub(b.from) > 7*24*time.Hour {
			start = b.from
		}
	}
	if !start.After(b.to) && b.market != nil {
		data, err := b.market.GetHistoricalData(&models.EODRequest{Symbol: symbol, From: start.Format("2006-01-02"), To: b.to.Format("2006-01-02"), Period: "d", Order: "a"})
		if err != nil {
			b.failures[symbol] = err
		} else {
			added := []models.MarketPrice{}
			for _, d := range data {
				if portfolio.ValidNumber(d.Close) && d.Close > 0 {
					added = append(added, models.MarketPrice{Symbol: symbol, Date: d.Date, Close: d.Close})
				}
			}
			if len(added) > 0 {
				if err := b.db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(added, 500).Error; err != nil {
					b.failures[symbol] = err
				}
				rows = append(rows, added...)
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Date < rows[j].Date })
	b.series[symbol] = rows
}
func (b *priceBook) price(symbol string, date time.Time) (float64, string, error) {
	b.load(symbol)
	day := date.UTC().Format("2006-01-02")
	rows := b.series[symbol]
	i := sort.Search(len(rows), func(i int) bool { return rows[i].Date > day }) - 1
	if i >= 0 {
		d, err := time.Parse("2006-01-02", rows[i].Date)
		if err == nil && portfolio.Day(date).Sub(d) <= 7*24*time.Hour {
			return rows[i].Close, rows[i].Date, nil
		}
	}
	return 0, "", fmt.Errorf("price unavailable: %s (%s)", symbol, day)
}
func (b *priceBook) rate(from, to string, date time.Time) (float64, error) {
	if from == to {
		return 1, nil
	}
	v, _, err := b.price(from+to+".FOREX", date)
	if err == nil {
		return v, nil
	}
	inverse, _, e := b.price(to+from+".FOREX", date)
	if e == nil {
		return portfolio.Div(1, inverse), nil
	}
	return 0, fmt.Errorf("exchange rate unavailable: %s/%s on %s; enter the actual transaction rate", from, to, date.Format("2006-01-02"))
}
func resolveOpeningRates(db *gorm.DB, market *MarketService, account models.Account, events []models.Activity) error {
	for i := range events {
		a := &events[i]
		if !a.Ledger || a.ExchangeRate > 0 {
			continue
		}
		if a.Type != "opening" {
			return errors.New("activity exchange rate is missing")
		}
		rate, err := newPriceBook(db, market, a.Date, a.Date).rate(a.Currency, account.Currency, a.Date)
		if err != nil {
			return err
		}
		a.ExchangeRate = rate
		if err := db.Model(&models.Activity{}).Where("id = ? AND exchange_rate = 0", a.ID).Update("exchange_rate", rate).Error; err != nil {
			return err
		}
	}
	return nil
}
