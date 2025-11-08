package controllers

import (
	"keepnode/app/models"
	"keepnode/app/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Call service layer
	user, err := ctrl.userService.SignUp(signUp)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "User created successfully",
		"user":    user,
	})
}

// UserSignIn method to auth user and set HTTP-only cookie.
// @Description Auth user and set HTTP-only cookie.
// @Summary auth user and set HTTP-only cookie
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
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Call service layer
	user, token, err := ctrl.userService.SignIn(signIn)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Set HTTP-only cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
		MaxAge:   24 * 60 * 60, // 24 hours
		Path:     "/",
	})

	// Return success response (without token in body)
	return c.JSON(fiber.Map{
		"error":   false,
		"message": "User signed in successfully",
		"user":    user,
	})
}

// UserSignOut method to clear authentication cookie.
// @Description Clear authentication cookie.
// @Summary logout user
// @Tags User
// @Success 200 {string} status "ok"
// @Router /api/v1/user/signout [post]
func (ctrl *AuthController) UserSignOut(c *fiber.Ctx) error {
	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   -1, // Delete cookie
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "User signed out successfully",
	})
}

// GetCurrentUser method to get current authenticated user.
// @Description Get current authenticated user from cookie.
// @Summary get current user
// @Tags User
// @Success 200 {object} models.User
// @Router /api/v1/user/me [get]
func (ctrl *AuthController) GetCurrentUser(c *fiber.Ctx) error {
	// Get user info from context (set by JWT middleware)
	userLocal := c.Locals("user")
	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User not authenticated",
		})
	}

	// Parse token to get claims
	token, ok := userLocal.(*jwt.Token)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid token",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid token claims",
		})
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid user ID in token",
		})
	}

	// Parse UUID from string
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid user ID format",
		})
	}

	// Get user from database
	user, err := ctrl.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"user":  user,
	})
}
