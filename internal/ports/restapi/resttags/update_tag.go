package resttags

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
	"github.com/gofiber/fiber/v3"
)

type UpdateTagReq struct {
	Name        string  `json:"name"`
	NewParentID *string `json:"new_parent_id"`
}

type UpdateTagResp struct {
	TagResp
}

// UpdateTag изменяет имя и/или родителя тега. При смене родителя всё поддерево перемещается.
//
//	@Summary		Обновить тег
//	@Description	Позволяет переименовать тег (поле `name`) и/или переместить его под другого родителя (`new_parent_id` – UUID родительского тега). Поддерево обновляется рекурсивно.
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"UUID тега"
//	@Param			request	body		UpdateTagReq		true	"Новое имя и (опционально) UUID нового родителя"
//	@Success		200		{object}	UpdateTagResp		"Обновлённый тег"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный запрос"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		404		{object}	restapi.ErrorResp	"Тег не найден"
//	@Failure		409		{object}	restapi.ErrorResp	"Конфликт (тег с таким путём уже существует или попытка создать цикл)"
//	@Router			/api/v1/tags/{id} [put]
func (h *Handler) UpdateTag() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		tagID := c.Params("id")
		req := UpdateTagReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return restapi.SendJSONParseError(c, err)
		}

		tag, err := h.tagUC.UpdateTag(c.Context(), userID, tagID, req.Name, req.NewParentID)
		if err != nil {
			return restapi.SendError(c, err)
		}

		return c.JSON(UpdateTagResp{
			TagResp: TagResp{
				ID:   tag.ID,
				Path: tag.Path,
				Name: tag.Name,
			},
		})
	}
}
