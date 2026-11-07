package account

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
)

// AccountQueries struct for queries from Account model.
type AccountQueries struct {
	DB *gorm.DB
}

// GetAccountByID query for getting one Account by given ID.
func (q *AccountQueries) GetAccountByID(accountID uuid.UUID) (accountmodel.Account, error) {
	// Define Account variable.
	account := accountmodel.Account{}

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
func (q *AccountQueries) GetAccountsByUserID(userID uuid.UUID) ([]accountmodel.Account, error) {
	// Define Accounts slice.
	var accounts []accountmodel.Account

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
func (q *AccountQueries) CreateAccount(account *accountmodel.Account) error {
	err := q.DB.Create(&account).Error
	if err != nil {
		return err
	}

	return nil
}

// UpdateAccount query for updating account by given ID.
func (q *AccountQueries) UpdateAccount(accountID uuid.UUID, account *accountmodel.Account) error {
	// Update account in database.
	err := q.DB.Model(&accountmodel.Account{}).Where("id = ?", accountID).Select("account_name", "initial_balance", "platform", "description", "is_excluded", "opening_date").Updates(account).Error
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
	err := q.DB.Delete(&accountmodel.Account{}, accountID).Error
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateAccountStatus query for updating account status by given ID.
func (q *AccountQueries) UpdateAccountStatus(accountID uuid.UUID, isActive bool) error {
	err := q.DB.Model(&accountmodel.Account{}).Where("id = ?", accountID).Update("is_active", isActive).Error
	if err != nil {
		return err
	}
	return nil
}
