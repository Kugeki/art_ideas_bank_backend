package restusers

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

type UserLoginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserLoginResp struct {
	Token string `json:"token"`
}

// UserLogin выполняет аутентификацию пользователя.
//
//	@Summary		Вход в систему
//	@Description	Проверяет email/пароль, возвращает JWT-токен и устанавливает HttpOnly cookie.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		UserLoginReq		true	"Учётные данные"
//	@Success		200		{object}	UserLoginResp		"Токен авторизации"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный запрос"
//	@Failure		401		{object}	restapi.ErrorResp	"Неверные учётные данные"
//	@Failure		500		{object}	restapi.ErrorResp	"Внутренняя ошибка"
//	@Router			/api/v1/users/login [post]
func (h *Handler) UserLogin() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

		req := UserLoginReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return restapi.SendJSONParseError(c, err)
		}

		if err := h.validate.Struct(req); err != nil {
			return restapi.SendValidationError(c, err)
		}

		user, err := h.userUC.VerifyUser(c.Context(), req.Email, req.Password)
		if err != nil {
			return restapi.SendError(c, err)
		}

		token, err := h.jwtAuth.GenerateToken(user.ID)
		if err != nil {
			return restapi.SendError(c, err)
		}

		// Устанавливаем cookie
		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    token,
			HTTPOnly: true,
			Secure:   false, // для разработки; в production true при HTTPS
			SameSite: fiber.CookieSameSiteLaxMode,
			MaxAge:   24 * 3600, // 24 часа, синхронно с временем жизни токена
			Path:     "/",
		})

		return c.Status(fiber.StatusOK).JSON(UserLoginResp{Token: "Bearer " + token})
	}
}
