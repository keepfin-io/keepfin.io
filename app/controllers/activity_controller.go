package controllers

import (
	"keepnode/app/models"
	"keepnode/app/services"
	"keepnode/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ActivityController struct {
	activityService *services.ActivityService
	accountService  *services.AccountService
}

func NewActivityController(activityService *services.ActivityService, accountService *services.AccountService) *ActivityController {
	return &ActivityController{
		activityService: activityService,
		accountService:  accountService,
	}
}

// CreateActivity creates a new activity for an account
// @Summary Create activity
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param activity body models.CreateActivityRequest true "Activity data"
// @Success 201 {object} models.Activity
// @Router /api/v1/accounts/{id}/activities [post]
func (ctrl *ActivityController) CreateActivity(c *fiber.Ctx) error {
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

	req := &models.CreateActivityRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	activity, err := ctrl.activityService.CreateActivity(accountID, nil, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":    false,
		"message":  "Activity created successfully",
		"activity": activity,
	})
}

// GetActivities returns all activities for an account
// @Summary Get account activities
// @Tags Activities
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {array} models.Activity
// @Router /api/v1/accounts/{id}/activities [get]
func (ctrl *ActivityController) GetActivities(c *fiber.Ctx) error {
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

	activities, err := ctrl.activityService.GetActivities(accountID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch activities",
		})
	}

	return c.JSON(fiber.Map{
		"error":      false,
		"activities": activities,
	})
}

// DeleteActivity deletes an activity
// @Summary Delete activity
// @Tags Activities
// @Produce json
// @Param accountId path string true "Account ID"
// @Param id path string true "Activity ID"
// @Success 200 {string} string "Activity deleted successfully"
// @Router /api/v1/accounts/{accountId}/activities/{id} [delete]
func (ctrl *ActivityController) DeleteActivity(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	accountIDStr := c.Params("accountId")
	if accountIDStr == "" {
		accountIDStr = c.Params("id")
	}
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid account ID format",
		})
	}

	activityIDStr := c.Params("activityId")
	if activityIDStr == "" {
		activityIDStr = c.Params("id")
	}
	activityID, err := uuid.Parse(activityIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid activity ID format",
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

	err = ctrl.activityService.DeleteActivity(activityID, accountID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Activity deleted successfully",
	})
}
