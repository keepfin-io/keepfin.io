package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"keepnode/app/models"
	"keepnode/pkg/portfolio"
	"time"
)

// initializePortfolios preserves existing cash and holdings as opening positions.
// Old activities stay visible but do not replay against the current snapshots.
// The marker and opening rows commit together, making restarts safe.
func initializePortfolios(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var accounts []models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("ledger_ready = false").Find(&accounts).Error; err != nil {
			return err
		}
		now := portfolio.Day(time.Now())
		for _, a := range accounts {
			var holdings []models.Holding
			if err := tx.Where("account_id = ?", a.ID).Find(&holdings).Error; err != nil {
				return err
			}
			for _, h := range holdings {
				rate := 1.0
				if h.BuyCurrency != a.Currency {
					rate = 0
				}
				activity := models.Activity{AccountID: a.ID, HoldingID: &h.ID, Symbol: h.Symbol, Name: h.Name, Type: "opening", Date: now, Quantity: h.Quantity, UnitPrice: h.BuyPrice, Currency: h.BuyCurrency, ExchangeRate: rate, Ledger: true, Notes: "Opening position; earlier activities retained as history"}
				if err := tx.Create(&activity).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&a).Updates(map[string]any{"ledger_ready": true, "opening_date": now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
