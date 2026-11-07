package portfolio

import (
	"github.com/google/uuid"
	model "keepnode/app/models/wealth"
	"keepnode/pkg/portfolio"
	"math"
	"time"
)

type ReportOptions struct {
	AccountIDs []uuid.UUID
	TagID      *uuid.UUID
	Method     string
}
type Performance struct {
	Method           string   `json:"method"`
	PeriodReturn     *float64 `json:"period_return"`
	AnnualizedReturn *float64 `json:"annualized_return"`
	Profit           *float64 `json:"profit"`
	PriceEffect      *float64 `json:"price_effect"`
	CurrencyEffect   *float64 `json:"currency_effect"`
}

func validMethod(s string) bool {
	return s == "" || s == "dietz" || s == "roai" || s == "mwr" || s == "twr"
}
func finitePtr(v float64) *float64 {
	if !portfolio.ValidNumber(v) {
		return nil
	}
	return &v
}

// XIRR uses dated external flows. Ambiguous (multiple-root) results remain unknown.
func xirr(flows []float64, days []float64) *float64 {
	if len(flows) < 2 || len(days) != len(flows) {
		return nil
	}
	positive, negative := false, false
	for _, v := range flows {
		positive = positive || v > 0
		negative = negative || v < 0
	}
	if !positive || !negative {
		return nil
	}
	npv := func(x float64) float64 {
		v := 0.
		for i, f := range flows {
			v += f * math.Exp(-x*days[i]/365)
		}
		return v
	}
	// Standard accumulation has exactly one sign change. Bracket it directly;
	// sparse external flows keep daily chart evaluation inexpensive.
	signChanges, lastSign := 0, 0
	for _, f := range flows {
		sign := 0
		if f > 0 {
			sign = 1
		} else if f < 0 {
			sign = -1
		}
		if sign != 0 {
			if lastSign != 0 && sign != lastSign {
				signChanges++
			}
			lastSign = sign
		}
	}
	steps := 128
	if signChanges == 1 {
		steps = 1
	}
	roots := []float64{}
	left, fl := -13.8, npv(-13.8)
	for i := 1; i <= steps; i++ {
		right := -13.8 + 27.6*float64(i)/float64(steps)
		fr := npv(right)
		if fl == 0 {
			roots = append(roots, left)
		} else if math.Signbit(fl) != math.Signbit(fr) {
			lo, hi, flo := left, right, fl
			for n := 0; n < 64; n++ {
				mid := (lo + hi) / 2
				fm := npv(mid)
				if fm == 0 {
					lo, hi = mid, mid
					break
				}
				if math.Signbit(flo) != math.Signbit(fm) {
					hi = mid
				} else {
					lo, flo = mid, fm
				}
			}
			roots = append(roots, (lo+hi)/2)
		}
		left, fl = right, fr
	}
	// Nonconventional flows can have closely spaced/tangent roots that a scan
	// cannot certify as unique. Return unknown instead of choosing a root.
	if signChanges != 1 || len(roots) != 1 {
		return nil
	}
	return finitePtr(math.Expm1(roots[0]))
}
func applyPerformance(r *PortfolioReport, method string) {
	if method == "" {
		method = "dietz"
	}
	r.Performance = Performance{Method: method}
	r.Method = map[string]string{"dietz": "daily-linked Modified Dietz (mid-day external cash flows)", "twr": "daily-linked TWR (external flows at end of day)", "roai": "ROAI (daily average net invested capital, mid-day external flows)", "mwr": "MWR / XIRR (dated external cash flows, actual/365)"}[method]
	if len(r.History) == 0 {
		return
	}
	first := r.History[0]
	if first.Value == nil {
		for i := range r.History {
			r.History[i].Return = nil
		}
		return
	}
	start, _ := time.Parse("2006-01-02", first.Date)
	netFlows, averageCapital, capital, linked := 0., 0., *first.Value, 1.
	complete := true
	fxComplete := first.ValueWithoutFX != nil
	frozenFlows := 0.
	flows := []float64{-*first.Value}
	days := []float64{0}
	for i := range r.History {
		point := &r.History[i]
		date, _ := time.Parse("2006-01-02", point.Date)
		elapsed := date.Sub(start).Hours() / 24
		if point.Value == nil {
			complete = false
		}
		if point.ValueWithoutFX == nil {
			fxComplete = false
		}
		if i == 0 {
			if method != "dietz" {
				point.Return = finitePtr(0)
			}
			continue
		}
		netFlows += point.NetFlow
		frozenFlows += point.FlowWithoutFX
		previous := r.History[i-1].Value
		averageCapital += capital + point.NetFlow/2
		capital += point.NetFlow
		if point.NetFlow != 0 {
			flows = append(flows, -point.NetFlow)
			days = append(days, elapsed)
		}
		if !complete {
			point.Return = nil
			continue
		}
		gain := *point.Value - *first.Value - netFlows
		switch method {
		case "twr":
			if previous == nil || *previous <= 0 {
				if previous != nil && *previous == 0 && *point.Value == point.NetFlow {
					point.Return = finitePtr((linked - 1) * 100)
				} else {
					complete = false
					point.Return = nil
				}
			} else {
				linked *= (*point.Value - point.NetFlow) / *previous
				point.Return = finitePtr((linked - 1) * 100)
			}
		case "roai":
			if averageCapital > 0 {
				point.Return = finitePtr(gain / (averageCapital / float64(i)) * 100)
			} else {
				point.Return = nil
			}
		case "mwr":
			f := append(append([]float64{}, flows...), *point.Value)
			d := append(append([]float64{}, days...), elapsed)
			rate := xirr(f, d)
			point.Return = nil
			if rate != nil && elapsed > 0 {
				point.Return = finitePtr((math.Pow(1+*rate, elapsed/365) - 1) * 100)
			}
		}
	}
	last := r.History[len(r.History)-1]
	end, _ := time.Parse("2006-01-02", last.Date)
	elapsed := end.Sub(start).Hours() / 24
	if complete && last.Value != nil {
		r.Performance.PeriodReturn = last.Return
		r.Performance.Profit = finitePtr(*last.Value - *first.Value - netFlows)
		if last.Return != nil && elapsed > 0 && *last.Return > -100 {
			r.Performance.AnnualizedReturn = finitePtr((math.Pow(1+*last.Return/100, 365/elapsed) - 1) * 100)
		}
		if fxComplete && last.ValueWithoutFX != nil {
			r.Performance.PriceEffect = finitePtr(*last.ValueWithoutFX - *first.ValueWithoutFX - frozenFlows)
			if r.Performance.PriceEffect != nil && r.Performance.Profit != nil {
				r.Performance.CurrencyEffect = finitePtr(*r.Performance.Profit - *r.Performance.PriceEffect)
			}
		}
	}
}

