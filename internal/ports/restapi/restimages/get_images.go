package restimages

import (
	"strings"

	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/resttags"

	"github.com/gofiber/fiber/v3"
)

type GetImagesReq struct {
	TagPaths []string `json:"tag_paths"`
	TagIDs   []string `json:"tag_ids"`
}

type GetImagesResp struct {
	Images []ImageResp `json:"images"`
}

// GetImages ищет изображения по тегам.
//
//	@Summary		Поиск изображений по тегам
//	@Description	Поддерживает два режима: по UUID тегов (параметр `tag_ids`) и по текстовым путям (параметр `tag_paths`). Оба используют конъюнкцию с учётом иерархии. Если указаны оба, приоритет у tag_paths. Без ввода параметров выдает все изображения.
//	@Tags			images
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		GetImagesReq		true	"Массив UUID тегов или путей тегов (например, ['животные.кошки','пейзаж.горы'])"
//	@Success		200		{object}	GetImagesResp		"Список изображений"
//	@Failure		400		{object}	restapi.ErrorResp	"Неверный запрос"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Router			/api/v1/images [post]
func (h *Handler) GetImages() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		req := GetImagesReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return restapi.SendJSONParseError(c, err)
		}

		var (
			images []domain.Image
		)
		if len(req.TagPaths) > 0 {
			for i := range req.TagPaths {
				req.TagPaths[i] = strings.TrimSpace(req.TagPaths[i])
			}
			images, err = h.imageUC.SearchImagesByTagPaths(c.Context(), userID, req.TagPaths)
		} else {
			for i := range req.TagIDs {
				req.TagIDs[i] = strings.TrimSpace(req.TagIDs[i])
			}
			images, err = h.imageUC.SearchImagesByTags(c.Context(), userID, req.TagIDs)
		}
		if err != nil {
			return restapi.SendError(c, err)
		}

		resp := GetImagesResp{Images: make([]ImageResp, 0)}
		for _, v := range images {
			if len(v.Tags) <= 0 {
				v.Tags = make([]domain.Tag, 0)
			}

			tagsResp := make([]resttags.TagResp, 0, len(v.Tags))
			for _, t := range v.Tags {
				tagsResp = append(tagsResp, resttags.TagResp{
					ID:   t.ID,
					Path: t.Path,
					Name: t.Name,
				})
			}

			resp.Images = append(resp.Images, ImageResp{
				ID:         v.ID,
				Extension:  v.Extension,
				UploadedAt: v.UploadedAt,
				Tags:       tagsResp,
			})
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}
