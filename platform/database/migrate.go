package database

import (
	"keepnode/app/models"

	"gorm.io/gorm"
)

// AutoMigrate runs auto migration for all models.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Holding{},
		&models.Activity{},
		&models.MarketPrice{},
	); err != nil {
		return err
	}
	return initializePortfolios(db)
}
