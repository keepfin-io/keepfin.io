package wealth

import (
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	currencymodel "keepnode/app/models/currency"
	holdingmodel "keepnode/app/models/holding"
	model "keepnode/app/models/wealth"
	activityservice "keepnode/app/services/activity"
	marketservice "keepnode/app/services/market"
	portfolioservice "keepnode/app/services/portfolio"
	"keepnode/pkg/portfolio"
	"keepnode/pkg/worker"
)

var ErrNotFound = errors.New("record not found")
var symbolPattern = regexp.MustCompile(`^[A-Z0-9.^=\-]{1,20}$`)

type Service struct {
	DB               *gorm.DB
	Market           *marketservice.MarketService
	Activities       *activityservice.ActivityService
	Portfolio        *portfolioservice.PortfolioService
	marketWake       worker.Signal
	providerMu       sync.Mutex
	providerLoadedAt time.Time
	defaultProvider  string
}

func New(db *gorm.DB, market *marketservice.MarketService, activities *activityservice.ActivityService, report *portfolioservice.PortfolioService) *Service {
	return &Service{
		DB: db, Market: market, Activities: activities, Portfolio: report,
		defaultProvider: market.Provider(),
	}
}
func validCurrency(v string) bool {
	for _, c := range currencymodel.SupportedCurrencies {
		if c == v {
			return true
		}
	}
	return false
}
func validDate(v string) bool {
	d, e := time.Parse("2006-01-02", v)
	return e == nil && !d.After(portfolio.Day(time.Now())) && d.Year() >= 1900
}
func validAmount(v float64) bool { return portfolio.ValidNumber(v) && v >= 0 }
func (s *Service) Account(user, id uuid.UUID) (accountmodel.Account, error) {
	var a accountmodel.Account
	err := s.DB.First(&a, "id = ? AND user_id = ?", id, user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrNotFound
	}
	return a, err
}
func (s *Service) Watchlist(user uuid.UUID) ([]model.WatchItem, error) {
	rows := []model.WatchItem{}
	err := s.DB.Where("user_id = ?", user).Order("created_at desc, id").Find(&rows).Error
	return rows, err
}
func (s *Service) AddWatch(user uuid.UUID, item model.WatchItem) (*model.WatchItem, error) {
	item.ID = uuid.New()
	item.UserID = user
	item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
	item.Name = strings.TrimSpace(item.Name)
	if !symbolPattern.MatchString(item.Symbol) || len(item.Name) > 100 || len(item.AssetClass) > 40 || !validCurrency(item.Currency) {
		return nil, errors.New("invalid watchlist asset")
	}
	if strings.HasPrefix(item.Symbol, "MANUAL-") {
		var a model.ManualAsset
		if err := s.DB.First(&a, "symbol = ? AND user_id = ?", item.Symbol, user).Error; err != nil {
			return nil, ErrNotFound
		}
	}
	var count int64
	if err := s.DB.Model(&model.WatchItem{}).Where("user_id = ?", user).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= 500 {
		return nil, errors.New("watchlist limit is 500 assets")
	}
	err := s.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "symbol"}}, DoNothing: true}).Create(&item).Error
	if err == nil {
		item.ID = uuid.Nil
		err = s.DB.First(&item, "user_id = ? AND symbol = ?", user, item.Symbol).Error
	}
	return &item, err
}
func (s *Service) DeleteWatch(user, id uuid.UUID) error {
	r := s.DB.Where("id = ? AND user_id = ?", id, user).Delete(&model.WatchItem{})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type AssetRecord struct {
	model.ManualAsset
	Valuations []model.ManualValuation `json:"valuations"`
}

func (s *Service) Assets(user uuid.UUID) ([]AssetRecord, error) {
	assets := []model.ManualAsset{}
	if err := s.DB.Where("user_id = ?", user).Order("name,id").Find(&assets).Error; err != nil {
		return nil, err
	}
	result := make([]AssetRecord, 0, len(assets))
	for _, a := range assets {
		v := []model.ManualValuation{}
		if err := s.DB.Where("asset_id = ?", a.ID).Order("date desc").Find(&v).Error; err != nil {
			return nil, err
		}
		result = append(result, AssetRecord{a, v})
	}
	return result, nil
}
func (s *Service) SaveAsset(user uuid.UUID, id *uuid.UUID, input model.ManualAsset) (*model.ManualAsset, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 100 || !validCurrency(input.Currency) || len(input.AssetClass) > 40 || len(input.Sector) > 100 || len(input.Country) > 100 {
		return nil, errors.New("invalid manual asset")
	}
	var asset model.ManualAsset
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if id == nil {
			input.ID = uuid.New()
			input.UserID = user
			input.Symbol = "MANUAL-" + strings.ToUpper(strings.ReplaceAll(input.ID.String(), "-", "")[:12])
			asset = input
			return tx.Create(&asset).Error
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&asset, "id = ? AND user_id = ?", *id, user).Error; err != nil {
			return ErrNotFound
		}
		if input.Currency != asset.Currency {
			return errors.New("asset currency cannot be changed")
		}
		if err := tx.Model(&asset).Updates(map[string]any{"name": input.Name, "asset_class": input.AssetClass, "sector": input.Sector, "country": input.Country}).Error; err != nil {
			return err
		}
		if err := tx.Model(&holdingmodel.Holding{}).Where("symbol = ? AND account_id IN (?)", asset.Symbol, tx.Model(&accountmodel.Account{}).Select("id").Where("user_id = ?", user)).Updates(map[string]any{"name": input.Name, "asset_class": input.AssetClass, "sector": input.Sector, "country": input.Country}).Error; err != nil {
			return err
		}
		return tx.First(&asset, "id = ?", asset.ID).Error
	})
	return &asset, err
}
func (s *Service) SaveValuation(user, id uuid.UUID, date string, price float64) error {
	if !validDate(date) || !validAmount(price) {
		return errors.New("invalid valuation date or price")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var a model.ManualAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ? AND user_id = ?", id, user).Error; err != nil {
			return ErrNotFound
		}
		v := model.ManualValuation{AssetID: id, Date: date, Price: price}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "asset_id"}, {Name: "date"}}, DoUpdates: clause.AssignmentColumns([]string{"price", "updated_at"})}).Create(&v).Error
	})
}
func (s *Service) DeleteAsset(user, id uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var a model.ManualAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ? AND user_id = ?", id, user).Error; err != nil {
			return ErrNotFound
		}
		var n int64
		if err := tx.Model(&activitymodel.Activity{}).Where("symbol = ?", a.Symbol).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errors.New("this asset has financial history and cannot be deleted")
		}
		return tx.Delete(&a).Error
	})
}

