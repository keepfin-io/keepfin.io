package routes

import (
	"os"

	"keepnode/app/controllers"
	"keepnode/app/services"
	"keepnode/pkg/authsession"
	"keepnode/pkg/middleware"
	"keepnode/platform/database"

	"github.com/gofiber/fiber/v2"
)

// Register registers all routes
func Register(app *fiber.App) error {
	client, err := authsession.NewClient(os.Getenv("AUTH_SERVICE_URL"))
	if err != nil {
		return err
	}
	originCheck, err := middleware.SameOrigin(os.Getenv("APP_ORIGIN"))
	if err != nil {
		return err
	}
	// Initialize services
	db := database.GetDB()
	userService := services.NewUserService(db)
	accountService := services.NewAccountService(db)
	marketService := services.NewMarketService()
	holdingService := services.NewHoldingService(db, marketService)
	activityService := services.NewActivityService(db, marketService)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	accountController := controllers.NewAccountController(accountService)
	marketController := controllers.NewMarketController(marketService)
	holdingController := controllers.NewHoldingController(holdingService, accountService)
	activityController := controllers.NewActivityController(activityService, accountService)

	// All application routes use the server-verified Better Auth session.
	api := app.Group("/api/v1", originCheck, middleware.Session(client, userService))

	// User routes
	api.Get("/user/me", userController.GetCurrentUser)
	api.Put("/user/profile", userController.UpdateProfile)
	api.Put("/user/base-currency", userController.UpdateUserBaseCurrency)
	api.Put("/user/theme", userController.UpdateUserTheme)
	api.Put("/user/locale", userController.UpdateUserLocale)

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

	// Activity routes
	api.Post("/accounts/:id/activities", activityController.CreateActivity)
	api.Get("/accounts/:id/activities", activityController.GetActivities)
	api.Delete("/accounts/:accountId/activities/:activityId", activityController.DeleteActivity)
	return nil
}
