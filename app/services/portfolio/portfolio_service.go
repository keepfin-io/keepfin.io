package portfolio

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	holdingmodel "keepnode/app/models/holding"
	usermodel "keepnode/app/models/user"
	activityquery "keepnode/app/queries/activity"
	marketservice "keepnode/app/services/market"
	"keepnode/pkg/portfolio"
)

type PortfolioService struct {
	DB     *gorm.DB
	market *marketservice.MarketService
}

func NewPortfolioService(db *gorm.DB, market *marketservice.MarketService) *PortfolioService {
	return &PortfolioService{db, market}
}

type PortfolioHolding struct {
	Symbol     string   `json:"symbol"`
	Name       string   `json:"name"`
	Currency   string   `json:"currency"`
	Quantity   float64  `json:"quantity"`
	Value      float64  `json:"value"`
	Cost       float64  `json:"cost"`
	ProfitLoss float64  `json:"profit_loss"`
	Weight     float64  `json:"weight"`
	Price      float64  `json:"price"`
	PriceDate  string   `json:"price_date"`
	AssetClass string   `json:"asset_class"`
	Sector     string   `json:"sector"`
	Country    string   `json:"country"`
	Complete   bool     `json:"complete"`
	Accounts   []string `json:"accounts"`
}
type PortfolioSummary struct {
	Value            float64  `json:"value"`
	Equity           float64  `json:"equity"`
	Cash             float64  `json:"cash"`
	Cost             float64  `json:"cost"`
	NetContributions float64  `json:"net_contributions"`
	ProfitLoss       float64  `json:"profit_loss"`
	Realized         float64  `json:"realized"`
	Unrealized       float64  `json:"unrealized"`
	Dividends        float64  `json:"dividends"`
	Interest         float64  `json:"interest"`
	Fees             float64  `json:"fees"`
	DailyChange      *float64 `json:"daily_change"`
}
type HistoryPoint struct {
	Date      string   `json:"date"`
	Value     *float64 `json:"value"`
	NetFlow   float64  `json:"net_flow"`
	Return    *float64 `json:"return"`
	Benchmark *float64 `json:"benchmark"`
}
type IncomePeriod struct {
	Period    string  `json:"period"`
	Dividends float64 `json:"dividends"`
	Interest  float64 `json:"interest"`
	Fees      float64 `json:"fees"`
	Net       float64 `json:"net"`
}
type Allocation struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Weight float64 `json:"weight"`
}
type PortfolioReport struct {
	Currency    string                  `json:"currency"`
	Summary     PortfolioSummary        `json:"summary"`
	Holdings    []PortfolioHolding      `json:"holdings"`
	History     []HistoryPoint          `json:"history"`
	Income      []IncomePeriod          `json:"income"`
	Allocations map[string][]Allocation `json:"allocations"`
	Warnings    []string                `json:"warnings"`
	Complete    bool                    `json:"complete"`
	From        string                  `json:"from"`
	To          string                  `json:"to"`
	Method      string                  `json:"method"`
	Benchmark   string                  `json:"benchmark"`
}
type accountLedger struct {
	account  accountmodel.Account
	events   []activitymodel.Activity
	holdings map[string]holdingmodel.Holding
}

