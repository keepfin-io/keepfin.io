package queries

import (
	"keepnode/app/models"

	"gorm.io/gorm"
)

type AccountQueries struct {
	DB *gorm.DB
}

func (q *AccountQueries) CreateAccount(account *models.Account) error {
	err := q.DB.Create(&account).Error
	if err != nil {
		return err
	}
	return nil
}
