package routes

import (
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	wealthcontroller "keepnode/app/controllers/wealth"
	wealthservice "keepnode/app/services/wealth"
	"time"

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
	"keepnode/pkg/typesense"
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
	searchClient, err := typesense.New(os.Getenv("TYPESENSE_URL"), os.Getenv("TYPESENSE_API_KEY"))
	if err != nil {
		return err
	}
	searchService := holdingservice.NewSearchService(db, searchClient, holdingservice.HoldingsCollection)
	activityService := activityservice.NewActivityService(db, marketService)

	// Initialize controllers
	userController := usercontroller.NewUserController(userService)
	accountController := accountcontroller.NewAccountController(accountService)
	marketController := marketcontroller.NewMarketController(marketService, db)
	holdingController := holdingcontroller.NewHoldingController(holdingService, accountService, searchService)
	activityController := activitycontroller.NewActivityController(activityService, accountService)

	portfolioService := portfolioservice.NewPortfolioService(db, marketService)
	portfolioController := portfoliocontroller.NewPortfolioController(portfolioService, activityService, accountService)
	wealthService := wealthservice.New(db, marketService, activityService, portfolioService)
	if err := wealthService.LoadProvider(); err != nil {
		return err
	}
	wealthController := wealthcontroller.New(wealthService)
	app.Post("/api/public/portfolio", originCheck, limiter.New(limiter.Config{Max: 20, Expiration: time.Minute}), wealthController.PublicReport)

	// All application routes use the server-verified Better Auth session.
	api := app.Group("/api/v1", originCheck, middleware.Session(client, userService), middleware.HoldingSearchSync(searchService))
	workerContext, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	go func() { defer close(workerDone); searchService.Run(workerContext) }()
	marketDone := make(chan struct{})
	go func() { defer close(marketDone); wealthService.RunMarketWorker(workerContext) }()
	app.Hooks().OnShutdown(func() error { stopWorker(); <-workerDone; <-marketDone; return nil })

	wealthController.Register(api)

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
	api.Get("/accounts/:id/holdings/search", holdingController.SearchHoldings)
	api.Put("/accounts/:accountId/holdings/:id", holdingController.UpdateHolding)
	api.Delete("/accounts/:accountId/holdings/:id", holdingController.DeleteHolding)

	// Activity routes
	api.Post("/accounts/:id/activities", activityController.CreateActivity)
	api.Get("/accounts/:id/activities", activityController.GetActivities)
	api.Delete("/accounts/:accountId/activities/:activityId", activityController.DeleteActivity)
	return nil
}
