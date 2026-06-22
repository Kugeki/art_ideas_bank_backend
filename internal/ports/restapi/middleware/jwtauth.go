package middleware

import (
	"errors"
	"strings"

	"github.com/Kugeki/art_ideas_bank_backend/internal/adapters/jwtauth"

	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrInvalidAuthFormat = errors.New("invalid authorization format")
	ErrMissingAuthToken  = errors.New("missing authorization token")
)

func AuthRequired(jwtAuth *jwtauth.JWT) fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenStr string

		header := c.Get("Authorization")
		if header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.Status(fiber.StatusUnauthorized).JSON(restapi.NewErrorResp(ErrInvalidAuthFormat))
			}
			tokenStr = parts[1]
		} else {
			tokenStr = c.Cookies("token")
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(restapi.NewErrorResp(ErrMissingAuthToken))
		}

		claims, err := jwtAuth.ParseToken(tokenStr)
		if err != nil {
			return restapi.SendError(c, err)
		}
		c.Locals("userID", claims.UserID)
		return c.Next()
	}
}
