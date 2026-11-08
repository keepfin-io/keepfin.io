package database

import (
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	holdingmodel "keepnode/app/models/holding"
	marketmodel "keepnode/app/models/market"
	usermodel "keepnode/app/models/user"
	wealthmodel "keepnode/app/models/wealth"
)

// AutoMigrate runs auto migration for all models.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&usermodel.User{},
		&accountmodel.Account{},
		&holdingmodel.Holding{},
		&activitymodel.Activity{},
		&marketmodel.MarketPrice{},
		&holdingmodel.SearchState{},
		&wealthmodel.WatchItem{}, &wealthmodel.ManualAsset{}, &wealthmodel.ManualValuation{},
		&wealthmodel.Liability{}, &wealthmodel.LiabilityBalance{}, &wealthmodel.Tag{}, &wealthmodel.TagAssignment{},
		&wealthmodel.Preferences{}, &wealthmodel.Share{}, &wealthmodel.MarketJob{}, &wealthmodel.MarketConfig{}, &wealthmodel.PriceOverride{},
	); err != nil {
		return err
	}
	if err := initializePortfolios(db); err != nil {
		return err
	}
	if err := initializeHoldingSearch(db); err != nil {
		return err
	}
	return initializeWorkerNotifications(db)
}
