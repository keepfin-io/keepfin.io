package portfolio

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	activitymodel "keepnode/app/models/activity"
	accountservice "keepnode/app/services/account"
	activityservice "keepnode/app/services/activity"
	portfolioservice "keepnode/app/services/portfolio"
	"keepnode/pkg/utils"
)

type PortfolioController struct {
	portfolio  *portfolioservice.PortfolioService
	activities *activityservice.ActivityService
	accounts   *accountservice.AccountService
}

func NewPortfolioController(p *portfolioservice.PortfolioService, a *activityservice.ActivityService, accounts *accountservice.AccountService) *PortfolioController {
	return &PortfolioController{p, a, accounts}
}
func (ctrl *PortfolioController) account(c *fiber.Ctx) (uuid.UUID, error) {
	user, err := utils.GetUserID(c)
	if err != nil {
		return uuid.Nil, fiber.NewError(401, "Authentication required")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(400, "Invalid account ID")
	}
	if _, err = ctrl.accounts.GetAccountByID(id, user); err != nil {
		return uuid.Nil, fiber.NewError(404, "Account not found")
	}
	return id, nil
}
func (ctrl *PortfolioController) Report(c *fiber.Ctx) error {
	user, err := utils.GetUserID(c)
	if err != nil {
		return fiber.NewError(401, "Authentication required")
	}
	var account *uuid.UUID
	if c.Query("account_id") != "" {
		id, e := uuid.Parse(c.Query("account_id"))
		if e != nil {
			return fiber.NewError(400, "Invalid account ID")
		}
		if _, e = ctrl.accounts.GetAccountByID(id, user); e != nil {
			return fiber.NewError(404, "Account not found")
		}
		account = &id
	}
	report, err := ctrl.portfolio.Report(user, account, c.Query("range", "1m"), c.Query("benchmark"))
	if err != nil {
		return fiber.NewError(422, err.Error())
	}
	return c.JSON(fiber.Map{"error": false, "portfolio": report})
}
func (ctrl *PortfolioController) UpdateActivity(c *fiber.Ctx) error {
	account, err := ctrl.account(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("activityId"))
	if err != nil {
		return fiber.NewError(400, "Invalid activity ID")
	}
	var req activitymodel.CreateActivityRequest
	if err = c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid activity")
	}
	a, err := ctrl.activities.UpdateActivity(id, account, &req)
	if err != nil {
		return fiber.NewError(422, err.Error())
	}
	return c.JSON(fiber.Map{"error": false, "activity": a})
}
func (ctrl *PortfolioController) Transfer(c *fiber.Ctx) error {
	user, err := utils.GetUserID(c)
	if err != nil {
		return fiber.NewError(401, "Authentication required")
	}
	var req activityservice.TransferRequest
	if err = c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid transfer")
	}
	if err = ctrl.activities.Transfer(user, req); err != nil {
		return fiber.NewError(422, err.Error())
	}
	return c.JSON(fiber.Map{"error": false})
}
func (ctrl *PortfolioController) Import(c *fiber.Ctx) error {
	account, err := ctrl.account(c)
	if err != nil {
		return err
	}
	var req struct {
		CSV    string `json:"csv"`
		Commit bool   `json:"commit"`
	}
	if err = c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid CSV request")
	}
	result, err := ctrl.activities.ImportCSV(account, req.CSV, req.Commit)
	if err != nil {
		return fiber.NewError(422, err.Error())
	}
	return c.JSON(fiber.Map{"error": false, "result": result})
}
func (ctrl *PortfolioController) Export(c *fiber.Ctx) error {
	account, err := ctrl.account(c)
	if err != nil {
		return err
	}
	data, err := ctrl.activities.ExportCSV(account)
	if err != nil {
		return fiber.NewError(500, "Unable to export activities")
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=keepfin-activities.csv")
	return c.Send(data)
}
