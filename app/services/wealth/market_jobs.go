package wealth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	holdingmodel "keepnode/app/models/holding"
	marketmodel "keepnode/app/models/market"
	usermodel "keepnode/app/models/user"
	model "keepnode/app/models/wealth"
	"strings"
	"time"
)

func (s *Service) IsAdmin(user uuid.UUID) bool {
	var n int64
	return s.DB.Model(&usermodel.User{}).Where("id = ? AND user_role = 'admin' AND user_status = 1", user).Count(&n).Error == nil && n == 1
}
func (s *Service) MarketConfiguration() (map[string]any, error) {
	if err := s.LoadProvider(); err != nil {
		return nil, err
	}
	return map[string]any{"provider": s.Market.Provider(), "providers": []map[string]any{{"id": "eodhd", "configured": s.Market.Configured("eodhd")}, {"id": "alpha-vantage", "configured": s.Market.Configured("alpha-vantage")}}}, nil
}
func (s *Service) LoadProvider() error {
	var config model.MarketConfig
	err := s.DB.First(&config, "id = 1").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.Market.SetProvider(config.Provider)
}
func (s *Service) SetProvider(user uuid.UUID, name string) error {
	if !s.IsAdmin(user) {
		return ErrNotFound
	}
	if !s.Market.Configured(name) {
		return errors.New("configure the provider API key on the server first")
	}
	if name != "eodhd" && name != "alpha-vantage" {
		return errors.New("invalid provider")
	}
	if err := s.DB.Save(&model.MarketConfig{ID: 1, Provider: name}).Error; err != nil {
		return err
	}
	return s.Market.SetProvider(name)
}
func (s *Service) Jobs(user uuid.UUID) ([]model.MarketJob, error) {
	rows := []model.MarketJob{}
	e := s.DB.Where("user_id = ?", user).Order("created_at desc").Limit(100).Find(&rows).Error
	return rows, e
}
func (s *Service) QueueJob(user uuid.UUID, symbol, from, to string) (*model.MarketJob, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if !symbolPattern.MatchString(symbol) || strings.HasPrefix(symbol, "MANUAL-") || !validDate(from) || !validDate(to) || from > to {
		return nil, errors.New("invalid market data request")
	}
	start, _ := time.Parse("2006-01-02", from)
	end, _ := time.Parse("2006-01-02", to)
	if end.Sub(start) > 3660*24*time.Hour {
		return nil, errors.New("request up to ten years at a time")
	}
	var n int64
	if err := s.DB.Model(&holdingmodel.Holding{}).Joins("JOIN accounts a ON a.id = holdings.account_id AND a.deleted_at IS NULL").Where("a.user_id = ? AND holdings.symbol = ?", user, symbol).Count(&n).Error; err != nil {
		return nil, err
	}
	if n == 0 {
		if err := s.DB.Model(&model.WatchItem{}).Where("user_id = ? AND symbol = ?", user, symbol).Count(&n).Error; err != nil {
			return nil, err
		}
	}
	if n == 0 && !s.IsAdmin(user) {
		return nil, errors.New("add this asset to your watchlist or accounts first")
	}
	var job model.MarketJob
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Serialize this user's requests so limits and duplicate checks are atomic.
		var owner usermodel.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&owner, "id = ?", user).Error; err != nil {
			return err
		}
		e := tx.Where("user_id = ? AND symbol = ? AND \"from\" = ? AND \"to\" = ?", user, symbol, from, to).Order("created_at desc").First(&job).Error
		if e == nil {
			return nil
		}
		if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if e = tx.Model(&model.MarketJob{}).Where("user_id = ? AND status IN ?", user, []string{"queued", "running"}).Count(&n).Error; e != nil {
			return e
		}
		if n >= 20 {
			return errors.New("wait for pending jobs before adding more")
		}
		job = model.MarketJob{UserID: user, Symbol: symbol, From: from, To: to, Status: "queued", AvailableAt: time.Now().UTC()}
		return tx.Create(&job).Error
	})
	return &job, err
}
func (s *Service) RetryJob(user, id uuid.UUID) error {
	result := s.DB.Model(&model.MarketJob{}).Where("id = ? AND user_id = ? AND status = 'failed'", id, user).Updates(map[string]any{"status": "queued", "attempts": 0, "error": "", "available_at": time.Now().UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("only failed jobs can be retried")
	}
	return nil
}
func (s *Service) RunMarketJob(ctx context.Context) error {
	db := s.DB.WithContext(ctx)
	var job model.MarketJob
	err := db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Model(&model.MarketJob{}).Where("status = 'running' AND updated_at < ?", time.Now().Add(-5*time.Minute)).Updates(map[string]any{"status": gorm.Expr("CASE WHEN attempts >= 3 THEN 'failed' ELSE 'queued' END"), "available_at": time.Now().UTC()}).Error; e != nil {
			return e
		}
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = 'queued' AND available_at <= ?", time.Now().UTC()).Order("available_at,id").First(&job).Error; e != nil {
			return e
		}
		return tx.Model(&job).Updates(map[string]any{"status": "running", "attempts": job.Attempts + 1}).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	data, fetchErr := s.Market.GetHistoricalData(&marketmodel.EODRequest{Symbol: job.Symbol, From: job.From, To: job.To, Period: "d", Order: "a"})
	if fetchErr == nil && len(data) == 0 {
		fetchErr = errors.New("provider returned no prices for this range")
	}
	if fetchErr == nil {
		rows := []marketmodel.MarketPrice{}
		for _, v := range data {
			if validDate(v.Date) && v.Date >= job.From && v.Date <= job.To && validAmount(v.Close) && v.Close > 0 {
				rows = append(rows, marketmodel.MarketPrice{Symbol: job.Symbol, Date: v.Date, Close: v.Close})
			}
		}
		if len(rows) == 0 {
			fetchErr = errors.New("provider returned no valid prices")
		} else {
			fetchErr = db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(rows, 500).Error
		}
	}
	update := map[string]any{"status": "complete", "error": ""}
	if fetchErr != nil {
		update["error"] = "Market data could not be refreshed; check the provider configuration, plan and quota."
		update["status"] = "failed"
		if job.Attempts < 3 {
			update["status"] = "queued"
			update["available_at"] = time.Now().Add(time.Duration(job.Attempts*job.Attempts) * time.Minute)
		}
	}
	return db.Model(&model.MarketJob{}).Where("id = ? AND status = 'running'", job.ID).Updates(update).Error
}
func (s *Service) RunMarketWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.LoadProvider() == nil {
				_ = s.RunMarketJob(ctx)
			}
		}
	}
}
func (s *Service) Overrides(user uuid.UUID) ([]model.PriceOverride, error) {
	if !s.IsAdmin(user) {
		return nil, ErrNotFound
	}
	rows := []model.PriceOverride{}
	e := s.DB.Order("updated_at desc").Limit(200).Find(&rows).Error
	return rows, e
}
func (s *Service) CorrectPrice(user uuid.UUID, v model.PriceOverride) error {
	if !s.IsAdmin(user) {
		return ErrNotFound
	}
	v.Symbol = strings.ToUpper(strings.TrimSpace(v.Symbol))
	v.Reason = strings.TrimSpace(v.Reason)
	if !symbolPattern.MatchString(v.Symbol) || strings.HasPrefix(v.Symbol, "MANUAL-") || !validDate(v.Date) || !validAmount(v.Close) || v.Close <= 0 || v.Reason == "" || len(v.Reason) > 500 {
		return errors.New("a symbol, past date, positive price and reason are required")
	}
	v.EditorID = user
	return s.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&v).Error
}
func (s *Service) DeleteCorrection(user uuid.UUID, symbol, date string) error {
	if !s.IsAdmin(user) {
		return ErrNotFound
	}
	r := s.DB.Where("symbol = ? AND date = ?", symbol, date).Delete(&model.PriceOverride{})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
