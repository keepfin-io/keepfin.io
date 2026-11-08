package routes

import (
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	wealthcontroller "keepnode/app/controllers/wealth"
	wealthservice "keepnode/app/services/wealth"

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
	"keepnode/pkg/authhandler"
	"keepnode/pkg/authsession"
	"keepnode/pkg/configs"
	"keepnode/pkg/middleware"
	"keepnode/pkg/typesense"
	"keepnode/pkg/worker"
	"keepnode/platform/database"
)

// Register registers all routes
func Register(app *fiber.App) error {
	ipResolver, err := middleware.NewClientIPResolver(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if err != nil {
		return err
	}
	ipLimit, err := configs.RateLimit("API_IP_RATE_LIMIT", 600)
	if err != nil {
		return err
	}
	userLimit, err := configs.RateLimit("API_USER_RATE_LIMIT", 300)
	if err != nil {
		return err
	}
	app.Use("/api", ipResolver.Middleware, middleware.IPRequestLimit(ipLimit))
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
	app.Post(
		"/api/public/portfolio",
		originCheck,
		middleware.IPRequestLimit(20),
		wealthController.PublicReport,
	)

	// All application routes use the server-verified Better Auth session.
	api := app.Group("/api/v1", originCheck, middleware.Session(client, userService), middleware.UserRequestLimit(userLimit), middleware.HoldingSearchSync(searchService))
	workerContext, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	go func() { defer close(workerDone); searchService.Run(workerContext) }()
	marketDone := make(chan struct{})
	go func() { defer close(marketDone); wealthService.RunMarketWorker(workerContext) }()
	notificationsDone := make(chan struct{})
	go func() {
		defer close(notificationsDone)
		worker.Listen(workerContext, configs.GetDatabaseDSN(), map[string]func(){
			worker.HoldingsChanged:     searchService.Wake,
			worker.MarketJobsChanged:   wealthService.WakeMarketWorker,
			worker.MarketConfigChanged: wealthService.InvalidateProvider,
		})
	}()
	app.Hooks().OnShutdown(func() error {
		stopWorker()
		<-notificationsDone
		<-workerDone
		<-marketDone
		return nil
	})

	wealthController.Register(api)

	// User routes
	api.Get("/user/me", authhandler.WithUser(userController.GetCurrentUser))
	api.Put("/user/profile", authhandler.WithUser(userController.UpdateProfile))
	api.Put("/user/base-currency", authhandler.WithUser(userController.UpdateUserBaseCurrency))
	api.Put("/user/theme", authhandler.WithUser(userController.UpdateUserTheme))
	api.Put("/user/locale", authhandler.WithUser(userController.UpdateUserLocale))

	// Market data routes
	api.Get("/market/eod/:symbol", authhandler.WithUser(marketController.GetEODData))
	api.Get("/market/live", authhandler.WithUser(marketController.GetMultipleLivePrices))
	api.Get("/market/live/:symbol", authhandler.WithUser(marketController.GetLivePrice))
	api.Get("/market/search", authhandler.WithUser(marketController.SearchAssets))

	api.Get("/portfolio", authhandler.WithUser(portfolioController.GetPortfolioReport))
	api.Post("/transfers", authhandler.WithUser(portfolioController.CreateAccountTransfer))
	api.Post("/accounts/:id/activities/import", authhandler.WithUser(portfolioController.ImportAccountActivities))
	api.Get("/accounts/:id/activities/export", authhandler.WithUser(portfolioController.ExportAccountActivities))
	api.Put("/accounts/:id/activities/:activityId", authhandler.WithUser(portfolioController.UpdateAccountActivity))
	// Account routes
	api.Post("/accounts/create", authhandler.WithUser(accountController.CreateAccount))
	api.Get("/accounts", authhandler.WithUser(accountController.GetUserAccounts))
	api.Get("/accounts/:id", authhandler.WithUser(accountController.GetAccount))
	api.Put("/accounts/:id", authhandler.WithUser(accountController.UpdateAccount))
	api.Delete("/accounts/:id", authhandler.WithUser(accountController.DeleteAccount))
	api.Patch("/accounts/:id/toggle-status", authhandler.WithUser(accountController.ToggleAccountStatus))

	// Holding routes
	api.Post("/accounts/:id/holdings", authhandler.WithUser(holdingController.AddHolding))
	api.Get("/accounts/:id/holdings", authhandler.WithUser(holdingController.GetHoldings))
	api.Get("/accounts/:id/holdings/search", authhandler.WithUser(holdingController.SearchHoldings))
	api.Put("/accounts/:accountId/holdings/:id", authhandler.WithUser(holdingController.UpdateHolding))
	api.Delete("/accounts/:accountId/holdings/:id", authhandler.WithUser(holdingController.DeleteHolding))

	// Activity routes
	api.Post("/accounts/:id/activities", authhandler.WithUser(activityController.CreateActivity))
	api.Get("/accounts/:id/activities", authhandler.WithUser(activityController.GetActivities))
	api.Delete("/accounts/:accountId/activities/:activityId", authhandler.WithUser(activityController.DeleteActivity))
	return nil
}
