package controllers

import (
	"keepnode/app/models"
	"keepnode/app/services"

	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication-related requests
type AuthController struct {
	userService *services.UserService
}

// NewAuthController creates a new AuthController instance
func NewAuthController(userService *services.UserService) *AuthController {
	return &AuthController{
		userService: userService,
	}
}

// UserSignUp method to create a new user.
// @Description Create a new user.
// @Summary create a new user
// @Tags User
// @Accept json
// @Produce json
// @Param email body string true "Email"
// @Param password body string true "Password"
// @Success 200 {object} models.User
// @Router /api/v1/user/sign/up [post]
func (ctrl *AuthController) UserSignUp(c *fiber.Ctx) error {
	// Parse request body
	signUp := &models.UserSignUp{}
	if err := c.BodyParser(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body",
		})
	}

	// Call service layer
	user, err := ctrl.userService.SignUp(signUp)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   "User created successfully",
		"user":  user,
	})
}

// UserSignIn method to auth user and return access token.
// @Description Auth user and return access token.
// @Summary auth user and return access token
// @Tags User
// @Accept json
// @Produce json
// @Param email body string true "Email"
// @Param password body string true "Password"
// @Success 200 {string} status "ok"
// @Router /api/v1/user/sign/in [post]
func (ctrl *AuthController) UserSignIn(c *fiber.Ctx) error {
	// Parse request body
	signIn := &models.UserSignIn{}
	if err := c.BodyParser(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body",
		})
	}

	// Call service layer
	user, token, err := ctrl.userService.SignIn(signIn)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"error": false,
		"msg":   "User signed in successfully",
		"user":  user,
		"token": token,
	})
}
