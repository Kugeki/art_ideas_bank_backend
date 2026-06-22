package resttags

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/gofiber/fiber/v3"
)

type SuggestTagsResp struct {
	TagsIDs   []string `json:"tags_ids"`
	TagsPaths []string `json:"tags_paths"`

	Tags []TagResp `json:"tags"`
}

// SuggestTags предлагает варианты автодополнения тегов по префиксу пути.
//
//	@Summary		Автодополнение тегов
//	@Description	Возвращает список тегов, путь которых начинается с заданной строки (query-параметр `q`). Можно ограничить количество результатов через `limit` (по умолчанию 20).
//	@Tags			tags
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string				true	"Префикс для поиска (например, 'hair.p')"
//	@Param			limit	query		int					false	"Максимальное количество результатов (1-100, по умолчанию 20)"
//	@Success		200		{object}	SuggestTagsResp		"Подходящие теги"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный limit"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Router			/api/v1/tags/suggest [get]
func (h *Handler) SuggestTags() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		q := strings.TrimSpace(c.Query("q"))

		limitString := c.Query("limit", "20")
		limit, err := strconv.Atoi(limitString)
		if err != nil || limit <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(restapi.NewErrorResp(errors.New("wrong limit format: need a number")))
		}

		tags, err := h.tagUC.SuggestTags(c.Context(), userID, q, limit)
		if err != nil {
			return restapi.SendError(c, err)
		}

		resp := SuggestTagsResp{
			TagsIDs:   make([]string, 0),
			TagsPaths: make([]string, 0),
			Tags:      make([]TagResp, 0),
		}
		for _, v := range tags {
			resp.TagsPaths = append(resp.TagsPaths, v.Path)
			resp.TagsIDs = append(resp.TagsIDs, v.ID)
			resp.Tags = append(resp.Tags, TagResp{
				ID:   v.ID,
				Path: v.Path,
				Name: v.Name,
			})
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}
