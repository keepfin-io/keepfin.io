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
	accountService := services.NewAccountService(db)

	// Initialize controllers
	authController := controllers.NewAuthController(userService)
	accountController := controllers.NewAccountController(accountService)

	// Public routes
	app.Post("/api/v1/user/signup", authController.UserSignUp)
	app.Post("/api/v1/user/signin", authController.UserSignIn)

	// Protected routes
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())
	api.Get("/user/me", authController.GetCurrentUser)
	api.Post("/user/signout", authController.UserSignOut)

	// Account routes
	api.Post("/accounts/create", accountController.CreateAccount)
	api.Get("/accounts", accountController.GetUserAccounts)
	api.Get("/accounts/:id", accountController.GetAccount)
	api.Put("/accounts/:id", accountController.UpdateAccount)
	api.Delete("/accounts/:id", accountController.DeleteAccount)
	api.Patch("/accounts/:id/toggle-status", accountController.ToggleAccountStatus)

}
