package queries

import (
	"keepnode/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountQueries struct for queries from Account model.
type AccountQueries struct {
	DB *gorm.DB
}

// GetAccountByID query for getting one Account by given ID.
func (q *AccountQueries) GetAccountByID(accountID uuid.UUID) (models.Account, error) {
	// Define Account variable.
	account := models.Account{}

	// Send query to database.
	err := q.DB.Where("id = ?", accountID).First(&account).Error
	if err != nil {
		// Return empty object and error.
		return account, err
	}

	// Return query result.
	return account, nil
}

// GetAccountsByUserID query for getting all Accounts by given User ID.
func (q *AccountQueries) GetAccountsByUserID(userID uuid.UUID) ([]models.Account, error) {
	// Define Accounts slice.
	var accounts []models.Account

	// Send query to database.
	err := q.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&accounts).Error
	if err != nil {
		// Return empty slice and error.
		return accounts, err
	}

	// Return query result.
	return accounts, nil
}

// CreateAccount query for creating a new account.
func (q *AccountQueries) CreateAccount(account *models.Account) error {
	err := q.DB.Create(&account).Error
	if err != nil {
		return err
	}

	return nil
}

// UpdateAccount query for updating account by given ID.
func (q *AccountQueries) UpdateAccount(accountID uuid.UUID, account *models.Account) error {
	// Update account in database.
	err := q.DB.Model(&models.Account{}).Where("id = ?", accountID).Select("account_name", "initial_balance", "platform", "description", "is_excluded", "opening_date").Updates(account).Error
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// DeleteAccount query for deleting account by given ID.
func (q *AccountQueries) DeleteAccount(accountID uuid.UUID) error {
	// Delete account in database.
	err := q.DB.Delete(&models.Account{}, accountID).Error
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateAccountStatus query for updating account status by given ID.
func (q *AccountQueries) UpdateAccountStatus(accountID uuid.UUID, isActive bool) error {
	err := q.DB.Model(&models.Account{}).Where("id = ?", accountID).Update("is_active", isActive).Error
	if err != nil {
		return err
	}
	return nil
}
