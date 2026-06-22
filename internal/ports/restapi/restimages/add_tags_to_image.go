package restimages

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

type AddTagsToImageReq struct {
	TagIDs []string `json:"tag_ids"`
}

// AddTagsToImage привязывает теги к изображению.
//
//	@Summary		Добавить теги к изображению
//	@Description	Принимает список UUID тегов и связывает их с изображением. Теги должны принадлежать текущему пользователю.
//	@Tags			images
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"UUID изображения"
//	@Param			request	body	AddTagsToImageReq	true	"Массив UUID тегов"
//	@Success		204		"Теги добавлены"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный запрос"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		403		{object}	restapi.ErrorResp	"Не найдено"
//	@Router			/api/v1/images/{id}/tags [post]
func (h *Handler) AddTagsToImage() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		imageID := c.Params("id")

		req := AddTagsToImageReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return restapi.SendJSONParseError(c, err)
		}
		if err := h.imageUC.AddTagsToImage(c.Context(), userID, imageID, req.TagIDs); err != nil {
			return restapi.SendError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
