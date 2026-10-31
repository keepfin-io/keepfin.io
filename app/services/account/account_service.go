package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	accountquery "keepnode/app/queries/account"
	activityquery "keepnode/app/queries/activity"
	"keepnode/pkg/portfolio"
	"keepnode/pkg/utils"
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
func (s *AccountService) CreateAccount(userID uuid.UUID, account *accountmodel.Account) (*accountmodel.Account, error) {
	// Set user ID and timestamps FIRST
	account.ID = uuid.New()
	account.UserID = userID
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	account.LedgerReady = true
	if account.OpeningDate == nil {
		d := portfolio.Day(time.Now())
		account.OpeningDate = &d
	}
	if account.OpeningDate.After(time.Now()) {
		return nil, errors.New("opening date cannot be in the future")
	}
	if !portfolio.ValidNumber(account.InitialBalance) {
		return nil, errors.New("invalid opening balance")
	}
	// Validate account data AFTER setting required fields
	validate := utils.NewValidator()
	if err := validate.Struct(account); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	// Create account in database
	if err := (&accountquery.AccountQueries{DB: s.DB}).CreateAccount(account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountByID retrieves an account by ID
func (s *AccountService) GetAccountByID(accountID uuid.UUID, userID uuid.UUID) (*accountmodel.Account, error) {
	account, err := (&accountquery.AccountQueries{DB: s.DB}).GetAccountByID(accountID)
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

	account.CashBalance, err = accountCash(s.DB, account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetAccountsByUserID retrieves all accounts for a user
func (s *AccountService) GetAccountsByUserID(userID uuid.UUID) ([]accountmodel.Account, error) {
	accounts, err := (&accountquery.AccountQueries{DB: s.DB}).GetAccountsByUserID(userID)
	if err != nil {
		return nil, err
	}

	for i := range accounts {
		accounts[i].CashBalance, err = accountCash(s.DB, accounts[i])
		if err != nil {
			return nil, err
		}
	}
	return accounts, nil
}

// UpdateAccount updates an existing account
func (s *AccountService) UpdateAccount(accountID uuid.UUID, userID uuid.UUID, updateData *accountmodel.Account) (*accountmodel.Account, error) {
	// Check if account exists and user owns it
	existingAccount, err := s.GetAccountByID(accountID, userID)
	if err != nil {
		return nil, err
	}

	if updateData.Currency != existingAccount.Currency {
		return nil, errors.New("account currency cannot be changed; create another account and transfer cash")
	}
	if !portfolio.ValidNumber(updateData.InitialBalance) {
		return nil, errors.New("invalid opening balance")
	}
	if updateData.OpeningDate == nil {
		updateData.OpeningDate = existingAccount.OpeningDate
	}
	if updateData.OpeningDate != nil && updateData.OpeningDate.After(time.Now()) {
		return nil, errors.New("opening date cannot be in the future")
	}
	updateData.LedgerReady = true
	// Preserve certain fields BEFORE validation
	updateData.ID = existingAccount.ID
	updateData.UserID = existingAccount.UserID
	updateData.CreatedAt = existingAccount.CreatedAt
	updateData.UpdatedAt = time.Now()

	// Validate update data AFTER setting required fields
	validate := utils.NewValidator()
	if err := validate.Struct(updateData); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	// Update account in database
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := accountquery.GetAccountForUpdate(tx, accountID); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&activitymodel.Activity{}).Where("account_id = ? AND ledger = true AND date < ?", accountID, updateData.OpeningDate).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("opening date must not follow existing activities")
		}
		return (&accountquery.AccountQueries{DB: tx}).UpdateAccount(accountID, updateData)
	}); err != nil {
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
	if err := (&accountquery.AccountQueries{DB: s.DB}).DeleteAccount(accountID); err != nil {
		return err
	}

	return nil
}

// ToggleAccountStatus toggles the is_active status of an account
func (s *AccountService) ToggleAccountStatus(accountID uuid.UUID, userID uuid.UUID) (*accountmodel.Account, error) {
	// Check if account exists and user owns it
	account, err := s.GetAccountByID(accountID, userID)
	if err != nil {
		return nil, err
	}

	// Toggle the status
	newStatus := !account.IsActive

	// Update in database
	if err := (&accountquery.AccountQueries{DB: s.DB}).UpdateAccountStatus(accountID, newStatus); err != nil {
		return nil, err
	}

	// Return updated account
	return s.GetAccountByID(accountID, userID)
}

// Cash is available independently of market prices and opening-position FX.
func accountCash(db *gorm.DB, a accountmodel.Account) (float64, error) {
	events, err := activityquery.GetLedgerEvents(db, a.ID)
	if err != nil {
		return 0, err
	}
	cash := a.InitialBalance
	for _, e := range events {
		amount := portfolio.Mul(portfolio.Mul(e.Quantity, e.UnitPrice), e.ExchangeRate)
		fee := portfolio.Mul(e.Fee, e.ExchangeRate)
		switch e.Type {
		case "buy", "withdrawal", "transfer_out", "fee":
			cash = portfolio.Add(cash, -amount-fee)
		case "sell", "deposit", "transfer_in", "dividend", "interest":
			cash = portfolio.Add(cash, amount-fee)
		}
	}
	return cash, nil
}
