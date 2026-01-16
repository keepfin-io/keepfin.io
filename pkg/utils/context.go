package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var ErrUserNotAuthenticated = fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")

func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrUserNotAuthenticated
	}
	return userID, nil
}
