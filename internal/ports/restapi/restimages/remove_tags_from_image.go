package restimages

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

type RemoveTagsFromImageReq struct {
	TagIDs []string `json:"tag_ids"`
}

// RemoveTagsFromImage открепляет теги от изображения.
//
//	@Summary		Удалить теги с изображения
//	@Description	Принимает список UUID тегов и удаляет их связь с изображением.
//	@Tags			images
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string					true	"UUID изображения"
//	@Param			request	body	RemoveTagsFromImageReq	true	"Массив UUID тегов"
//	@Success		204		"Теги удалены"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный запрос"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		403		{object}	restapi.ErrorResp	"Доступ запрещён"
//	@Failure		404		{object}	restapi.ErrorResp	"Не найдено"
//	@Router			/api/v1/images/{id}/tags [delete]
func (h *Handler) RemoveTagsFromImage() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}
		imageID := c.Params("id")

		req := RemoveTagsFromImageReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return restapi.SendJSONParseError(c, err)
		}
		if err := h.imageUC.RemoveTagsFromImage(c.Context(), userID, imageID, req.TagIDs); err != nil {
			return restapi.SendError(c, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
