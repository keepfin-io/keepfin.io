package activity

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"keepnode/pkg/apiresponse"

	activitymodel "keepnode/app/models/activity"
	accountservice "keepnode/app/services/account"
	activityservice "keepnode/app/services/activity"
)

type ActivityController struct {
	activityService *activityservice.ActivityService
	accountService  *accountservice.AccountService
}

func NewActivityController(activityService *activityservice.ActivityService, accountService *accountservice.AccountService) *ActivityController {
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
func (ctrl *ActivityController) CreateActivity(c *fiber.Ctx, userID uuid.UUID) error {
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	req := &activitymodel.CreateActivityRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	activity, err := ctrl.activityService.CreateActivity(accountID, nil, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondCreated(c, activity)
}

// GetActivities returns all activities for an account
// @Summary Get account activities
// @Tags Activities
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {array} models.Activity
// @Router /api/v1/accounts/{id}/activities [get]
func (ctrl *ActivityController) GetActivities(c *fiber.Ctx, userID uuid.UUID) error {
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	activities, err := ctrl.activityService.GetActivities(accountID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch activities")
	}

	return apiresponse.RespondOK(c, activities)
}

// DeleteActivity deletes an activity
// @Summary Delete activity
// @Tags Activities
// @Produce json
// @Param accountId path string true "Account ID"
// @Param id path string true "Activity ID"
// @Success 200 {string} string "Activity deleted successfully"
// @Router /api/v1/accounts/{accountId}/activities/{id} [delete]
func (ctrl *ActivityController) DeleteActivity(c *fiber.Ctx, userID uuid.UUID) error {
	accountIDStr := c.Params("accountId")
	if accountIDStr == "" {
		accountIDStr = c.Params("id")
	}
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	activityIDStr := c.Params("activityId")
	if activityIDStr == "" {
		activityIDStr = c.Params("id")
	}
	activityID, err := uuid.Parse(activityIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid activity ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	err = ctrl.activityService.DeleteActivity(activityID, accountID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, fiber.Map{"message": "Activity deleted successfully"})
}
