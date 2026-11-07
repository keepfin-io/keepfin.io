package wealth

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	activityservice "keepnode/app/services/activity"
	"keepnode/pkg/portfolio"
	"math"
	"time"
)

type DividendPreview struct {
	ExDate   string                       `json:"ex_date"`
	Shares   float64                      `json:"shares"`
	PerShare *float64                     `json:"per_share"`
	Activity activityservice.JSONActivity `json:"activity"`
	Ready    bool                         `json:"ready"`
	Reason   string                       `json:"reason,omitempty"`
}

func (s *Service) PreviewDividends(user, account uuid.UUID, symbol, from, to string) ([]DividendPreview, error) {
	a, err := s.Account(user, account)
	if err != nil {
		return nil, err
	}
	if !symbolPattern.MatchString(symbol) || !validDate(from) || !validDate(to) || from > to {
		return nil, errors.New("invalid symbol or date range")
	}
	events, err := s.Activities.GetActivities(account)
	if err != nil {
		return nil, err
	}
	known := false
	for _, event := range events {
		if event.Symbol == symbol {
			known = true
			break
		}
	}
	if !known {
		return nil, errors.New("choose an asset with account activity")
	}
	dividends, err := s.Market.Dividends(symbol, from, to)
	if err != nil {
		return nil, err
	}
	out := []DividendPreview{}
	for _, d := range dividends {
		day, e := time.Parse("2006-01-02", d.Date)
		if e != nil || d.Date < from || d.Date > to {
			continue
		}
		state, e := portfolio.Replay(a, events, day.Add(-time.Nanosecond))
		if e != nil {
			return nil, e
		}
		p := state.Positions[symbol]
		if p == nil || p.Quantity <= 0 {
			continue
		}
		row := DividendPreview{ExDate: d.Date, Shares: p.Quantity, PerShare: d.UnadjustedValue}
		row.Activity = activityservice.JSONActivity{Reference: fmt.Sprintf("eodhd-dividend:%s:%s:%s", symbol, d.Date, d.Currency), Date: d.PaymentDate, Type: "dividend", Symbol: symbol, Quantity: 1, Currency: d.Currency, Notes: "EODHD dividend, ex-date " + d.Date}
		if d.UnadjustedValue == nil || !validAmount(*d.UnadjustedValue) || *d.UnadjustedValue <= 0 {
			row.Reason = "Unadjusted dividend amount is unavailable."
		} else if !validCurrency(d.Currency) {
			row.Reason = "Dividend currency is unavailable."
		} else if !validDate(d.PaymentDate) {
			row.Reason = "Verify a past payment date before importing."
		} else {
			row.Ready = true
		}
		if d.UnadjustedValue != nil {
			row.Activity.UnitPrice = math.Round(portfolio.Mul(p.Quantity, *d.UnadjustedValue)*1e4) / 1e4
		}
		out = append(out, row)
	}
	return out, nil
}
