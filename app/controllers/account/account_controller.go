package account

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"keepnode/pkg/apiresponse"

	accountmodel "keepnode/app/models/account"
	accountservice "keepnode/app/services/account"
)

// AccountController handles account-related requests
type AccountController struct {
	accountService *accountservice.AccountService
}

// NewAccountController creates a new AccountController instance
func NewAccountController(accountService *accountservice.AccountService) *AccountController {
	return &AccountController{
		accountService: accountService,
	}
}

// CreateAccount method to create a new account.
// @Description Create a new account.
// @Summary create a new account
// @Tags Account
// @Accept json
// @Produce json
// @Param account body models.Account true "Account data"
// @Success 201 {object} models.Account
// @Router /api/v1/accounts/create [post]
func (ctrl *AccountController) CreateAccount(c *fiber.Ctx, userID uuid.UUID) error {
	account := &accountmodel.Account{}
	if err := c.BodyParser(account); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	createdAccount, err := ctrl.accountService.CreateAccount(userID, account)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondCreated(c, createdAccount)
}

// GetAccount method to get an account by ID.
// @Description Get an account by ID.
// @Summary get an account by ID
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Router /api/v1/accounts/{id} [get]
func (ctrl *AccountController) GetAccount(c *fiber.Ctx, userID uuid.UUID) error {
	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Call service layer
	account, err := ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	// Return success response
	return apiresponse.RespondOK(c, account)
}

// GetUserAccounts method to get all accounts for the current user.
// @Description Get all accounts for the current user.
// @Summary get user accounts
// @Tags Account
// @Accept json
// @Produce json
// @Success 200 {array} models.Account
// @Router /api/v1/accounts [get]
func (ctrl *AccountController) GetUserAccounts(c *fiber.Ctx, userID uuid.UUID) error {
	// Call service layer
	accounts, err := ctrl.accountService.GetAccountsByUserID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve accounts")
	}

	// Return success response
	return apiresponse.RespondOK(c, accounts)
}

// UpdateAccount method to update an account.
// @Description Update an account.
// @Summary update an account
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param account body models.Account true "Account data"
// @Success 200 {object} models.Account
// @Router /api/v1/accounts/{id} [put]
func (ctrl *AccountController) UpdateAccount(c *fiber.Ctx, userID uuid.UUID) error {
	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Parse request body
	updateData := &accountmodel.Account{}
	if err := c.BodyParser(updateData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Call service layer
	updatedAccount, err := ctrl.accountService.UpdateAccount(accountID, userID, updateData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Return success response
	return apiresponse.RespondOK(c, updatedAccount)
}

// DeleteAccount method to delete an account.
// @Description Delete an account.
// @Summary delete an account
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {string} string "Account deleted successfully"
// @Router /api/v1/accounts/{id} [delete]
func (ctrl *AccountController) DeleteAccount(c *fiber.Ctx, userID uuid.UUID) error {
	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Call service layer
	err = ctrl.accountService.DeleteAccount(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Return success response
	return apiresponse.RespondOK(c, fiber.Map{"message": "Account deleted successfully"})
}

// ToggleAccountStatus method to toggle account active status.
// @Description Toggle an account's active status.
// @Summary toggle account status
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Router /api/v1/accounts/{id}/toggle-status [patch]
func (ctrl *AccountController) ToggleAccountStatus(c *fiber.Ctx, userID uuid.UUID) error {
	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Call service layer
	account, err := ctrl.accountService.ToggleAccountStatus(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Return success response
	return apiresponse.RespondOK(c, account)
}
