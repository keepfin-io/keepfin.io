package controllers

import (
	"keepnode/app/services"

	"github.com/gofiber/fiber/v2"
)

type AccountController struct {
	accountService *services.AccountService
}

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
// @Param name body string true "Account Name"
// @Param type body string true "Account Type"
// @Success 200 {object} models.Account
// @Router /api/v1/account [post]
func (ctrl *AccountController) CreateAccount(c *fiber.Ctx) error {
	// Implementation for creating an account
	return nil
}

// GetAccount method to get an account by ID.
// @Description Get an account by ID.
// @Summary get an account by ID
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Router /api/v1/account/{id} [get]
func (ctrl *AccountController) GetAccount(c *fiber.Ctx) error {
	// Implementation for getting an account
	return nil
}

// UpdateAccount method to update an account.
// @Description Update an account.
// @Summary update an account
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param name body string false "Account Name"
// @Param type body string false "Account Type"
// @Success 200 {object} models.Account
// @Router /api/v1/account/{id} [put]
func (ctrl *AccountController) UpdateAccount(c *fiber.Ctx) error {
	// Implementation for updating an account
	return nil
}

// DeleteAccount method to delete an account.
// @Description Delete an account.
// @Summary delete an account
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {string} string "Account deleted successfully"
// @Router /api/v1/account/{id} [delete]
func (ctrl *AccountController) DeleteAccount(c *fiber.Ctx) error {
	// Implementation for deleting an account
	return nil
}
