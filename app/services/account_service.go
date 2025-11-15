package services

import (
	"keepnode/app/models"
	"keepnode/app/queries"

	"gorm.io/gorm"
)

type AccountService struct {
	DB *gorm.DB
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{DB: db}
}

func (s *AccountService) CreateAccount(account *models.Account) error {
	err := (&queries.AccountQueries{DB: s.DB}).CreateAccount(account)
	if err != nil {
		return err
	}
	return nil
}
