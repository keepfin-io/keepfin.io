package services

import (
	"errors"
	"time"

	"keepnode/app/models"
	"keepnode/app/queries"
	"keepnode/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountService handles account-related business logic
type AccountService struct {
	DB *gorm.DB
}

// NewAccountService creates a new AccountService instance
func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{DB: db}
}

// CreateAccount creates a new account for a user
func (s *AccountService) CreateAccount(userID uuid.UUID, account *models.Account) (*models.Account, error) {
	// Set user ID and timestamps FIRST
	account.ID = uuid.New()
	account.UserID = userID
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	// Validate account data AFTER setting required fields
	validate := utils.NewValidator()
	if err := validate.Struct(account); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	// Create account in database
	if err := (&queries.AccountQueries{DB: s.DB}).CreateAccount(account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountByID retrieves an account by ID
func (s *AccountService) GetAccountByID(accountID uuid.UUID, userID uuid.UUID) (*models.Account, error) {
	account, err := (&queries.AccountQueries{DB: s.DB}).GetAccountByID(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	// Verify ownership
	if account.UserID != userID {
		return nil, errors.New("unauthorized access to account")
	}

	return &account, nil
}

// GetAccountsByUserID retrieves all accounts for a user
func (s *AccountService) GetAccountsByUserID(userID uuid.UUID) ([]models.Account, error) {
	accounts, err := (&queries.AccountQueries{DB: s.DB}).GetAccountsByUserID(userID)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(accountID uuid.UUID, userID uuid.UUID, updateData *models.Account) (*models.Account, error) {
	// Check if account exists and user owns it
	existingAccount, err := s.GetAccountByID(accountID, userID)
	if err != nil {
		return nil, err
	}

	// Validate update data
	validate := utils.NewValidator()
	if err := validate.Struct(updateData); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	// Preserve certain fields
	updateData.ID = existingAccount.ID
	updateData.UserID = existingAccount.UserID
	updateData.CreatedAt = existingAccount.CreatedAt
	updateData.UpdatedAt = time.Now()

	// Update account in database
	if err := (&queries.AccountQueries{DB: s.DB}).UpdateAccount(accountID, updateData); err != nil {
		return nil, err
	}

	// Return updated account
	return s.GetAccountByID(accountID, userID)
}

// DeleteAccount deletes an account
func (s *AccountService) DeleteAccount(accountID uuid.UUID, userID uuid.UUID) error {
	// Check if account exists and user owns it
	_, err := s.GetAccountByID(accountID, userID)
	if err != nil {
		return err
	}

	// Delete account
	if err := (&queries.AccountQueries{DB: s.DB}).DeleteAccount(accountID); err != nil {
		return err
	}

	return nil
}
