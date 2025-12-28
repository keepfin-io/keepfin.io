package controllers

import (
	"keepnode/app/models"
	"keepnode/app/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AccountController handles account-related requests
type AccountController struct {
	accountService *services.AccountService
}

// NewAccountController creates a new AccountController instance
func NewAccountController(accountService *services.AccountService) *AccountController {
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
func (ctrl *AccountController) CreateAccount(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	account := &models.Account{}
	if err := c.BodyParser(account); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	createdAccount, err := ctrl.accountService.CreateAccount(userID, account)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "Account created successfully",
		"account": createdAccount,
	})
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
func (ctrl *AccountController) GetAccount(c *fiber.Ctx) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Call service layer
	account, err := ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"error":   false,
		"account": account,
	})
}

// GetUserAccounts method to get all accounts for the current user.
// @Description Get all accounts for the current user.
// @Summary get user accounts
// @Tags Account
// @Accept json
// @Produce json
// @Success 200 {array} models.Account
// @Router /api/v1/accounts [get]
func (ctrl *AccountController) GetUserAccounts(c *fiber.Ctx) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	// Call service layer
	accounts, err := ctrl.accountService.GetAccountsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to retrieve accounts",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"error":    false,
		"accounts": accounts,
	})
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
func (ctrl *AccountController) UpdateAccount(c *fiber.Ctx) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Parse request body
	updateData := &models.Account{}
	if err := c.BodyParser(updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Call service layer
	updatedAccount, err := ctrl.accountService.UpdateAccount(accountID, userID, updateData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Account updated successfully",
		"account": updatedAccount,
	})
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
func (ctrl *AccountController) DeleteAccount(c *fiber.Ctx) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Call service layer
	err = ctrl.accountService.DeleteAccount(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Account deleted successfully",
	})
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
func (ctrl *AccountController) ToggleAccountStatus(c *fiber.Ctx) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	// Get account ID from URL params
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Call service layer
	account, err := ctrl.accountService.ToggleAccountStatus(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Account status updated successfully",
		"account": account,
	})
}