func addWarning(r *PortfolioReport, message string) {
	for _, s := range r.Warnings {
		if s == message {
			return
		}
	}
	if len(r.Warnings) < 30 {
		r.Warnings = append(r.Warnings, message)
	}
}
func reportStart(period string, end, earliest time.Time) (time.Time, error) {
	switch period {
	case "1d":
		return end.AddDate(0, 0, -1), nil
	case "1m":
		return end.AddDate(0, -1, 0), nil
	case "3m":
		return end.AddDate(0, -3, 0), nil
	case "ytd":
		return time.Date(end.Year(), 1, 1, 0, 0, 0, 0, time.UTC), nil
	case "1y":
		return end.AddDate(-1, 0, 0), nil
	case "5y":
		return end.AddDate(-5, 0, 0), nil
	case "max", "":
		return earliest, nil
	default:
		return time.Time{}, errors.New("invalid range")
	}
}
func (s *PortfolioService) Report(userID uuid.UUID, accountID *uuid.UUID, period, benchmark string) (*PortfolioReport, error) {
	r := &PortfolioReport{Holdings: []PortfolioHolding{}, History: []HistoryPoint{}, Income: []IncomePeriod{}, Allocations: map[string][]Allocation{}, Warnings: []string{}, Complete: true, Method: "daily-linked Modified Dietz (mid-day external cash flows)", Benchmark: benchmark}
	currency, ledgers, err := s.loadLedgers(userID, accountID)
	if err != nil {
		return nil, err
	}
	r.Currency = currency
	end := portfolio.Day(time.Now())
	earliest := end
	for _, l := range ledgers {
		if l.account.OpeningDate != nil && portfolio.Day(*l.account.OpeningDate).Before(earliest) {
			earliest = portfolio.Day(*l.account.OpeningDate)
		}
		if err = marketservice.ResolveOpeningRates(s.DB, s.market, l.account, l.events); err != nil {
			return nil, err
		}
	}
	start, err := reportStart(period, end, earliest)
	if err != nil {
		return nil, err
	}
	if start.Before(earliest) {
		start = earliest
	}
	if end.Sub(start) > 3660*24*time.Hour {
		return nil, errors.New("choose a range of at most ten years")
	}
	r.From = start.Format("2006-01-02")
	r.To = end.Format("2006-01-02")
	book := marketservice.NewPriceBook(s.DB, s.market, earliest, end)
	transferCounts := map[uuid.UUID]int{}
	for _, l := range ledgers {
		for _, a := range l.events {
			if a.TransferID != nil {
				transferCounts[*a.TransferID]++
			}
		}
	}
	internalTransfer := func(a activitymodel.Activity) bool { return a.TransferID != nil && transferCounts[*a.TransferID] == 2 }
	grouped := map[string]*PortfolioHolding{}
	allocationValues := map[string]map[string]float64{}
	income := map[string]*IncomePeriod{}
	allocation := func(kind, name string, value float64) {
		if name == "" {
			name = "Unclassified"
		}
		if allocationValues[kind] == nil {
			allocationValues[kind] = map[string]float64{}
		}
		allocationValues[kind][name] += value
	}
	// End-of-day totals use transaction-date FX for cash flows/costs and current FX for valuations.
	for _, l := range ledgers {
		state, err := portfolio.Replay(l.account, l.events, end.Add(24*time.Hour-time.Nanosecond))
		if err != nil {
			return nil, err
		}
		fx, err := book.Rate(l.account.Currency, r.Currency, end)
		if err != nil {
			r.Complete = false
			addWarning(r, err.Error())
			continue
		}
		r.Summary.Cash += state.Cash * fx
		baseAccount := l.account
		baseAccount.Currency = r.Currency
		baseAccount.InitialBalance = 0
		baseEvents := append([]activitymodel.Activity(nil), l.events...)
		baseValid := true
		for i := range baseEvents {
			rate, e := book.Rate(l.account.Currency, r.Currency, baseEvents[i].Date)
			if e != nil {
				baseValid = false
				r.Complete = false
				addWarning(r, e.Error())
				break
			}
			baseEvents[i].ExchangeRate = portfolio.Mul(baseEvents[i].ExchangeRate, rate)
		}
		baseState := state
		if baseValid {
			baseState, err = portfolio.Replay(baseAccount, baseEvents, end.Add(24*time.Hour-time.Nanosecond))
			if err != nil {
				return nil, err
			}
			r.Summary.Realized += baseState.Realized
		}
		allocation("currency", l.account.Currency, state.Cash*fx)
		allocation("asset_class", "Cash", state.Cash*fx)
		allocation("account", l.account.AccountName, state.Cash*fx)
		allocation("sector", "Cash", state.Cash*fx)
		allocation("country", "Cash", state.Cash*fx)
		// Initial cash is a contribution on the account opening date.
		if l.account.OpeningDate != nil {
			rate, e := book.Rate(l.account.Currency, r.Currency, *l.account.OpeningDate)
			if e != nil {
				r.Complete = false
				addWarning(r, e.Error())
			} else {
				r.Summary.NetContributions += l.account.InitialBalance * rate
			}
		}
		for symbol, p := range state.Positions {
			if p.Quantity <= 0 {
				continue
			}
			meta := l.holdings[symbol]
			key := symbol + "/" + p.Currency
			h := grouped[key]
			if h == nil {
				h = &PortfolioHolding{Symbol: symbol, Name: meta.Name, Currency: p.Currency, AssetClass: meta.AssetClass, Sector: meta.Sector, Country: meta.Country, Complete: true, Accounts: []string{}}
				grouped[key] = h
			}
			h.Quantity += p.Quantity
			h.Accounts = append(h.Accounts, l.account.AccountName)
			if baseValid {
				cost := baseState.Positions[symbol].AccountCost
				h.Cost += cost
				r.Summary.Cost += cost
			} else {
				h.Complete = false
			}
			price, date, e := book.Price(symbol, end)
			rate, fe := book.Rate(p.Currency, r.Currency, end)
			if e != nil || fe != nil {
				h.Complete = false
				r.Complete = false
				if e != nil {
					addWarning(r, e.Error())
				}
				if fe != nil {
					addWarning(r, fe.Error())
				}
				continue
			}
			h.Price = price
			h.PriceDate = date
			value := p.Quantity * price * rate
			h.Value += value
			r.Summary.Equity += value
			allocation("account", l.account.AccountName, value)
			allocation("asset_class", meta.AssetClass, value)
			allocation("currency", p.Currency, value)
			allocation("sector", meta.Sector, value)
			allocation("country", meta.Country, value)
		}
		for _, a := range l.events {
			rate, e := book.Rate(l.account.Currency, r.Currency, a.Date)
			if e != nil {
				r.Complete = false
				addWarning(r, e.Error())
				continue
			}
			amount := portfolio.Mul(a.Quantity, a.UnitPrice) * a.ExchangeRate * rate
			fee := a.Fee * a.ExchangeRate * rate
			if !internalTransfer(a) {
				switch a.Type {
				case "deposit", "transfer_in", "opening":
					r.Summary.NetContributions += amount
				case "withdrawal", "transfer_out":
					r.Summary.NetContributions -= amount
				}
			}
			if a.Date.Before(start) || a.Date.After(end.Add(24*time.Hour-time.Nanosecond)) {
				continue
			}
			month := a.Date.Format("2006-01")
			row := income[month]
			if row == nil {
				row = &IncomePeriod{Period: month}
				income[month] = row
			}
			row.Fees += fee
			switch a.Type {
			case "dividend":
				row.Dividends += amount
				r.Summary.Dividends += amount
			case "interest":
				row.Interest += amount
				r.Summary.Interest += amount
			case "fee":
				row.Fees += amount
			}
		}
	}
	for _, row := range income {
		row.Net = row.Dividends + row.Interest - row.Fees
		r.Summary.Fees += row.Fees
		r.Income = append(r.Income, *row)
	}
	sort.Slice(r.Income, func(i, j int) bool { return r.Income[i].Period < r.Income[j].Period })
	r.Summary.Value = r.Summary.Cash + r.Summary.Equity
	r.Summary.Unrealized = r.Summary.Equity - r.Summary.Cost
	r.Summary.ProfitLoss = r.Summary.Value - r.Summary.NetContributions
	for _, h := range grouped {
		h.ProfitLoss = h.Value - h.Cost
		if r.Summary.Value > 0 {
			h.Weight = h.Value / r.Summary.Value * 100
		}
		r.Holdings = append(r.Holdings, *h)
	}
	sort.Slice(r.Holdings, func(i, j int) bool { return r.Holdings[i].Value > r.Holdings[j].Value })
	for kind, values := range allocationValues {
		rows := []Allocation{}
		for name, value := range values {
			weight := 0.0
			if r.Summary.Value > 0 {
				weight = value / r.Summary.Value * 100
			}
			rows = append(rows, Allocation{Name: name, Value: value, Weight: weight})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].Value > rows[j].Value })
		r.Allocations[kind] = rows
	}
	if benchmark != "" && benchmark != "SPY.US" && benchmark != "VT.US" && benchmark != "QQQ.US" {
		return nil, errors.New("invalid benchmark symbol")
	}
	var previous *float64
	linked := 1.0
	continuous := true
	benchmarkStart := 0.0
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		point := HistoryPoint{Date: day.Format("2006-01-02")}
		value := 0.0
		valid := true
		for _, l := range ledgers {
			if l.account.OpeningDate != nil && day.Before(portfolio.Day(*l.account.OpeningDate)) {
				continue
			}
			state, e := portfolio.Replay(l.account, l.events, day.Add(24*time.Hour-time.Nanosecond))
			if e != nil {
				return nil, e
			}
			fx, e := book.Rate(l.account.Currency, r.Currency, day)
			if e != nil {
				valid = false
				continue
			}
			value += state.Cash * fx
			if l.account.OpeningDate != nil && day.Equal(portfolio.Day(*l.account.OpeningDate)) {
				point.NetFlow += l.account.InitialBalance * fx
			}
			for _, p := range state.Positions {
				if p.Quantity <= 0 {
					continue
				}
				price, _, e := book.Price(p.Symbol, day)
				rate, fe := book.Rate(p.Currency, r.Currency, day)
				if e != nil || fe != nil {
					valid = false
					continue
				}
				value += p.Quantity * price * rate
			}
			for _, a := range l.events {
				if !portfolio.Day(a.Date).Equal(day) {
					continue
				}
				sign := 0.0
				switch a.Type {
				case "deposit", "transfer_in", "opening":
					sign = 1
				case "withdrawal", "transfer_out":
					sign = -1
				}
				if sign != 0 && !internalTransfer(a) {
					rate, e := book.Rate(l.account.Currency, r.Currency, day)
					if e != nil {
						valid = false
						continue
					}
					point.NetFlow += sign * a.Quantity * a.UnitPrice * a.ExchangeRate * rate
				}
			}
		}
		if valid {
			v := value
			point.Value = &v
			if previous == nil {
				if len(r.History) == 0 {
					v := 0.0
					point.Return = &v
				} else {
					continuous = false
				}
			} else {
				denom := *previous + point.NetFlow/2
				if continuous && denom > 0 {
					daily := (value - *previous - point.NetFlow) / denom
					linked *= 1 + daily
					v := (linked - 1) * 100
					if portfolio.ValidNumber(v) {
						point.Return = &v
					} else {
						continuous = false
					}
				} else if value == 0 && *previous == 0 && point.NetFlow == 0 && continuous {
					v := (linked - 1) * 100
					point.Return = &v
				} else {
					continuous = false
				}
				if day.Equal(end) {
					v := value - *previous - point.NetFlow
					r.Summary.DailyChange = &v
				}
			}
			previous = &v
		} else {
			previous = nil
			continuous = false
			addWarning(r, "Some historical prices or exchange rates are missing; gaps are not treated as zero returns.")
		}
		if benchmark != "" {
			price, _, e := book.Price(benchmark, day)
			rate, fxErr := book.Rate("USD", r.Currency, day)
			if e == nil && fxErr == nil {
				price *= rate
				if day.Equal(start) {
					benchmarkStart = price
				}
				if benchmarkStart > 0 {
					v := (price/benchmarkStart - 1) * 100
					point.Benchmark = &v
				}
			} else {
				addWarning(r, "Benchmark price unavailable for part of the selected period.")
			}
		}
		r.History = append(r.History, point)
	}
	if r.Summary.Cash < 0 {
		addWarning(r, "Cash is negative. Review deposits, fees and opening balances.")
	}
	if len(r.Holdings) > 0 && r.Holdings[0].Weight > 25 {
		addWarning(r, fmt.Sprintf("%s represents %.1f%% of the portfolio.", r.Holdings[0].Symbol, r.Holdings[0].Weight))
	}
	if !r.Complete {
		addWarning(r, "Totals are incomplete until missing prices and exchange rates are available.")
	}
	for _, h := range r.Holdings {
		if h.AssetClass == "" || h.Sector == "" || h.Country == "" {
			addWarning(r, "Some holdings are unclassified. Edit their classification to complete allocation analysis.")
			break
		}
	}
	return r, nil
}

