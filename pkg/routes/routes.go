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
	app.Post("/api/v1/user/sign/up", authController.UserSignUp)
	app.Post("/api/v1/user/sign/in", authController.UserSignIn)

	// Protected routes example
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())
	// api.Get("/user/profile", authController.UserProfile) // örnek korumalı route
}
