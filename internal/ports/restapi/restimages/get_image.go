package restimages

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/resttags"

	"github.com/gofiber/fiber/v3"
)

type GetImageResp struct {
	ImageResp
}

// GetImage возвращает информацию об изображении и его тегах.
//
//	@Summary		Получить информацию об изображении
//	@Description	Возвращает метаданные изображения и массив привязанных тегов.
//	@Tags			images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"UUID изображения"
//	@Success		200	{object}	GetImageResp		"Информация об изображении"
//	@Failure		400	{object}	restapi.ErrorResp	"Пустой ID"
//	@Failure		401	{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		404	{object}	restapi.ErrorResp	"Изображение не найдено"
//	@Router			/api/v1/images/{id} [get]
func (h *Handler) GetImage() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		imageID := c.Params("id")

		img, err := h.imageUC.GetImage(c.Context(), userID, imageID)
		if err != nil {
			return restapi.SendError(c, err)
		}

		tagsResp := make([]resttags.TagResp, 0, len(img.Tags))
		for _, v := range img.Tags {
			tagsResp = append(tagsResp, resttags.TagResp{
				ID:   v.ID,
				Path: v.Path,
				Name: v.Name,
			})
		}
		return c.Status(fiber.StatusOK).JSON(GetImageResp{
			ImageResp: ImageResp{
				ID:         img.ID,
				Extension:  img.Extension,
				UploadedAt: img.UploadedAt,
				Tags:       tagsResp,
			},
		})
	}
}
