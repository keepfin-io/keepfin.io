package market

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	marketmodel "keepnode/app/models/market"
	model "keepnode/app/models/wealth"
	"keepnode/pkg/apiresponse"
	"strings"
	"time"
)

func (ctrl *MarketController) manual(c *fiber.Ctx, userID uuid.UUID, symbol string) error {
	if ctrl.db == nil {
		return fiber.NewError(404, "Manual asset not found")
	}
	var asset model.ManualAsset
	if err := ctrl.db.First(&asset, "user_id = ? AND symbol = ?", userID, symbol).Error; err != nil {
		return fiber.NewError(404, "Manual asset not found")
	}
	rows := []model.ManualValuation{}
	q := ctrl.db.Where("asset_id = ? AND date <= ?", asset.ID, time.Now().UTC().Format("2006-01-02"))
	live := strings.Contains(c.Path(), "/live/")
	if !live {
		for _, key := range []string{"from", "to"} {
			if value := c.Query(key); value != "" {
				if _, err := time.Parse("2006-01-02", value); err != nil {
					return fiber.NewError(400, "Invalid date")
				}
				if key == "from" {
					q = q.Where("date >= ?", value)
				} else {
					q = q.Where("date <= ?", value)
				}
			}
		}
	}
	if live {
		q = q.Order("date desc").Limit(1)
	} else {
		q = q.Order("date")
	}
	if err := q.Find(&rows).Error; err != nil {
		return fiber.NewError(500, "Unable to load valuations")
	}
	if live {
		if len(rows) == 0 {
			return fiber.NewError(404, "Add a dated valuation for this asset")
		}
		day, _ := time.Parse("2006-01-02", rows[0].Date)
		return apiresponse.RespondOK(c, marketmodel.LivePrice{Code: symbol, Close: rows[0].Price, Timestamp: day.Unix()})
	}
	data := []marketmodel.EODData{}
	for _, v := range rows {
		data = append(data, marketmodel.EODData{Date: v.Date, Open: v.Price, High: v.Price, Low: v.Price, Close: v.Price, AdjustedClose: v.Price})
	}
	return apiresponse.RespondOK(c, data)
}
