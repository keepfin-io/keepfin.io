package portfolio

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"keepnode/pkg/apiresponse"

	activitymodel "keepnode/app/models/activity"
	accountservice "keepnode/app/services/account"
	activityservice "keepnode/app/services/activity"
	portfolioservice "keepnode/app/services/portfolio"
)

type PortfolioController struct {
	portfolio  *portfolioservice.PortfolioService
	activities *activityservice.ActivityService
	accounts   *accountservice.AccountService
}

func NewPortfolioController(p *portfolioservice.PortfolioService, a *activityservice.ActivityService, accounts *accountservice.AccountService) *PortfolioController {
	return &PortfolioController{p, a, accounts}
}

func (ctrl *PortfolioController) requireOwnedAccountID(c *fiber.Ctx, userID uuid.UUID) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(400, "Invalid account ID")
	}
	if _, err = ctrl.accounts.GetAccountByID(id, userID); err != nil {
		return uuid.Nil, fiber.NewError(404, "Account not found")
	}
	return id, nil
}

func (ctrl *PortfolioController) GetPortfolioReport(c *fiber.Ctx, userID uuid.UUID) error {
	var account *uuid.UUID
	if c.Query("account_id") != "" {
		id, e := uuid.Parse(c.Query("account_id"))
		if e != nil {
			return fiber.NewError(400, "Invalid account ID")
		}
		if _, e = ctrl.accounts.GetAccountByID(id, userID); e != nil {
			return fiber.NewError(404, "Account not found")
		}
		account = &id
	}
	opts := portfolioservice.ReportOptions{Method: c.Query("method")}
	if c.Query("tag_id") != "" {
		id, e := uuid.Parse(c.Query("tag_id"))
		if e != nil {
			return fiber.NewError(400, "Invalid tag ID")
		}
		opts.TagID = &id
	}
	report, err := ctrl.portfolio.Report(
		userID,
		account,
		c.Query("range", "1m"),
		c.Query("benchmark"),
		opts,
	)
	if err != nil {
		return fiber.NewError(422, err.Error())
	}
	return apiresponse.RespondOK(c, report)
}

func (ctrl *PortfolioController) UpdateAccountActivity(c *fiber.Ctx, userID uuid.UUID) error {
	account, err := ctrl.requireOwnedAccountID(c, userID)
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
	return apiresponse.RespondOK(c, a)
}

func (ctrl *PortfolioController) CreateAccountTransfer(c *fiber.Ctx, userID uuid.UUID) error {
	var req activityservice.TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid transfer")
	}
	if err := ctrl.activities.Transfer(userID, req); err != nil {
		return fiber.NewError(422, err.Error())
	}
	return apiresponse.RespondOK[any](c, nil)
}

func (ctrl *PortfolioController) ImportAccountActivities(c *fiber.Ctx, userID uuid.UUID) error {
	account, err := ctrl.requireOwnedAccountID(c, userID)
	if err != nil {
		return err
	}
	var req struct {
		CSV    string `json:"csv"`
		JSON   string `json:"json"`
		Commit bool   `json:"commit"`
	}
	if err = c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid CSV request")
	}
	var result *activityservice.ImportResult
	if req.JSON != "" {
		result, err = ctrl.activities.ImportJSON(account, req.JSON, req.Commit)
	} else {
		result, err = ctrl.activities.ImportCSV(account, req.CSV, req.Commit)
	}
	if err != nil {
		return fiber.NewError(422, err.Error())
	}
	return apiresponse.RespondOK(c, result)
}

func (ctrl *PortfolioController) ExportAccountActivities(c *fiber.Ctx, userID uuid.UUID) error {
	account, err := ctrl.requireOwnedAccountID(c, userID)
	if err != nil {
		return err
	}
	if c.Query("format") == "json" {
		data, e := ctrl.activities.ExportJSON(account)
		if e != nil {
			return fiber.NewError(500, "Unable to export activities")
		}
		c.Set("Content-Type", "application/json")
		c.Set("Content-Disposition", "attachment; filename=keepfin-activities.json")
		return c.Send(data)
	}
	data, err := ctrl.activities.ExportCSV(account)
	if err != nil {
		return fiber.NewError(500, "Unable to export activities")
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=keepfin-activities.csv")
	return c.Send(data)
}
