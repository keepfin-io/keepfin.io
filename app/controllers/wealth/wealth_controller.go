package wealth

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	model "keepnode/app/models/wealth"
	service "keepnode/app/services/wealth"
	"keepnode/pkg/apiresponse"
	"keepnode/pkg/authhandler"
)

type Controller struct{ Service *service.Service }

func New(s *service.Service) *Controller { return &Controller{s} }
func result(c *fiber.Ctx, value any, err error) error {
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
	return apiresponse.RespondOK(c, value)
}
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
	api.Get("/market/providers", authhandler.WithUser(ctrl.Providers))
	api.Put("/market/providers", authhandler.WithUser(ctrl.SetProvider))
	api.Get("/market/jobs", authhandler.WithUser(ctrl.Jobs))
	api.Post("/market/jobs", authhandler.WithUser(ctrl.QueueJob))
	api.Post("/market/jobs/:id/retry", authhandler.WithUser(ctrl.RetryJob))
	api.Get("/market/corrections", authhandler.WithUser(ctrl.Corrections))
	api.Put("/market/corrections", authhandler.WithUser(ctrl.CorrectPrice))
	api.Delete("/market/corrections", authhandler.WithUser(ctrl.DeleteCorrection))
	api.Get("/accounts/:id/dividends/preview", authhandler.WithUser(ctrl.Dividends))
	api.Get("/watchlist", authhandler.WithUser(ctrl.Watchlist))
	api.Post("/watchlist", authhandler.WithUser(ctrl.SaveWatch))
	api.Delete("/watchlist/:id", authhandler.WithUser(ctrl.DeleteWatch))
	api.Get("/assets/manual", authhandler.WithUser(ctrl.Assets))
	api.Post("/assets/manual", authhandler.WithUser(ctrl.SaveAsset))
	api.Put("/assets/manual/:id", authhandler.WithUser(ctrl.SaveAsset))
	api.Delete("/assets/manual/:id", authhandler.WithUser(ctrl.DeleteAsset))
	api.Put("/assets/manual/:id/valuations", authhandler.WithUser(ctrl.Valuation))
	api.Get("/liabilities", authhandler.WithUser(ctrl.Liabilities))
	api.Post("/liabilities", authhandler.WithUser(ctrl.CreateLiability))
	api.Put("/liabilities/:id/balances", authhandler.WithUser(ctrl.LiabilityBalance))
	api.Get("/tags", authhandler.WithUser(ctrl.Tags))
	api.Post("/tags", authhandler.WithUser(ctrl.SaveTag))
	api.Put("/tags/:id", authhandler.WithUser(ctrl.SaveTag))
	api.Delete("/tags/:id", authhandler.WithUser(ctrl.DeleteTag))
	api.Get("/tags/:kind/:id", authhandler.WithUser(ctrl.EntityTags))
	api.Put("/tags/:kind/:id", authhandler.WithUser(ctrl.AssignTags))
	api.Get("/activities", authhandler.WithUser(ctrl.Activities))
	api.Get("/preferences", authhandler.WithUser(ctrl.Preferences))
	api.Put("/preferences", authhandler.WithUser(ctrl.SavePreferences))
	api.Get("/risk", authhandler.WithUser(ctrl.Risk))
	api.Get("/fire", authhandler.WithUser(ctrl.Fire))
	api.Get("/shares", authhandler.WithUser(ctrl.Shares))
	api.Post("/shares", authhandler.WithUser(ctrl.CreateShare))
	api.Delete("/shares/:id", authhandler.WithUser(ctrl.RevokeShare))
	api.Post("/shared/report", authhandler.WithUser(ctrl.SharedReport))
}
func (ctrl *Controller) Watchlist(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Watchlist(userID)
	return result(c, v, e)
}
func (ctrl *Controller) SaveWatch(c *fiber.Ctx, userID uuid.UUID) error {
	var body model.WatchItem
	if e := parse(c, &body); e != nil {
		return e
	}
	v, e := ctrl.Service.AddWatch(userID, body)
	return result(c, v, e)
}
func (ctrl *Controller) DeleteWatch(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, true, ctrl.Service.DeleteWatch(userID, id))
}
func (ctrl *Controller) Assets(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Assets(userID)
	return result(c, v, e)
}
func (ctrl *Controller) SaveAsset(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := optionalID(c.Params("id"))
	if e != nil {
		return e
	}
	var body model.ManualAsset
	if e := parse(c, &body); e != nil {
		return e
	}
	v, e := ctrl.Service.SaveAsset(userID, id, body)
	return result(c, v, e)
}
func (ctrl *Controller) DeleteAsset(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, true, ctrl.Service.DeleteAsset(userID, id))
}
func (ctrl *Controller) Valuation(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	var b struct {
		Date  string   `json:"date"`
		Price *float64 `json:"price"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	if b.Price == nil {
		return fiber.NewError(400, "Price is required")
	}
	return result(c, true, ctrl.Service.SaveValuation(userID, id, b.Date, *b.Price))
}
func (ctrl *Controller) Liabilities(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Liabilities(userID)
	return result(c, v, e)
}
func (ctrl *Controller) CreateLiability(c *fiber.Ctx, userID uuid.UUID) error {
	var b struct {
		Name     string   `json:"name"`
		Currency string   `json:"currency"`
		Date     string   `json:"date"`
		Amount   *float64 `json:"amount"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	if b.Amount == nil {
		return fiber.NewError(400, "Amount is required")
	}
	v, e := ctrl.Service.CreateLiability(userID, b.Name, b.Currency, b.Date, *b.Amount)
	return result(c, v, e)
}
func (ctrl *Controller) LiabilityBalance(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	var b struct {
		Date   string   `json:"date"`
		Amount *float64 `json:"amount"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	if b.Amount == nil {
		return fiber.NewError(400, "Amount is required")
	}
	return result(c, true, ctrl.Service.SaveLiabilityBalance(userID, id, b.Date, *b.Amount))
}
func (ctrl *Controller) Tags(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Tags(userID)
	return result(c, v, e)
}
func (ctrl *Controller) SaveTag(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := optionalID(c.Params("id"))
	if e != nil {
		return e
	}
	var b struct {
		Name string `json:"name"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	v, e := ctrl.Service.SaveTag(userID, id, b.Name)
	return result(c, v, e)
}
func (ctrl *Controller) DeleteTag(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, true, ctrl.Service.DeleteTag(userID, id))
}
func (ctrl *Controller) EntityTags(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.EntityTags(userID, id, c.Params("kind"))
	return result(c, v, e)
}
func (ctrl *Controller) AssignTags(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	var b struct {
		IDs []uuid.UUID `json:"tag_ids"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	return result(c, true, ctrl.Service.AssignTags(userID, id, c.Params("kind"), b.IDs))
}
func (ctrl *Controller) Activities(c *fiber.Ctx, userID uuid.UUID) error {
	account, e := optionalID(c.Query("account_id"))
	if e != nil {
		return e
	}
	tag, e := optionalID(c.Query("tag_id"))
	if e != nil {
		return e
	}
	v, e := ctrl.Service.AllActivities(userID, service.ActivityFilter{AccountID: account, TagID: tag, Query: c.Query("q"), Type: c.Query("type"), From: c.Query("from"), To: c.Query("to"), Direction: c.Query("direction"), Page: c.QueryInt("page", 1), PerPage: c.QueryInt("per_page", 25)})
	if e != nil {
		return result(c, nil, e)
	}
	return apiresponse.RespondOK(c, v.Activities, apiresponse.WithPagination(v.Page, v.PerPage, v.Total))
}
func (ctrl *Controller) Preferences(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Preferences(userID)
	return result(c, v, e)
}
func (ctrl *Controller) SavePreferences(c *fiber.Ctx, userID uuid.UUID) error {
	var b model.Preferences
	if e := parse(c, &b); e != nil {
		return e
	}
	v, e := ctrl.Service.SavePreferences(userID, b)
	return result(c, v, e)
}
func (ctrl *Controller) Risk(c *fiber.Ctx, userID uuid.UUID) error {
	p, e := ctrl.Service.Preferences(userID)
	if e != nil {
		return result(c, nil, e)
	}
	r, e := ctrl.Service.Portfolio.Report(userID, nil, c.Query("range", "1y"), "")
	if e != nil {
		return result(c, nil, e)
	}
	return result(c, service.RiskChecks(r, p), nil)
}
func (ctrl *Controller) Fire(c *fiber.Ctx, userID uuid.UUID) error {
	p, e := ctrl.Service.Preferences(userID)
	if e != nil {
		return result(c, nil, e)
	}
	r, e := ctrl.Service.Portfolio.Report(userID, nil, "1d", "")
	if e != nil {
		return result(c, nil, e)
	}
	return result(c, service.FireProjection(r, p), nil)
}
func (ctrl *Controller) Shares(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Shares(userID)
	return result(c, v, e)
}
func (ctrl *Controller) CreateShare(c *fiber.Ctx, userID uuid.UUID) error {
	var b service.ShareRequest
	if e := parse(c, &b); e != nil {
		return e
	}
	v, token, e := ctrl.Service.CreateShare(userID, b)
	c.Set("Cache-Control", "no-store")
	return result(c, fiber.Map{"record": v, "token": token}, e)
}
func (ctrl *Controller) RevokeShare(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, true, ctrl.Service.RevokeShare(userID, id))
}
func (ctrl *Controller) SharedReport(c *fiber.Ctx, userID uuid.UUID) error {
	return ctrl.shared(c, &userID)
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
	return result(c, v, e)
}
