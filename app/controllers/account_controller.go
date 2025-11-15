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

func (ctrl *AccountController) CreateAccount(c *fiber.Ctx) error {
	// Implementation for creating an account
	return nil
}
