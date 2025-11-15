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

func (s *AccountService) GetAccountByID(accountID string) (models.Account, error) {
	account, err := (&queries.AccountQueries{DB: s.DB}).GetAccountByID(accountID)
	if err != nil {
		return models.Account{}, err
	}
	return account, nil
}

func (s *AccountService) CreateAccount(account *models.Account) error {
	err := (&queries.AccountQueries{DB: s.DB}).CreateAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccountService) UpdateAccount(account *models.Account) error {
	err := (&queries.AccountQueries{DB: s.DB}).UpdateAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccountService) DeleteAccount(accountID string) error {
	err := (&queries.AccountQueries{DB: s.DB}).DeleteAccount(accountID)
	if err != nil {
		return err
	}
	return nil
}
