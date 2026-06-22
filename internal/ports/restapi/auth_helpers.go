package restapi

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrMissingUserID = errors.New("login error: missing user id") // rest
)

func UserID(c fiber.Ctx) (int, error) {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return 0, ErrMissingUserID
	}
	return userID, nil
}
