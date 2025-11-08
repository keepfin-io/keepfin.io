package routes

import (
	"keepnode/app/controllers"
	"keepnode/app/services"
	"keepnode/pkg/middleware"
	"keepnode/platform/database"

	"github.com/gofiber/fiber/v2"
)

// Register registers all routes
func Register(app *fiber.App) {
	// Initialize services
	db := database.GetDB()
	userService := services.NewUserService(db)

	// Initialize controllers
	authController := controllers.NewAuthController(userService)

	// Public routes
	app.Post("/api/v1/user/signup", authController.UserSignUp)
	app.Post("/api/v1/user/signin", authController.UserSignIn)

	// Protected routes
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())
	api.Get("/user/me", authController.GetCurrentUser)
	api.Post("/user/signout", authController.UserSignOut)
}
