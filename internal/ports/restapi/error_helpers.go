package restapi

import (
	"errors"
	"fmt"

	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/gofiber/fiber/v3"
)

func ErrorToStatus(err error) int {
	ErrIs := func(target error) bool {
		return errors.Is(err, target)
	}

	switch {
	case ErrIs(domain.ErrNotFound):
		fallthrough
	case ErrIs(domain.ErrSomeTagsNotFound):
		return fiber.StatusNotFound

	case ErrIs(domain.ErrAlreadyExists):
		fallthrough
	case ErrIs(domain.ErrTagHasAssociatedImages):
		fallthrough
	case ErrIs(domain.ErrCannotMoveTagToDescendant):
		return fiber.StatusConflict

	case ErrIs(domain.ErrEmptyTagPath):
		fallthrough
	case ErrIs(domain.ErrTagIncorrectName):
		fallthrough
	case ErrIs(domain.ErrIncorrectContentType):
		fallthrough
	case ErrIs(domain.ErrLimitEqualOrLessThanZero):
		return fiber.StatusBadRequest

	case ErrIs(domain.ErrWrongCredentials):
		fallthrough
	case ErrIs(domain.ErrInvalidToken):
		fallthrough
	case ErrIs(ErrMissingUserID):
		return fiber.StatusUnauthorized
	}

	return fiber.StatusInternalServerError
}

type ErrorResp struct {
	Error string `json:"error"`
}

func NewErrorResp(err error) ErrorResp {
	return ErrorResp{Error: err.Error()}
}

// SendError is function to send domain error with auto status detection
func SendError(c fiber.Ctx, err error) error {
	resp := ErrorResp{Error: err.Error()}
	return c.Status(ErrorToStatus(err)).JSON(resp)
}

func SendStatusWithError(c fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(NewErrorResp(err))
}

func SendBadRequestError(c fiber.Ctx, err error) error {
	return SendStatusWithError(c, fiber.StatusBadRequest, err)
}

func SendJSONParseError(c fiber.Ctx, err error) error {
	return SendBadRequestError(c, fmt.Errorf("json parse error: %w", err))
}

func SendValidationError(c fiber.Ctx, err error) error {
	return SendBadRequestError(c, fmt.Errorf("validation error: %w", err))
}
