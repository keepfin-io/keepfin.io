package holding

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"keepnode/pkg/apiresponse"

	holdingmodel "keepnode/app/models/holding"
	accountservice "keepnode/app/services/account"
	holdingservice "keepnode/app/services/holding"
)

type HoldingController struct {
	holdingService *holdingservice.HoldingService
	accountService *accountservice.AccountService
	searchService  *holdingservice.SearchService
}

func NewHoldingController(
	holdingService *holdingservice.HoldingService,
	accountService *accountservice.AccountService,
	searchService *holdingservice.SearchService,
) *HoldingController {
	return &HoldingController{
		holdingService: holdingService,
		accountService: accountService,
		searchService:  searchService,
	}
}

// SearchHoldings returns one account-scoped page. Ownership is checked in the
// database before either Typesense or the outage fallback is consulted.
func (ctrl *HoldingController) SearchHoldings(c *fiber.Ctx, userID uuid.UUID) error {
	accountID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "Invalid account ID")
	}
	if _, err = ctrl.accountService.GetAccountByID(accountID, userID); err != nil {
		return fiber.NewError(404, "Account not found")
	}
	params := holdingmodel.SearchParams{
		Query: c.Query("q"), Currency: c.Query("currency"), Exchange: c.Query("exchange"),
		AssetClass: c.Query("asset_class"), Sector: c.Query("sector"), Country: c.Query("country"),
		Sort: c.Query("sort", "symbol"), Direction: c.Query("direction", "asc"),
		Page: c.QueryInt("page", 1), PerPage: c.QueryInt("per_page", 10),
	}
	if err = params.Validate(); err != nil {
		return fiber.NewError(400, err.Error())
	}
	page, err := ctrl.searchService.Search(c.UserContext(), userID, accountID, params)
	if err != nil {
		return fiber.NewError(503, "Holdings are temporarily unavailable")
	}
	rows := make([]holdingmodel.Holding, 0, len(page.Documents))
	for _, doc := range page.Documents {
		id, e := uuid.Parse(doc.ID)
		if e != nil {
			return fiber.NewError(503, "Invalid holding index")
		}
		rows = append(rows, holdingmodel.Holding{
			ID: id, AccountID: accountID, Symbol: doc.Symbol, Name: doc.Name,
			Quantity: doc.Quantity, BuyPrice: doc.BuyPrice, BuyCurrency: doc.Currency,
			AssetClass: doc.AssetClass, Sector: doc.Sector, Country: doc.Country,
		})
	}
	if len(rows) > 0 {
		page.Holdings, _, err = ctrl.holdingService.ValueHoldings(accountID, rows)
		if err != nil {
			return fiber.NewError(503, "Unable to value holdings")
		}
	}
	c.Set("Cache-Control", "no-store")
	return apiresponse.RespondOK(
		c,
		fiber.Map{"holdings": page.Holdings, "facets": page.Facets, "source": page.Source},
		apiresponse.WithPagination(page.Page, page.PerPage, int64(page.Total)),
	)
}

// AddHolding adds a new holding to an account
// @Summary Add holding to account
// @Tags Holdings
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param holding body models.CreateHoldingRequest true "Holding data"
// @Success 201 {object} models.Holding
// @Router /api/v1/accounts/{id}/holdings [post]
func (ctrl *HoldingController) AddHolding(c *fiber.Ctx, userID uuid.UUID) error {
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	req := &holdingmodel.CreateHoldingRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	holding, err := ctrl.holdingService.AddHolding(accountID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondCreated(c, holding)
}

// GetHoldings returns all holdings for an account with live data
// @Summary Get account holdings with live data
// @Tags Holdings
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {array} models.HoldingWithLiveData
// @Router /api/v1/accounts/{id}/holdings [get]
func (ctrl *HoldingController) GetHoldings(c *fiber.Ctx, userID uuid.UUID) error {
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	holdings, summary, err := ctrl.holdingService.GetHoldingsWithLiveData(accountID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch holdings")
	}

	return apiresponse.RespondOK(c, fiber.Map{"holdings": holdings, "summary": summary})
}

// UpdateHolding updates a holding
// @Summary Update holding
// @Tags Holdings
// @Accept json
// @Produce json
// @Param id path string true "Holding ID"
// @Param holding body models.UpdateHoldingRequest true "Update data"
// @Success 200 {object} models.Holding
// @Router /api/v1/holdings/{id} [put]
func (ctrl *HoldingController) UpdateHolding(c *fiber.Ctx, userID uuid.UUID) error {
	holdingIDStr := c.Params("id")
	holdingID, err := uuid.Parse(holdingIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid holding ID format")
	}

	accountIDStr := c.Params("accountId")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	req := &holdingmodel.UpdateHoldingRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	holding, err := ctrl.holdingService.UpdateHolding(holdingID, accountID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, holding)
}

// DeleteHolding deletes a holding
// @Summary Delete holding
// @Tags Holdings
// @Produce json
// @Param id path string true "Holding ID"
// @Success 200 {string} string "Holding deleted successfully"
// @Router /api/v1/holdings/{id} [delete]
func (ctrl *HoldingController) DeleteHolding(c *fiber.Ctx, userID uuid.UUID) error {
	holdingIDStr := c.Params("id")
	holdingID, err := uuid.Parse(holdingIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid holding ID format")
	}

	accountIDStr := c.Params("accountId")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid account ID format")
	}

	// Verify account belongs to user
	_, err = ctrl.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Account not found")
	}

	err = ctrl.holdingService.DeleteHolding(holdingID, accountID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return apiresponse.RespondOK(c, fiber.Map{"message": "Holding deleted successfully"})
}