func (s *PortfolioService) netWorth(user uuid.UUID, r *PortfolioReport, book interface {
	Rate(string, string, time.Time) (float64, error)
}, include bool) error {
	if !include {
		return nil
	}
	debts := []model.Liability{}
	if err := s.DB.Where("user_id = ?", user).Find(&debts).Error; err != nil {
		return err
	}
	balances := map[uuid.UUID][]model.LiabilityBalance{}
	for _, debt := range debts {
		rows := []model.LiabilityBalance{}
		if err := s.DB.Where("liability_id = ?", debt.ID).Order("date").Find(&rows).Error; err != nil {
			return err
		}
		balances[debt.ID] = rows
	}
	for i := range r.History {
		point := &r.History[i]
		total := 0.
		valid := true
		day, _ := time.Parse("2006-01-02", point.Date)
		for _, debt := range debts {
			balance := 0.
			for _, v := range balances[debt.ID] {
				if v.Date > point.Date {
					break
				}
				balance = v.Amount
			}
			if balance == 0 {
				continue
			}
			fx, err := book.Rate(debt.Currency, r.Currency, day)
			if err != nil {
				valid = false
				addWarning(r, "A liability exchange rate is unavailable; net worth is incomplete.")
				continue
			}
			total += balance * fx
		}
		if valid && point.Value != nil {
			point.NetWorth = finitePtr(*point.Value - total)
		}
		if i == len(r.History)-1 && valid {
			r.Summary.Liabilities = finitePtr(total)
			if r.Complete {
				r.Summary.NetWorth = finitePtr(r.Summary.Value - total)
			}
		}
	}
	return nil
}

func (s *PortfolioService) NetWorthReport(user uuid.UUID, period string) (*PortfolioReport, error) {
	return s.Report(user, nil, period, "")
}
