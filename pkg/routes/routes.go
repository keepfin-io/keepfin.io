package routes

import (
	"keepnode/app/controllers"
	"keepnode/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	// Public routes
	app.Post("/api/v1/user/sign/up", controllers.UserSignUp)
	app.Post("/api/v1/user/sign/in", controllers.UserSignIn)

	// Protected routes example
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())
	// api.Get("/user/profile", controllers.UserProfile) // örnek korumalı route
}
