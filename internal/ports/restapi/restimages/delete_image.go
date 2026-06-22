package restimages

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

// DeleteImage удаляет изображение
//
//	@Summary		Удалить изображение
//	@Description	Удаляет изображение из хранилища и базы данных. Все связанные теги открепляются.
//	@Tags			images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID изображения"
//	@Success		204	"Изображение удалено"
//	@Failure		400	{object}	restapi.ErrorResp	"Пустой ID"
//	@Failure		401	{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		404	{object}	restapi.ErrorResp	"Изображение не найдено"
//	@Failure		500	{object}	restapi.ErrorResp	"Внутренняя ошибка"
//	@Router			/api/v1/images/{id} [delete]
func (h *Handler) DeleteImage() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		imageID := c.Params("id")
		err = h.imageUC.DeleteImage(c.Context(), userID, imageID)
		if err != nil {
			return restapi.SendError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