type LiabilityRecord struct {
	model.Liability
	Balances []model.LiabilityBalance `json:"balances"`
}

func (s *Service) Liabilities(user uuid.UUID) ([]LiabilityRecord, error) {
	items := []model.Liability{}
	if err := s.DB.Where("user_id = ?", user).Order("created_at,id").Find(&items).Error; err != nil {
		return nil, err
	}
	out := []LiabilityRecord{}
	for _, item := range items {
		balances := []model.LiabilityBalance{}
		if err := s.DB.Where("liability_id = ?", item.ID).Order("date desc").Find(&balances).Error; err != nil {
			return nil, err
		}
		out = append(out, LiabilityRecord{item, balances})
	}
	return out, nil
}
func (s *Service) CreateLiability(user uuid.UUID, name, currency, date string, amount float64) (*model.Liability, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 || !validCurrency(currency) || !validDate(date) || !validAmount(amount) {
		return nil, errors.New("invalid liability")
	}
	item := model.Liability{UserID: user, Name: name, Currency: currency}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return tx.Create(&model.LiabilityBalance{LiabilityID: item.ID, Date: date, Amount: amount}).Error
	})
	return &item, err
}
func (s *Service) SaveLiabilityBalance(user, id uuid.UUID, date string, amount float64) error {
	if !validDate(date) || !validAmount(amount) {
		return errors.New("invalid balance")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var item model.Liability
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ? AND user_id = ?", id, user).Error; err != nil {
			return ErrNotFound
		}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "liability_id"}, {Name: "date"}}, DoUpdates: clause.AssignmentColumns([]string{"amount", "updated_at"})}).Create(&model.LiabilityBalance{LiabilityID: id, Date: date, Amount: amount}).Error
	})
}
