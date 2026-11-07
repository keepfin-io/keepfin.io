package wealth

import (
	"github.com/gofiber/fiber/v2"
	model "keepnode/app/models/wealth"
)

func (ctrl *Controller) Providers(c *fiber.Ctx) error {
	v, e := ctrl.Service.MarketConfiguration()
	return result(c, "configuration", v, e)
}
func (ctrl *Controller) SetProvider(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var b struct {
		Provider string `json:"provider"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	return result(c, "saved", true, ctrl.Service.SetProvider(u, b.Provider))
}
func (ctrl *Controller) Jobs(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Jobs(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) QueueJob(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var b struct {
		Symbol string `json:"symbol"`
		From   string `json:"from"`
		To     string `json:"to"`
	}
	if e = parse(c, &b); e != nil {
		return e
	}
	v, e := ctrl.Service.QueueJob(u, b.Symbol, b.From, b.To)
	return result(c, "item", v, e)
}
func (ctrl *Controller) RetryJob(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	return result(c, "queued", true, ctrl.Service.RetryJob(u, id))
}
func (ctrl *Controller) Corrections(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.Overrides(u)
	return result(c, "items", v, e)
}
func (ctrl *Controller) CorrectPrice(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	var b model.PriceOverride
	if e = parse(c, &b); e != nil {
		return e
	}
	return result(c, "saved", true, ctrl.Service.CorrectPrice(u, b))
}
func (ctrl *Controller) DeleteCorrection(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	return result(c, "deleted", true, ctrl.Service.DeleteCorrection(u, c.Query("symbol"), c.Query("date")))
}
func (ctrl *Controller) Dividends(c *fiber.Ctx) error {
	u, e := identity(c)
	if e != nil {
		return e
	}
	id, e := recordID(c)
	if e != nil {
		return e
	}
	v, e := ctrl.Service.PreviewDividends(u, id, c.Query("symbol"), c.Query("from"), c.Query("to"))
	return result(c, "items", v, e)
}
