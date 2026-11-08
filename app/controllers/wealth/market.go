package wealth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	model "keepnode/app/models/wealth"
)

func (ctrl *Controller) Providers(c *fiber.Ctx, _ uuid.UUID) error {
	v, e := ctrl.Service.MarketConfiguration()
	return result(c, v, e)
}
func (ctrl *Controller) SetProvider(c *fiber.Ctx, userID uuid.UUID) error {
	var b struct {
		Provider string `json:"provider"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	return result(c, true, ctrl.Service.SetProvider(userID, b.Provider))
}
func (ctrl *Controller) Jobs(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Jobs(userID)
	return result(c, v, e)
}
func (ctrl *Controller) QueueJob(c *fiber.Ctx, userID uuid.UUID) error {
	var b struct {
		Symbol string `json:"symbol"`
		From   string `json:"from"`
		To     string `json:"to"`
	}
	if e := parse(c, &b); e != nil {
		return e
	}
	v, e := ctrl.Service.QueueJob(userID, b.Symbol, b.From, b.To)
	return result(c, v, e)
}
func (ctrl *Controller) RetryJob(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, true, ctrl.Service.RetryJob(userID, id))
}
func (ctrl *Controller) Corrections(c *fiber.Ctx, userID uuid.UUID) error {
	v, e := ctrl.Service.Overrides(userID)
	return result(c, v, e)
}
func (ctrl *Controller) CorrectPrice(c *fiber.Ctx, userID uuid.UUID) error {
	var b model.PriceOverride
	if e := parse(c, &b); e != nil {
		return e
	}
	return result(c, true, ctrl.Service.CorrectPrice(userID, b))
}
func (ctrl *Controller) DeleteCorrection(c *fiber.Ctx, userID uuid.UUID) error {
	return result(c, true, ctrl.Service.DeleteCorrection(userID, c.Query("symbol"), c.Query("date")))
}
func (ctrl *Controller) Dividends(c *fiber.Ctx, userID uuid.UUID) error {
	id, e := recordID(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.PreviewDividends(userID, id, c.Query("symbol"), c.Query("from"), c.Query("to"))
	return result(c, v, e)
}
