package wealth

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	model "keepnode/app/models/wealth"
	service "keepnode/app/services/wealth"
	"keepnode/pkg/utils"
)

type Controller struct{ Service *service.Service }

func New(s *service.Service) *Controller { return &Controller{s} }
func result(c *fiber.Ctx, key string, value any, err error) error {
	if err != nil {
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(404, "Record not found")
		}
		var dbErr *pgconn.PgError
		if errors.As(err, &dbErr) {
			if dbErr.Code == "23505" {
				return fiber.NewError(409, "This record already exists")
			}
			return fiber.NewError(500, "Unable to save or load records")
		}
		return fiber.NewError(422, err.Error())
	}
	return c.JSON(fiber.Map{"error": false, key: value})
}
func identity(c *fiber.Ctx) (uuid.UUID, error) { return utils.GetUserID(c) }
func recordID(c *fiber.Ctx) (uuid.UUID, error) {
	id, e := uuid.Parse(c.Params("id"))
	if e != nil {
		return id, fiber.NewError(400, "Invalid record ID")
	}
	return id, nil
}
func parse(c *fiber.Ctx, body any) error {
	if c.BodyParser(body) != nil {
		return fiber.NewError(400, "Invalid request body")
	}
	return nil
}
func optionalID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, e := uuid.Parse(value)
	if e != nil {
		return nil, fiber.NewError(400, "Invalid filter ID")
	}
	return &id, nil
}

