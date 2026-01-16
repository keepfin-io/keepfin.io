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
	marketService := services.NewMarketService()
	holdingService := services.NewHoldingService(db, marketService)

	// Initialize controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	accountController := controllers.NewAccountController(accountService)
	marketController := controllers.NewMarketController(marketService)
	holdingController := controllers.NewHoldingController(holdingService, accountService)

	// Public routes (no auth required)
	app.Post("/api/v1/user/signup", authController.UserSignUp)
	app.Post("/api/v1/user/signin", authController.UserSignIn)

	// Protected routes (JWT required)
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())

	// Auth routes
	api.Post("/user/signout", authController.UserSignOut)

	// User routes
	api.Get("/user/me", userController.GetCurrentUser)
	api.Put("/user/profile", userController.UpdateProfile)
	api.Put("/user/base-currency", userController.UpdateUserBaseCurrency)
	api.Put("/user/theme", userController.UpdateUserTheme)

	// Market data routes
	api.Get("/market/eod/:symbol", marketController.GetEODData)
	api.Get("/market/live", marketController.GetMultipleLivePrices)
	api.Get("/market/live/:symbol", marketController.GetLivePrice)
	api.Get("/market/search", marketController.SearchAssets)

	// Account routes
	api.Post("/accounts/create", accountController.CreateAccount)
	api.Get("/accounts", accountController.GetUserAccounts)
	api.Get("/accounts/:id", accountController.GetAccount)
	api.Put("/accounts/:id", accountController.UpdateAccount)
	api.Delete("/accounts/:id", accountController.DeleteAccount)
	api.Patch("/accounts/:id/toggle-status", accountController.ToggleAccountStatus)

	// Holding routes
	api.Post("/accounts/:id/holdings", holdingController.AddHolding)
	api.Get("/accounts/:id/holdings", holdingController.GetHoldings)
	api.Put("/accounts/:accountId/holdings/:id", holdingController.UpdateHolding)
	api.Delete("/accounts/:accountId/holdings/:id", holdingController.DeleteHolding)
}
