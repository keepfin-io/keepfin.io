package services

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"keepnode/app/models"
	"keepnode/app/queries"
	"keepnode/pkg/portfolio"
	"math"
	"regexp"
	"strings"
	"time"
)

type ActivityService struct {
	DB            *gorm.DB
	marketService *MarketService
}

func NewActivityService(db *gorm.DB, market *MarketService) *ActivityService {
	return &ActivityService{DB: db, marketService: market}
}

var symbolPattern = regexp.MustCompile(`^[A-Z0-9.^=\-]{1,20}$`)

func validCurrency(s string) bool {
	for _, c := range models.SupportedCurrencies {
		if c == s {
			return true
		}
	}
	return false
}

func (s *ActivityService) prepare(account models.Account, req *models.CreateActivityRequest, opening bool) (models.Activity, error) {
	a := models.Activity{AccountID: account.ID, Ledger: true, Type: req.Type, Symbol: strings.ToUpper(strings.TrimSpace(req.Symbol)), Quantity: req.Quantity, UnitPrice: req.UnitPrice, Fee: req.Fee, Currency: strings.ToUpper(req.Currency), Notes: strings.TrimSpace(req.Notes), ExchangeRate: req.ExchangeRate}
	valid := false
	for _, v := range models.ValidActivityTypes() {
		if a.Type == v {
			valid = true
		}
	}
	if opening && a.Type == "opening" {
		valid = true
	}
	if !valid {
		return a, errors.New("invalid activity type")
	}
	d, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		d, err = time.Parse(time.RFC3339, req.Date)
	}
	if err != nil {
		return a, errors.New("invalid activity date")
	}
	a.Date = portfolio.Day(d)
	if a.Date.After(portfolio.Day(time.Now())) {
		return a, errors.New("future activities are not supported")
	}
	if account.OpeningDate != nil && a.Date.Before(portfolio.Day(*account.OpeningDate)) {
		return a, errors.New("activity precedes the account opening date")
	}
	if a.Currency == "" {
		a.Currency = account.Currency
	}
	if !validCurrency(a.Currency) {
		return a, errors.New("unsupported currency")
	}
	if !portfolio.ValidNumber(a.Quantity) || !portfolio.ValidNumber(a.UnitPrice) || !portfolio.ValidNumber(a.Fee) || !portfolio.ValidNumber(a.ExchangeRate) || a.Quantity <= 0 || (a.UnitPrice <= 0 && !(opening && a.UnitPrice == 0)) || a.Fee < 0 || a.ExchangeRate < 0 {
		return a, errors.New("amounts must be finite and positive; fees cannot be negative")
	}
	for _, field := range []struct {
		name   string
		value  float64
		places int
	}{{"quantity", a.Quantity, 8}, {"price", a.UnitPrice, 4}, {"fee", a.Fee, 4}, {"exchange rate", a.ExchangeRate, 10}} {
		rounded := math.Round(field.value*math.Pow10(field.places)) / math.Pow10(field.places)
		if math.Abs(field.value-rounded) > math.Pow10(-field.places)*0.0001 {
			return a, fmt.Errorf("%s supports at most %d decimal places", field.name, field.places)
		}
	}
	if a.Quantity >= 1e10 || len(a.Symbol) > 20 || len(a.Notes) > 2000 {
		return a, errors.New("quantity, symbol or notes exceed the supported limit")
	}
	if a.Type == "buy" || a.Type == "sell" || a.Type == "opening" {
		if !symbolPattern.MatchString(a.Symbol) {
			return a, errors.New("a valid symbol is required")
		}
	} else if a.Symbol == "" {
		a.Symbol = a.Currency
	}
	if a.Currency == account.Currency {
		a.ExchangeRate = 1
	} else if a.ExchangeRate == 0 {
		a.ExchangeRate, err = newPriceBook(s.DB, s.marketService, a.Date, a.Date).rate(a.Currency, account.Currency, a.Date)
		if err != nil {
			return a, err
		}
	}
	if a.ExchangeRate <= 0 {
		return a, errors.New("exchange rate must be positive")
	}
	if !portfolio.ValidNumber(portfolio.Mul(a.Quantity, a.UnitPrice) * a.ExchangeRate) {
		return a, errors.New("transaction total exceeds the supported limit")
	}
	a.Total = portfolio.Mul(a.Quantity, a.UnitPrice) + a.Fee
	return a, nil
}
func lockAccount(tx *gorm.DB, id uuid.UUID) (models.Account, error) {
	var a models.Account
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ?", id).Error
	return a, err
}
func ledgerEvents(db *gorm.DB, id uuid.UUID) ([]models.Activity, error) {
	a := []models.Activity{}
	err := db.Where("account_id = ? AND ledger = true", id).Order("date, created_at, id").Find(&a).Error
	return a, err
}
func (s *ActivityService) rebuild(tx *gorm.DB, account models.Account) error {
	events, err := ledgerEvents(tx, account.ID)
	if err != nil {
		return err
	}
	if err = resolveOpeningRates(tx, s.marketService, account, events); err != nil {
		return err
	}
	state, err := portfolio.Replay(account, events, time.Now().UTC())
	if err != nil {
		return err
	}
	var existing []models.Holding
	if err = tx.Where("account_id = ?", account.ID).Find(&existing).Error; err != nil {
		return err
	}
	bySymbol := map[string]models.Holding{}
	for _, h := range existing {
		bySymbol[h.Symbol] = h
	}
	for symbol, p := range state.Positions {
		h, exists := bySymbol[symbol]
		if !exists {
			h = models.Holding{AccountID: account.ID, Symbol: symbol, Name: p.Name, BuyCurrency: p.Currency}
		}
		h.Quantity = p.Quantity
		h.BuyPrice = portfolio.Div(p.Cost, p.Quantity)
		h.BuyCurrency = p.Currency
		if exists {
			err = tx.Model(&h).Updates(map[string]any{"quantity": h.Quantity, "buy_price": h.BuyPrice, "buy_currency": h.BuyCurrency}).Error
		} else {
			err = tx.Create(&h).Error
		}
		if err != nil {
			return err
		}
		if err = tx.Model(&models.Activity{}).Where("account_id = ? AND symbol = ? AND ledger = true AND type IN ?", account.ID, symbol, []string{"buy", "sell", "opening"}).Update("holding_id", h.ID).Error; err != nil {
			return err
		}
		delete(bySymbol, symbol)
	}
	for _, h := range bySymbol {
		if err = tx.Model(&h).Update("quantity", 0).Error; err != nil {
			return err
		}
	}
	return nil
}
func (s *ActivityService) CreateActivity(accountID uuid.UUID, _ *uuid.UUID, req *models.CreateActivityRequest) (*models.Activity, error) {
	var result models.Activity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		account, err := lockAccount(tx, accountID)
		if err != nil {
			return err
		}
		if !account.IsActive {
			return errors.New("account is inactive")
		}
		result, err = s.prepare(account, req, false)
		if err != nil {
			return err
		}
		if err = tx.Create(&result).Error; err != nil {
			return err
		}
		return s.rebuild(tx, account)
	})
	return &result, err
}
func (s *ActivityService) UpdateActivity(id, accountID uuid.UUID, req *models.CreateActivityRequest) (*models.Activity, error) {
	var result models.Activity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		account, err := lockAccount(tx, accountID)
		if err != nil {
			return err
		}
		var old models.Activity
		if err = tx.First(&old, "id = ? AND account_id = ?", id, accountID).Error; err != nil {
			return err
		}
		if !old.Ledger {
			return errors.New("archived history is read-only")
		}
		if old.TransferID != nil {
			return errors.New("delete the paired transfer and record it again")
		}
		if old.Type == "opening" && req.Type != "opening" {
			return errors.New("opening positions must remain opening positions")
		}
		result, err = s.prepare(account, req, old.Type == "opening")
		if err != nil {
			return err
		}
		result.ID = old.ID
		result.Base = old.Base
		result.ImportKey = old.ImportKey
		if err = tx.Model(&old).Select("symbol", "type", "date", "quantity", "unit_price", "fee", "currency", "exchange_rate", "notes").Updates(&result).Error; err != nil {
			return err
		}
		return s.rebuild(tx, account)
	})
	return &result, err
}
func (s *ActivityService) DeleteActivity(id, accountID uuid.UUID) error {
	var old models.Activity
	if err := s.DB.First(&old, "id = ? AND account_id = ?", id, accountID).Error; err != nil {
		return err
	}
	if !old.Ledger {
		return errors.New("archived history is read-only")
	}
	if old.TransferID != nil {
		return s.deleteTransfer(*old.TransferID, accountID)
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := lockAccount(tx, accountID)
		if err != nil {
			return err
		}
		if err = tx.Where("id = ? AND account_id = ?", id, accountID).Delete(&models.Activity{}).Error; err != nil {
			return err
		}
		return s.rebuild(tx, a)
	})
}
func (s *ActivityService) GetActivities(id uuid.UUID) ([]models.Activity, error) {
	return (&queries.ActivityQueries{DB: s.DB}).GetActivitiesByAccountID(id)
}
func (s *ActivityService) GetActivitiesByHolding(id uuid.UUID) ([]models.Activity, error) {
	return (&queries.ActivityQueries{DB: s.DB}).GetActivitiesByHoldingID(id)
}
func (s *ActivityService) GetActivityCount(id uuid.UUID) (int64, error) {
	return (&queries.ActivityQueries{DB: s.DB}).CountActivitiesByAccountID(id)
}
func (s *ActivityService) CreateActivityFromHolding(accountID uuid.UUID, h *models.Holding, fee float64) (*models.Activity, error) {
	return s.CreateActivity(accountID, nil, &models.CreateActivityRequest{Symbol: h.Symbol, Type: "buy", Date: time.Now().UTC().Format("2006-01-02"), Quantity: h.Quantity, UnitPrice: h.BuyPrice, Fee: fee, Currency: h.BuyCurrency})
}

