package resttags

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

type GetTagsResp struct {
	Tags []TagResp `json:"tags"`
}

// GetTags возвращает все теги пользователя.
//
//	@Summary		Список тегов пользователя
//	@Description	Возвращает все теги, созданные пользователем, отсортированные по пути.
//	@Tags			tags
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	GetTagsResp			"Список тегов"
//	@Failure		401	{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Router			/api/v1/tags [get]
func (h *Handler) GetTags() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		tags, err := h.tagUC.ListTags(c.Context(), userID)
		if err != nil {
			return restapi.SendError(c, err)
		}

		resp := GetTagsResp{
			Tags: make([]TagResp, 0, len(tags)),
		}
		for _, v := range tags {
			resp.Tags = append(resp.Tags, TagResp{
				ID:   v.ID,
				Path: v.Path,
				Name: v.Name,
			})
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}
