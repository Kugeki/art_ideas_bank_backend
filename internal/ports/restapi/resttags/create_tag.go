package resttags

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/gofiber/fiber/v3"
)

type CreateTagReq struct {
	Path string `json:"path" `
}

type CreateTagResp struct {
	TagResp
}

// CreateTag создаёт новый иерархический тег.
//
//	@Summary		Создать тег
//	@Description	Принимает полный путь тега (например "животные.кошки"). Автоматически создаёт все недостающие родительские узлы.
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateTagReq		true	"Полный путь тега"
//	@Success		201		{object}	CreateTagResp		"Созданный тег"
//	@Failure		400		{object}	restapi.ErrorResp	"Некорректный запрос (путь не задан)"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		409		{object}	restapi.ErrorResp	"Тег уже существует"
//	@Router			/api/v1/tags [post]
func (h *Handler) CreateTag() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		req := CreateTagReq{}
		if err := c.Bind().JSON(&req); err != nil || req.Path == "" {
			return restapi.SendJSONParseError(c, err)
		}
		tag, err := h.tagUC.CreateTag(c.Context(), userID, req.Path)
		if err != nil {
			return restapi.SendError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(CreateTagResp{
			TagResp: TagResp{
				ID:   tag.ID,
				Path: tag.Path,
				Name: tag.Name,
			},
		})
	}
}