type TransferRequest struct {
	FromAccountID  uuid.UUID `json:"from_account_id"`
	ToAccountID    uuid.UUID `json:"to_account_id"`
	Date           string    `json:"date"`
	Amount         float64   `json:"amount"`
	ReceivedAmount float64   `json:"received_amount"`
	Fee            float64   `json:"fee"`
	Notes          string    `json:"notes"`
}

func (s *ActivityService) Transfer(userID uuid.UUID, req TransferRequest) error {
	if req.FromAccountID == req.ToAccountID {
		return errors.New("choose two different accounts")
	}
	if !portfolio.ValidNumber(req.Amount) || !portfolio.ValidNumber(req.ReceivedAmount) || !portfolio.ValidNumber(req.Fee) || req.Amount <= 0 || req.ReceivedAmount <= 0 || req.Fee < 0 {
		return errors.New("invalid transfer amounts")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var accounts []models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ? AND user_id = ? AND is_active = true", []uuid.UUID{req.FromAccountID, req.ToAccountID}, userID).Order("id").Find(&accounts).Error; err != nil {
			return err
		}
		if len(accounts) != 2 {
			return errors.New("account not found")
		}
		var from, to models.Account
		for _, a := range accounts {
			if a.ID == req.FromAccountID {
				from = a
			} else {
				to = a
			}
		}
		if from.Currency == to.Currency && req.Amount != req.ReceivedAmount {
			return errors.New("same-currency transfer amounts must match; enter fees separately")
		}
		transferID := uuid.New()
		for _, item := range []struct {
			account     models.Account
			amount, fee float64
			kind        string
		}{{from, req.Amount, req.Fee, "transfer_out"}, {to, req.ReceivedAmount, 0, "transfer_in"}} {
			a, err := s.prepare(item.account, &models.CreateActivityRequest{Type: "deposit", Date: req.Date, Quantity: 1, UnitPrice: item.amount, Fee: item.fee, Currency: item.account.Currency, Notes: req.Notes}, false)
			if err != nil {
				return err
			}
			a.Type = item.kind
			a.TransferID = &transferID
			if err = tx.Create(&a).Error; err != nil {
				return err
			}
			if err = s.rebuild(tx, item.account); err != nil {
				return err
			}
		}
		return nil
	})
}
func (s *ActivityService) deleteTransfer(transferID, accountID uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var source models.Account
		if err := tx.First(&source, "id = ?", accountID).Error; err != nil {
			return err
		}
		var pair []models.Activity
		if err := tx.Where("transfer_id = ?", transferID).Find(&pair).Error; err != nil {
			return err
		}
		if len(pair) != 2 {
			return errors.New("transfer pair is incomplete")
		}
		var accounts []models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ? AND user_id = ?", []uuid.UUID{pair[0].AccountID, pair[1].AccountID}, source.UserID).Order("id").Find(&accounts).Error; err != nil {
			return err
		}
		if len(accounts) != 2 {
			return errors.New("both transfer accounts must exist")
		}
		if err := tx.Where("transfer_id = ?", transferID).Delete(&models.Activity{}).Error; err != nil {
			return err
		}
		for _, a := range accounts {
			if err := s.rebuild(tx, a); err != nil {
				return fmt.Errorf("transfer removal: %w", err)
			}
		}
		return nil
	})
}
