package database

import (
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	holdingmodel "keepnode/app/models/holding"
	marketmodel "keepnode/app/models/market"
	usermodel "keepnode/app/models/user"
)

// AutoMigrate runs auto migration for all models.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&usermodel.User{},
		&accountmodel.Account{},
		&holdingmodel.Holding{},
		&activitymodel.Activity{},
		&marketmodel.MarketPrice{},
	); err != nil {
		return err
	}
	return initializePortfolios(db)
}
