package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"keepnode/pkg/apiresponse"

	usermodel "keepnode/app/models/user"
	userservice "keepnode/app/services/user"
)

type UserController struct {
	userService *userservice.UserService
}

func NewUserController(userService *userservice.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (ctrl *UserController) GetCurrentUser(c *fiber.Ctx, userID uuid.UUID) error {
	user, err := ctrl.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return apiresponse.RespondOK(c, user)
}

func (ctrl *UserController) UpdateProfile(c *fiber.Ctx, userID uuid.UUID) error {
	req := &usermodel.UpdateProfileRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := ctrl.userService.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, user)
}

func (ctrl *UserController) UpdateUserBaseCurrency(c *fiber.Ctx, userID uuid.UUID) error {

	req := &usermodel.UpdateBaseCurrencyRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := ctrl.userService.UpdateUserBaseCurrency(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, user)
}

func (ctrl *UserController) UpdateUserTheme(c *fiber.Ctx, userID uuid.UUID) error {
	req := &usermodel.UpdateThemeRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := ctrl.userService.UpdateUserTheme(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, user)
}

func (ctrl *UserController) UpdateUserLocale(c *fiber.Ctx, userID uuid.UUID) error {
	req := &usermodel.UpdateLocaleRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := ctrl.userService.UpdateUserLocale(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, user)
}
