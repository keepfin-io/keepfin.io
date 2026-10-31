package routes

import (
	"os"

	"github.com/gofiber/fiber/v2"

	accountcontroller "keepnode/app/controllers/account"
	activitycontroller "keepnode/app/controllers/activity"
	holdingcontroller "keepnode/app/controllers/holding"
	marketcontroller "keepnode/app/controllers/market"
	portfoliocontroller "keepnode/app/controllers/portfolio"
	usercontroller "keepnode/app/controllers/user"
	accountservice "keepnode/app/services/account"
	activityservice "keepnode/app/services/activity"
	holdingservice "keepnode/app/services/holding"
	marketservice "keepnode/app/services/market"
	portfolioservice "keepnode/app/services/portfolio"
	userservice "keepnode/app/services/user"
	"keepnode/pkg/authsession"
	"keepnode/pkg/middleware"
	"keepnode/platform/database"
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
	userService := userservice.NewUserService(db)
	accountService := accountservice.NewAccountService(db)
	marketService := marketservice.NewMarketService()
	holdingService := holdingservice.NewHoldingService(db, marketService)
	activityService := activityservice.NewActivityService(db, marketService)

	// Initialize controllers
	userController := usercontroller.NewUserController(userService)
	accountController := accountcontroller.NewAccountController(accountService)
	marketController := marketcontroller.NewMarketController(marketService)
	holdingController := holdingcontroller.NewHoldingController(holdingService, accountService)
	activityController := activitycontroller.NewActivityController(activityService, accountService)

	portfolioController := portfoliocontroller.NewPortfolioController(portfolioservice.NewPortfolioService(db, marketService), activityService, accountService)

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

	api.Get("/portfolio", portfolioController.Report)
	api.Post("/transfers", portfolioController.Transfer)
	api.Post("/accounts/:id/activities/import", portfolioController.Import)
	api.Get("/accounts/:id/activities/export", portfolioController.Export)
	api.Put("/accounts/:id/activities/:activityId", portfolioController.UpdateActivity)
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