func (ctrl *Controller) Register(api fiber.Router) {
	api.Get("/market/providers", ctrl.Providers)
	api.Put("/market/providers", ctrl.SetProvider)
	api.Get("/market/jobs", ctrl.Jobs)
	api.Post("/market/jobs", ctrl.QueueJob)
	api.Post("/market/jobs/:id/retry", ctrl.RetryJob)
	api.Get("/market/corrections", ctrl.Corrections)
	api.Put("/market/corrections", ctrl.CorrectPrice)
	api.Delete("/market/corrections", ctrl.DeleteCorrection)
	api.Get("/accounts/:id/dividends/preview", ctrl.Dividends)
	api.Get("/watchlist", ctrl.Watchlist)
	api.Post("/watchlist", ctrl.SaveWatch)
	api.Delete("/watchlist/:id", ctrl.DeleteWatch)
	api.Get("/assets/manual", ctrl.Assets)
	api.Post("/assets/manual", ctrl.SaveAsset)
	api.Put("/assets/manual/:id", ctrl.SaveAsset)
	api.Delete("/assets/manual/:id", ctrl.DeleteAsset)
	api.Put("/assets/manual/:id/valuations", ctrl.Valuation)
	api.Get("/liabilities", ctrl.Liabilities)
	api.Post("/liabilities", ctrl.CreateLiability)
	api.Put("/liabilities/:id/balances", ctrl.LiabilityBalance)
	api.Get("/tags", ctrl.Tags)
	api.Post("/tags", ctrl.SaveTag)
	api.Put("/tags/:id", ctrl.SaveTag)
	api.Delete("/tags/:id", ctrl.DeleteTag)
	api.Get("/tags/:kind/:id", ctrl.EntityTags)
	api.Put("/tags/:kind/:id", ctrl.AssignTags)
	api.Get("/activities", ctrl.Activities)
	api.Get("/preferences", ctrl.Preferences)
	api.Put("/preferences", ctrl.SavePreferences)
	api.Get("/risk", ctrl.Risk)
	api.Get("/fire", ctrl.Fire)
	api.Get("/shares", ctrl.Shares)
	api.Post("/shares", ctrl.CreateShare)
	api.Delete("/shares/:id", ctrl.RevokeShare)
	api.Post("/shared/report", ctrl.SharedReport)
}
func (ctrl *Controller) Watchlist(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Watchlist(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) SaveWatch(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var body model.WatchItem
	if e = parse(c, &body); e != nil {
		return e
	}
	v, e := ctrl.Service.AddWatch(u, body)
	return result(c, "item", v, e)
}
func (ctrl *Controller) DeleteWatch(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, "deleted", true, ctrl.Service.DeleteWatch(u, id))
}
func (ctrl *Controller) Assets(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Assets(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) SaveAsset(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := optionalID(c.Params("id"))
	if e != nil {
		return e
	}
	var body model.ManualAsset
	if e = parse(c, &body); e != nil {
		return e
	}
	v, e := ctrl.Service.SaveAsset(u, id, body)
	return result(c, "item", v, e)
}
func (ctrl *Controller) DeleteAsset(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, "deleted", true, ctrl.Service.DeleteAsset(u, id))
}
func (ctrl *Controller) Valuation(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	var b struct {
		Date  string   `json:"date"`
		Price *float64 `json:"price"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	if b.Price == nil {
		return fiber.NewError(400, "Price is required")
	}
	return result(c, "saved", true, ctrl.Service.SaveValuation(u, id, b.Date, *b.Price))
}
func (ctrl *Controller) Liabilities(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Liabilities(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) CreateLiability(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var b struct {
		Name     string   `json:"name"`
		Currency string   `json:"currency"`
		Date     string   `json:"date"`
		Amount   *float64 `json:"amount"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	if b.Amount == nil {
		return fiber.NewError(400, "Amount is required")
	}
	v, e := ctrl.Service.CreateLiability(u, b.Name, b.Currency, b.Date, *b.Amount)
	return result(c, "item", v, e)
}
func (ctrl *Controller) LiabilityBalance(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	var b struct {
		Date   string   `json:"date"`
		Amount *float64 `json:"amount"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	if b.Amount == nil {
		return fiber.NewError(400, "Amount is required")
	}
	return result(c, "saved", true, ctrl.Service.SaveLiabilityBalance(u, id, b.Date, *b.Amount))
}
func (ctrl *Controller) Tags(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Tags(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) SaveTag(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := optionalID(c.Params("id"))
	if e != nil {
		return e
	}
	var b struct {
		Name string `json:"name"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	v, e := ctrl.Service.SaveTag(u, id, b.Name)
	return result(c, "item", v, e)
}
func (ctrl *Controller) DeleteTag(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, "deleted", true, ctrl.Service.DeleteTag(u, id))
}
func (ctrl *Controller) EntityTags(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.EntityTags(u, id, c.Params("kind"))
	return result(c, "items", v, e)
}
func (ctrl *Controller) AssignTags(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	var b struct {
		IDs []uuid.UUID `json:"tag_ids"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	return result(c, "saved", true, ctrl.Service.AssignTags(u, id, c.Params("kind"), b.IDs))
}
func (ctrl *Controller) Activities(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	account, e := optionalID(c.Query("account_id"))
	if e != nil {
		return e
	}
	tag, e := optionalID(c.Query("tag_id"))
	if e != nil {
		return e
	}
	v, e := ctrl.Service.AllActivities(u, service.ActivityFilter{AccountID: account, TagID: tag, Query: c.Query("q"), Type: c.Query("type"), From: c.Query("from"), To: c.Query("to"), Direction: c.Query("direction"), Page: c.QueryInt("page", 1), PerPage: c.QueryInt("per_page", 25)})
	return result(c, "result", v, e)
}
func (ctrl *Controller) Preferences(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Preferences(u)
	return result(c, "preferences", v, e)
}
func (ctrl *Controller) SavePreferences(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var b model.Preferences
	if e = parse(c, &b); e != nil {
		return e
	}
	v, e := ctrl.Service.SavePreferences(u, b)
	return result(c, "preferences", v, e)
}
func (ctrl *Controller) Risk(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	p, e := ctrl.Service.Preferences(u)
	if e != nil {
		return result(c, "report", nil, e)
	}
	r, e := ctrl.Service.Portfolio.Report(u, nil, c.Query("range", "1y"), "")
	if e != nil {
		return result(c, "report", nil, e)
	}
	return result(c, "report", service.RiskChecks(r, p), nil)
}
func (ctrl *Controller) Fire(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	p, e := ctrl.Service.Preferences(u)
	if e != nil {
		return result(c, "report", nil, e)
	}
	r, e := ctrl.Service.Portfolio.Report(u, nil, "1d", "")
	if e != nil {
		return result(c, "report", nil, e)
	}
	return result(c, "report", service.FireProjection(r, p), nil)
}
func (ctrl *Controller) Shares(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Shares(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) CreateShare(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var b service.ShareRequest
	if e = parse(c, &b); e != nil {
		return e
	}
	v, token, e := ctrl.Service.CreateShare(u, b)
	c.Set("Cache-Control", "no-store")
	return result(c, "share", fiber.Map{"record": v, "token": token}, e)
}
func (ctrl *Controller) RevokeShare(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, "revoked", true, ctrl.Service.RevokeShare(u, id))
}
func (ctrl *Controller) SharedReport(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	return ctrl.shared(c, &u)
}
func (ctrl *Controller) PublicReport(c *fiber.Ctx) error { return ctrl.shared(c, nil) }
func (ctrl *Controller) shared(c *fiber.Ctx, viewer *uuid.UUID) error {
	var b struct {
		Token string `json:"token"`
		Range string `json:"range"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	if b.Range == "" {
		b.Range = "1m"
	}
	c.Set("Cache-Control", "no-store")
	c.Set("Referrer-Policy", "no-referrer")
	v, e := ctrl.Service.SharedReport(b.Token, viewer, b.Range)
	return result(c, "portfolio", v, e)
}
