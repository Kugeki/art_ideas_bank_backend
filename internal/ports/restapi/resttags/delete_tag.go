package resttags

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

// DeleteTag удаляет тег и все его поддеревья, если с ними не связано ни одно изображение.
//
//	@Summary		Удалить тег
//	@Description	Удаляет тег и всех его потомков, только если ни к одному из них не привязаны изображения.
//	@Tags			tags
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID тега"
//	@Success		204	"Тег удалён"
//	@Failure		400	{object}	restapi.ErrorResp	"Пустой ID"
//	@Failure		401	{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		404	{object}	restapi.ErrorResp	"Тег не найден"
//	@Failure		409	{object}	restapi.ErrorResp	"Тег используется изображениями"
//	@Router			/api/v1/tags/{id} [delete]
func (h *Handler) DeleteTag() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		tagID := c.Params("id")
		if err := h.tagUC.DeleteTag(c.Context(), userID, tagID); err != nil {
			return restapi.SendError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
