package services

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"keepnode/app/models"
	"keepnode/app/queries"
	"keepnode/pkg/portfolio"
	"time"
)

type HoldingService struct {
	DB              *gorm.DB
	marketService   *MarketService
	activityQueries *queries.ActivityQueries
}

func NewHoldingService(db *gorm.DB, market *MarketService) *HoldingService {
	return &HoldingService{db, market, &queries.ActivityQueries{DB: db}}
}
func (s *HoldingService) AddHolding(accountID uuid.UUID, req *models.CreateHoldingRequest) (*models.Holding, error) {
	date := req.Date
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	_, err := NewActivityService(s.DB, s.marketService).CreateActivity(accountID, nil, &models.CreateActivityRequest{Type: "buy", Date: date, Symbol: req.Symbol, Quantity: req.Quantity, UnitPrice: req.BuyPrice, Fee: req.Fee, Currency: req.BuyCurrency, ExchangeRate: req.ExchangeRate, Notes: req.Notes})
	if err != nil {
		return nil, err
	}
	h, err := (&queries.HoldingQueries{DB: s.DB}).GetHoldingBySymbol(accountID, normalizeSymbol(req.Symbol))
	return &h, err
}
func (s *HoldingService) GetHoldingByID(id, accountID uuid.UUID) (*models.Holding, error) {
	h, err := (&queries.HoldingQueries{DB: s.DB}).GetHoldingByID(id, accountID)
	return &h, err
}
func (s *HoldingService) UpdateHolding(id, accountID uuid.UUID, req *models.UpdateHoldingRequest) (*models.Holding, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := lockAccount(tx, accountID)
		if err != nil {
			return err
		}
		var h models.Holding
		if err = tx.First(&h, "id = ? AND account_id = ?", id, accountID).Error; err != nil {
			return err
		}
		values := map[string]any{}
		if req.Notes != nil {
			values["notes"] = *req.Notes
		}
		if req.AssetClass != nil {
			values["asset_class"] = *req.AssetClass
		}
		if req.Sector != nil {
			values["sector"] = *req.Sector
		}
		if req.Country != nil {
			values["country"] = *req.Country
		}
		for _, key := range []string{"asset_class", "sector", "country"} {
			if v, ok := values[key]; ok && (len(v.(string)) > 100 || (key == "asset_class" && len(v.(string)) > 40)) {
				return errors.New("classification is too long")
			}
		}
		changed := (req.Quantity != nil && *req.Quantity != h.Quantity) || (req.BuyPrice != nil && *req.BuyPrice != h.BuyPrice) || (req.BuyCurrency != nil && *req.BuyCurrency != h.BuyCurrency)
		if changed {
			var events []models.Activity
			if err = tx.Where("account_id = ? AND symbol = ? AND ledger = true AND type IN ?", accountID, h.Symbol, []string{"buy", "sell", "opening"}).Find(&events).Error; err != nil {
				return err
			}
			if len(events) != 1 || (events[0].Type != "buy" && events[0].Type != "opening") {
				return errors.New("edit the individual activities to change a position with multiple transactions")
			}
			old := events[0]
			quantity, price, currency := old.Quantity, old.UnitPrice, old.Currency
			if req.Quantity != nil {
				quantity = *req.Quantity
			}
			if req.BuyPrice != nil {
				price = *req.BuyPrice
			}
			if req.BuyCurrency != nil {
				currency = *req.BuyCurrency
			}
			rate := old.ExchangeRate
			if currency != old.Currency {
				rate = 0
			}
			svc := NewActivityService(tx, s.marketService)
			updated, e := svc.prepare(a, &models.CreateActivityRequest{Type: old.Type, Date: old.Date.Format("2006-01-02"), Symbol: old.Symbol, Quantity: quantity, UnitPrice: price, Fee: old.Fee, Currency: currency, ExchangeRate: rate, Notes: old.Notes}, old.Type == "opening")
			if e != nil {
				return e
			}
			if err = tx.Model(&old).Select("quantity", "unit_price", "currency", "exchange_rate").Updates(&updated).Error; err != nil {
				return err
			}
			if err = svc.rebuild(tx, a); err != nil {
				return err
			}
		}
		if len(values) > 0 {
			return tx.Model(&h).Updates(values).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetHoldingByID(id, accountID)
}
func (s *HoldingService) DeleteHolding(id, accountID uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := lockAccount(tx, accountID)
		if err != nil {
			return err
		}
		var h models.Holding
		if err = tx.First(&h, "id = ? AND account_id = ?", id, accountID).Error; err != nil {
			return err
		}
		var events []models.Activity
		if err = tx.Where("account_id = ? AND symbol = ? AND ledger = true AND type IN ?", accountID, h.Symbol, []string{"buy", "sell", "opening"}).Find(&events).Error; err != nil {
			return err
		}
		if len(events) > 1 {
			return errors.New("this position has transaction history; record a sale or edit its activities")
		}
		if len(events) == 1 {
			if err = tx.Delete(&events[0]).Error; err != nil {
				return err
			}
		}
		return NewActivityService(tx, s.marketService).rebuild(tx, a)
	})
}
func (s *HoldingService) GetHoldingsWithLiveData(accountID uuid.UUID) ([]models.HoldingWithLiveData, *models.HoldingsSummary, error) {
	var a models.Account
	if err := s.DB.First(&a, "id = ?", accountID).Error; err != nil {
		return nil, nil, err
	}
	events, err := ledgerEvents(s.DB, accountID)
	if err != nil {
		return nil, nil, err
	}
	if err = resolveOpeningRates(s.DB, s.marketService, a, events); err != nil {
		return nil, nil, err
	}
	state, err := portfolio.Replay(a, events, time.Now().UTC())
	if err != nil {
		return nil, nil, err
	}
	holdings := []models.Holding{}
	if err = s.DB.Where("account_id = ? AND quantity > 0", accountID).Order("symbol").Find(&holdings).Error; err != nil {
		return nil, nil, err
	}
	result := []models.HoldingWithLiveData{}
	summary := &models.HoldingsSummary{CashBalance: state.Cash, RealizedProfitLoss: state.Realized, Dividend: state.Dividends, Interest: state.Interest, Fees: state.Fees, Complete: true, Warnings: []string{}}
	now := time.Now().UTC()
	book := newPriceBook(s.DB, s.marketService, now.AddDate(0, 0, -7), now)
	for _, h := range holdings {
		p := state.Positions[h.Symbol]
		if p == nil {
			continue
		}
		row := models.HoldingWithLiveData{ID: h.ID, AccountID: h.AccountID, Symbol: h.Symbol, Name: h.Name, Quantity: h.Quantity, BuyPrice: h.BuyPrice, BuyCurrency: h.BuyCurrency, FirstActivity: p.FirstActivity, AssetClass: h.AssetClass, Sector: h.Sector, Country: h.Country}
		row.EntryPrice = h.BuyPrice
		count := 0
		for _, event := range events {
			if event.Symbol == h.Symbol && (event.Type == "buy" || event.Type == "sell" || event.Type == "opening") {
				count++
				row.EntryPrice = event.UnitPrice
			}
		}
		row.CanEditPosition = count == 1
		price, date, e := book.price(h.Symbol, now)
		rate, fe := book.rate(h.BuyCurrency, a.Currency, now)
		if e != nil || fe != nil {
			summary.Complete = false
			if e != nil {
				summary.Warnings = append(summary.Warnings, e.Error())
			}
			if fe != nil {
				summary.Warnings = append(summary.Warnings, fe.Error())
			}
		} else {
			row.PriceAvailable = true
			row.PriceDate = date
			row.CurrentPrice = price
			row.Value = h.Quantity * price
			row.ProfitLoss = row.Value - p.Cost
			if p.Cost > 0 {
				row.Performance = row.ProfitLoss / p.Cost * 100
			}
			summary.TotalValue += row.Value * rate
			summary.TotalProfitLoss += row.Value*rate - p.AccountCost
			prior, _, pe := book.price(h.Symbol, now.AddDate(0, 0, -1))
			if pe == nil {
				row.Change = price - prior
				row.ChangePercent = row.Change / prior * 100
				summary.DailyChange += row.Change * h.Quantity * rate
			}
		}
		summary.TotalCost += p.AccountCost
		result = append(result, row)
	}
	if summary.TotalCost > 0 {
		summary.TotalPerformance = summary.TotalProfitLoss / summary.TotalCost * 100
	}
	return result, summary, nil
}
