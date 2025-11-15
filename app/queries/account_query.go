package queries

import (
	"keepnode/app/models"

	"gorm.io/gorm"
)

type AccountQueries struct {
	DB *gorm.DB
}

func (q *AccountQueries) GetAccountByID(accountID string) (models.Account, error) {
	account := models.Account{}
	err := q.DB.Where("id = ?", accountID).First(&account).Error
	if err != nil {
		return account, err
	}
	return account, nil
}

func (q *AccountQueries) CreateAccount(account *models.Account) error {
	err := q.DB.Create(&account).Error
	if err != nil {
		return err
	}
	return nil
}

func (q *AccountQueries) UpdateAccount(account *models.Account) error {
	err := q.DB.Save(&account).Error
	if err != nil {
		return err
	}
	return nil
}

func (q *AccountQueries) DeleteAccount(accountID string) error {
	err := q.DB.Delete(&models.Account{}, "id = ?", accountID).Error
	if err != nil {
		return err
	}
	return nil
}
