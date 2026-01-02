package database

import (
	"keepnode/app/models"

	"gorm.io/gorm"
)

// AutoMigrate runs auto migration for all models.
func AutoMigrate(db *gorm.DB) error {
	// Auto migrate all models.
	return db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Holding{},
	)
}
