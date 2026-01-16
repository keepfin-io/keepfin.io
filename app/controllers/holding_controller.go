package controllers

import (
	"keepnode/app/models"
	"keepnode/app/services"
	"keepnode/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type HoldingController struct {
	holdingService *services.HoldingService
	accountService *services.AccountService
}

func NewHoldingController(holdingService *services.HoldingService, accountService *services.AccountService) *HoldingController {
	return &HoldingController{
		holdingService: holdingService,
		accountService: accountService,
	}
}

// AddHolding adds a new holding to an account
// @Summary Add holding to account
// @Tags Holdings
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param holding body models.CreateHoldingRequest true "Holding data"
// @Success 201 {object} models.Holding
// @Router /api/v1/accounts/{id}/holdings [post]
func (ctrl *HoldingController) AddHolding(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Account not found",
		})
	}

	req := &models.CreateHoldingRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	holding, err := ctrl.holdingService.AddHolding(accountID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "Holding added successfully",
		"holding": holding,
	})
}

// GetHoldings returns all holdings for an account with live data
// @Summary Get account holdings with live data
// @Tags Holdings
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {array} models.HoldingWithLiveData
// @Router /api/v1/accounts/{id}/holdings [get]
func (ctrl *HoldingController) GetHoldings(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Account not found",
		})
	}

	holdings, summary, err := ctrl.holdingService.GetHoldingsWithLiveData(accountID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch holdings",
		})
	}

	return c.JSON(fiber.Map{
		"error":    false,
		"holdings": holdings,
		"summary":  summary,
	})
}

// UpdateHolding updates a holding
// @Summary Update holding
// @Tags Holdings
// @Accept json
// @Produce json
// @Param id path string true "Holding ID"
// @Param holding body models.UpdateHoldingRequest true "Update data"
// @Success 200 {object} models.Holding
// @Router /api/v1/holdings/{id} [put]
func (ctrl *HoldingController) UpdateHolding(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	holdingIDStr := c.Params("id")
	holdingID, err := uuid.Parse(holdingIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid holding ID format",
		})
	}

	accountIDStr := c.Params("accountId")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Account not found",
		})
	}

	req := &models.UpdateHoldingRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	holding, err := ctrl.holdingService.UpdateHolding(holdingID, accountID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Holding updated successfully",
		"holding": holding,
	})
}

// DeleteHolding deletes a holding
// @Summary Delete holding
// @Tags Holdings
// @Produce json
// @Param id path string true "Holding ID"
// @Success 200 {string} string "Holding deleted successfully"
// @Router /api/v1/holdings/{id} [delete]
func (ctrl *HoldingController) DeleteHolding(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	holdingIDStr := c.Params("id")
	holdingID, err := uuid.Parse(holdingIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid holding ID format",
		})
	}

	accountIDStr := c.Params("accountId")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Account not found",
		})
	}

	err = ctrl.holdingService.DeleteHolding(holdingID, accountID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Holding deleted successfully",
	})
}
