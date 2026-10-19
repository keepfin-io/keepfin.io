package database

import (
	"log"
	"os"

	"keepnode/app/models"
	"keepnode/pkg/utils"

	"gorm.io/gorm"
)

// SeedDefaultUser creates a default user when DEFAULT_USER_EMAIL and DEFAULT_USER_PASSWORD are set.
func SeedDefaultUser(db *gorm.DB) error {
	email := os.Getenv("DEFAULT_USER_EMAIL")
	password := os.Getenv("DEFAULT_USER_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	// -----> Unscoped also counts soft-deleted rows, because email index is unique
	var count int64
	if err := db.Unscoped().Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	user := &models.User{
		Base:         models.Base{IsDummy: true},
		Email:        email,
		PasswordHash: utils.GeneratePassword(password),
		UserStatus:   1,
		UserRole:     "user",
		Theme:        models.ThemeLight,
		BaseCurrency: "USD",
		Locale:       models.LocaleEnUS,
	}

	if err := db.Create(user).Error; err != nil {
		return err
	}

	log.Printf("Default user created: %s", email)
	return nil
}
