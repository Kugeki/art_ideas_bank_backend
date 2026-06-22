package restimages

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"log/slog"
	"path/filepath"

	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"
)

type UploadImageResp struct {
	ID        string `json:"id"`
	Extension string `json:"extension"`
	Link      string `json:"link"`
}

// UploadImage загружает новое изображение.
//
//	@Summary		Загрузить изображение
//	@Description	Загружает изображение в хранилище пользователя. Возвращает ID, расширение и прямую ссылку для скачивания.
//	@Tags			images
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			image	formData	file				true	"Файл изображения"
//	@Success		201		{object}	UploadImageResp		"Успешная загрузка"
//	@Failure		400		{object}	restapi.ErrorResp	"Отсутствует файл или неверный запрос"
//	@Failure		401		{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		500		{object}	restapi.ErrorResp	"Внутренняя ошибка"
//	@Router			/api/v1/images/upload [post]
func (h *Handler) UploadImage() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		file, err := c.FormFile("image")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(restapi.NewErrorResp(fmt.Errorf("need image file with 'image' key: %w", err)))
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(restapi.NewErrorResp(fmt.Errorf("can't open image: %w", err)))
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		contentType := file.Header.Get("Content-Type")

		if !domain.IsTypeImageOrVideo(contentType) {
			return restapi.SendError(c, fmt.Errorf("content type (%v): %w",
				contentType, domain.ErrIncorrectContentType))
		}

		img, err := h.imageUC.Upload(c.Context(), userID, src, ext)
		if err != nil {
			h.log.Error("image upload error", slog.Any("error", err))
			return restapi.SendError(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(UploadImageResp{
			ID:        img.ID,
			Extension: img.Extension,
			Link:      fmt.Sprintf("/api/v1/images/download/%v%v", img.ID, img.Extension),
		})
	}
}