// A single database snapshot prevents a transfer from appearing on only one side.
func (s *PortfolioService) loadLedgers(userID uuid.UUID, accountID *uuid.UUID) (string, []accountLedger, error) {
	currency := "USD"
	ledgers := []accountLedger{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var user usermodel.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}
		if user.BaseCurrency != "" {
			currency = user.BaseCurrency
		}
		accounts := []accountmodel.Account{}
		query := tx.Where("user_id = ?", userID)
		if accountID != nil {
			query = query.Where("id = ?", *accountID)
		} else {
			query = query.Where("is_excluded = false")
		}
		if err := query.Order("id").Find(&accounts).Error; err != nil {
			return err
		}
		if accountID != nil {
			if len(accounts) != 1 {
				return errors.New("account not found")
			}
			currency = accounts[0].Currency
		}
		for _, account := range accounts {
			events, err := activityquery.GetLedgerEvents(tx, account.ID)
			if err != nil {
				return err
			}
			holdings := []holdingmodel.Holding{}
			if err = tx.Where("account_id = ?", account.ID).Find(&holdings).Error; err != nil {
				return err
			}
			metadata := map[string]holdingmodel.Holding{}
			for _, holding := range holdings {
				metadata[holding.Symbol] = holding
			}
			ledgers = append(ledgers, accountLedger{account, events, metadata})
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return currency, ledgers, err
}
