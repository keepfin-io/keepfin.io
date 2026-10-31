package portfolio

import (
	"fmt"
	"keepnode/app/models"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"
)

type Position struct {
	Symbol, Currency, Name      string
	Quantity, Cost, AccountCost float64
	FirstActivity               time.Time
}
type State struct {
	Cash, Realized, Dividends, Interest, Fees, NetDeposits float64
	Positions                                              map[string]*Position
}

// Arithmetic is performed as decimal rationals and rounded to eight decimal
// places at ledger boundaries. Existing JSON number contracts remain unchanged.
func decimal(v float64) *big.Rat {
	r, _ := new(big.Rat).SetString(strconv.FormatFloat(v, 'f', -1, 64))
	return r
}
func rounded(r *big.Rat) float64 { v, _ := strconv.ParseFloat(r.FloatString(8), 64); return v }
func Add(a, b float64) float64   { return rounded(new(big.Rat).Add(decimal(a), decimal(b))) }
func Mul(a, b float64) float64   { return rounded(new(big.Rat).Mul(decimal(a), decimal(b))) }
func Div(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return rounded(new(big.Rat).Quo(decimal(a), decimal(b)))
}
func Proportion(total, part, whole float64) float64 {
	if whole == 0 {
		return 0
	}
	return rounded(new(big.Rat).Quo(new(big.Rat).Mul(decimal(total), decimal(part)), decimal(whole)))
}
func ValidNumber(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) <= 1e12 }
func Day(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

// Replay is the single source of truth for positions, weighted-average cost,
// realized gains and settled cash. Archived pre-migration activities are inert.
func Replay(account models.Account, activities []models.Activity, until time.Time) (*State, error) {
	s := &State{Positions: map[string]*Position{}}
	if account.OpeningDate != nil && Day(until).Before(Day(*account.OpeningDate)) {
		return s, nil
	}
	s.Cash = account.InitialBalance
	s.NetDeposits = account.InitialBalance
	events := append([]models.Activity(nil), activities...)
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Date.Equal(events[j].Date) {
			if events[i].Type == "opening" && events[j].Type != "opening" {
				return true
			}
			if events[j].Type == "opening" && events[i].Type != "opening" {
				return false
			}
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		return events[i].Date.Before(events[j].Date)
	})
	for _, a := range events {
		if !a.Ledger || a.Date.After(until) {
			continue
		}
		if !ValidNumber(a.Quantity) || !ValidNumber(a.UnitPrice) || !ValidNumber(a.Fee) || !ValidNumber(a.ExchangeRate) || a.Quantity <= 0 || a.UnitPrice < 0 || a.Fee < 0 || a.ExchangeRate <= 0 {
			return nil, fmt.Errorf("invalid amount or exchange rate for %s on %s", a.Symbol, a.Date.Format("2006-01-02"))
		}
		gross := Mul(a.Quantity, a.UnitPrice)
		fee := Mul(a.Fee, a.ExchangeRate)
		amount := Mul(gross, a.ExchangeRate)
		s.Fees = Add(s.Fees, fee)
		switch a.Type {
		case "buy", "opening", "sell":
			p := s.Positions[a.Symbol]
			if p == nil {
				p = &Position{Symbol: a.Symbol, Currency: a.Currency, Name: a.Name, FirstActivity: a.Date}
				s.Positions[a.Symbol] = p
			}
			if p.Currency != a.Currency {
				return nil, fmt.Errorf("%s already uses %s; use its trading currency", a.Symbol, p.Currency)
			}
			if a.Type == "sell" {
				if a.Quantity > p.Quantity {
					return nil, fmt.Errorf("insufficient %s quantity on %s", a.Symbol, a.Date.Format("2006-01-02"))
				}
				cost := Proportion(p.Cost, a.Quantity, p.Quantity)
				accountCost := Proportion(p.AccountCost, a.Quantity, p.Quantity)
				if a.Quantity == p.Quantity {
					cost = p.Cost
					accountCost = p.AccountCost
				}
				p.Quantity = Add(p.Quantity, -a.Quantity)
				p.Cost = Add(p.Cost, -cost)
				p.AccountCost = Add(p.AccountCost, -accountCost)
				s.Realized = Add(s.Realized, Add(amount, -fee)-accountCost)
				s.Cash = Add(s.Cash, Add(amount, -fee))
			} else {
				p.Quantity = Add(p.Quantity, a.Quantity)
				p.Cost = Add(p.Cost, Add(gross, a.Fee))
				p.AccountCost = Add(p.AccountCost, Add(amount, fee))
				if a.Type == "buy" {
					s.Cash = Add(s.Cash, -Add(amount, fee))
				} else {
					s.NetDeposits = Add(s.NetDeposits, Add(amount, fee))
				}
			}
		case "deposit", "transfer_in":
			s.Cash = Add(s.Cash, amount-fee)
			s.NetDeposits = Add(s.NetDeposits, amount)
		case "withdrawal", "transfer_out":
			s.Cash = Add(s.Cash, -Add(amount, fee))
			s.NetDeposits = Add(s.NetDeposits, -amount)
		case "dividend":
			s.Cash = Add(s.Cash, amount-fee)
			s.Dividends = Add(s.Dividends, amount)
		case "interest":
			s.Cash = Add(s.Cash, amount-fee)
			s.Interest = Add(s.Interest, amount)
		case "fee":
			s.Cash = Add(s.Cash, -Add(amount, fee))
			s.Fees = Add(s.Fees, amount)
		default:
			return nil, fmt.Errorf("unsupported activity type: %s", a.Type)
		}
	}
	return s, nil
}
