package wealth

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	model "keepnode/app/models/wealth"
	portfolioservice "keepnode/app/services/portfolio"
	"keepnode/pkg/portfolio"
	"math"
)

func DefaultPreferences(user uuid.UUID) model.Preferences {
	return model.Preferences{UserID: user, MaxAssetWeight: 25, MaxAccountWeight: 70, MaxCurrencyWeight: 80, MaxFeeRatio: 1, EmergencyMonths: 6, AnnualReturn: 5, WithdrawalRate: 4, HorizonYears: 30, DisabledRules: []string{}}
}
func (s *Service) Preferences(user uuid.UUID) (model.Preferences, error) {
	p := DefaultPreferences(user)
	err := s.DB.First(&p, "user_id = ?", user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultPreferences(user), nil
	}
	if p.DisabledRules == nil {
		p.DisabledRules = []string{}
	}
	return p, err
}
func (s *Service) SavePreferences(user uuid.UUID, p model.Preferences) (model.Preferences, error) {
	for _, v := range []float64{p.MaxAssetWeight, p.MaxAccountWeight, p.MaxCurrencyWeight} {
		if !portfolio.ValidNumber(v) || v <= 0 || v > 100 {
			return p, errors.New("weight thresholds must be between 0 and 100")
		}
	}
	for _, v := range []float64{p.EmergencyFund, p.MonthlyExpenses, p.MonthlyContribution, p.AnnualExpenses} {
		if !validAmount(v) {
			return p, errors.New("planning amounts must be finite and nonnegative")
		}
	}
	if !portfolio.ValidNumber(p.MaxFeeRatio) || p.MaxFeeRatio < 0 || p.MaxFeeRatio > 100 || !portfolio.ValidNumber(p.EmergencyMonths) || p.EmergencyMonths < 0 || p.EmergencyMonths > 120 || !portfolio.ValidNumber(p.AnnualReturn) || p.AnnualReturn <= -100 || p.AnnualReturn > 100 || !portfolio.ValidNumber(p.WithdrawalRate) || p.WithdrawalRate <= 0 || p.WithdrawalRate > 100 || p.HorizonYears < 1 || p.HorizonYears > 80 {
		return p, errors.New("invalid planning assumptions")
	}
	valid := map[string]bool{"asset": true, "account": true, "currency": true, "fees": true, "emergency": true, "liquidity": true}
	for _, id := range p.DisabledRules {
		if !valid[id] {
			return p, errors.New("invalid risk rule")
		}
	}
	p.UserID = user
	err := s.DB.Save(&p).Error
	return p, err
}

type RiskRule struct {
	ID      string   `json:"id"`
	Status  string   `json:"status"`
	Value   *float64 `json:"value"`
	Limit   float64  `json:"limit"`
	Unit    string   `json:"unit"`
	Subject string   `json:"subject,omitempty"`
	Enabled bool     `json:"enabled"`
}
type RiskReport struct {
	Rules    []RiskRule `json:"rules"`
	Currency string     `json:"currency"`
	Complete bool       `json:"complete"`
	Warnings []string   `json:"warnings"`
}

func RiskChecks(r *portfolioservice.PortfolioReport, p model.Preferences) RiskReport {
	out := RiskReport{Rules: []RiskRule{}, Currency: r.Currency, Complete: r.Complete, Warnings: r.Warnings}
	disabled := map[string]bool{}
	for _, id := range p.DisabledRules {
		disabled[id] = true
	}
	add := func(id, unit, subject string, value *float64, limit float64, minimum bool) {
		status := "unknown"
		if value != nil && r.Complete {
			status = "ok"
			if !minimum && *value > limit || minimum && *value < limit {
				status = "warning"
			}
		} else {
			value = nil
		}
		out.Rules = append(out.Rules, RiskRule{id, status, value, limit, unit, subject, !disabled[id]})
	}
	var asset *float64
	subject := ""
	for _, h := range r.Holdings {
		if asset == nil || h.Weight > *asset {
			v := h.Weight
			asset = &v
			subject = h.Symbol
		}
	}
	if len(r.Holdings) == 0 && r.Complete {
		v := 0.
		asset = &v
	}
	add("asset", "percent", subject, asset, p.MaxAssetWeight, false)
	for _, entry := range []struct {
		id    string
		limit float64
	}{{"account", p.MaxAccountWeight}, {"currency", p.MaxCurrencyWeight}} {
		var highest *float64
		name := ""
		for _, a := range r.Allocations[entry.id] {
			if highest == nil || a.Weight > *highest {
				v := a.Weight
				highest = &v
				name = a.Name
			}
		}
		add(entry.id, "percent", name, highest, entry.limit, false)
	}
	var fees *float64
	if r.Summary.NetContributions > 0 {
		v := r.Summary.Fees / r.Summary.NetContributions * 100
		fees = &v
	}
	add("fees", "percent", "", fees, p.MaxFeeRatio, false)
	var months *float64
	if p.MonthlyExpenses > 0 {
		v := math.Max(0, math.Min(p.EmergencyFund, r.Summary.Cash)) / p.MonthlyExpenses
		months = &v
	}
	add("emergency", "months", "", months, p.EmergencyMonths, true)
	cash := r.Summary.Cash
	add("liquidity", "money", "", &cash, 0, true)
	return out
}

type FirePoint struct {
	Year          int     `json:"year"`
	Value         float64 `json:"value"`
	Contributions float64 `json:"contributions"`
}
type FireReport struct {
	Currency      string      `json:"currency"`
	Complete      bool        `json:"complete"`
	Investable    *float64    `json:"investable"`
	Target        *float64    `json:"target"`
	YearsToTarget *float64    `json:"years_to_target"`
	Projection    []FirePoint `json:"projection"`
	Warnings      []string    `json:"warnings"`
}

func FireProjection(r *portfolioservice.PortfolioReport, p model.Preferences) FireReport {
	out := FireReport{Currency: r.Currency, Complete: r.Complete && r.Summary.NetWorth != nil, Projection: []FirePoint{}, Warnings: r.Warnings}
	if !out.Complete {
		return out
	}
	initial := math.Max(0, *r.Summary.NetWorth-p.EmergencyFund)
	out.Investable = &initial
	if p.AnnualExpenses > 0 && p.WithdrawalRate > 0 {
		target := p.AnnualExpenses / (p.WithdrawalRate / 100)
		out.Target = &target
	}
	value, contributions := initial, 0.
	rate := math.Pow(1+p.AnnualReturn/100, 1./12) - 1
	out.Projection = append(out.Projection, FirePoint{0, value, 0})
	if out.Target != nil && value >= *out.Target {
		v := 0.
		out.YearsToTarget = &v
	}
	for month := 1; month <= p.HorizonYears*12; month++ {
		value = value*(1+rate) + p.MonthlyContribution
		contributions += p.MonthlyContribution
		if !portfolio.ValidNumber(value) {
			out.Complete = false
			out.Projection = []FirePoint{}
			out.YearsToTarget = nil
			out.Warnings = append(out.Warnings, "Projection exceeds the supported amount range.")
			return out
		}
		if out.Target != nil && out.YearsToTarget == nil && value >= *out.Target {
			v := float64(month) / 12
			out.YearsToTarget = &v
		}
		if month%12 == 0 {
			out.Projection = append(out.Projection, FirePoint{month / 12, value, contributions})
		}
	}
	return out
}
