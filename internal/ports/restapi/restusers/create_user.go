package restusers

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/gofiber/fiber/v3"
)

type UsersCreateReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=8"`
}

// CreateUser регистрирует нового пользователя
//
//	@Summary		Регистрация пользователя
//	@Description	Создаёт учётную запись пользователя по email и паролю.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body	UsersCreateReq	true	"Данные пользователя"
//	@Success		201		"Пользователь создан"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный запрос"
//	@Failure		409		{object}	restapi.ErrorResp	"Уже существует"
//	@Failure		500		{object}	restapi.ErrorResp	"Внутренняя ошибка"
//	@Router			/api/v1/users/create [post]
func (h *Handler) CreateUser() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

		req := UsersCreateReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return restapi.SendJSONParseError(c, err)
		}

		if err := h.validate.Struct(req); err != nil {
			return restapi.SendValidationError(c, err)
		}

		err := h.userUC.CreateUser(c.Context(), &domain.User{Email: req.Email}, req.Password)
		if err != nil {
			return restapi.SendError(c, err)
		}

		return c.SendStatus(fiber.StatusCreated)
	}
}
