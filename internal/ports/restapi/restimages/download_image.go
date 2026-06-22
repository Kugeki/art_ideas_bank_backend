package restimages

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"log/slog"
	"strings"

	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi"

	"github.com/gofiber/fiber/v3/log"
)

// DownloadImage отдаёт файл изображения.
//
//	@Summary		Скачать изображение
//	@Description	Возвращает бинарные данные изображения по его ID. Поддерживает cookie-аутентификацию для прямого открытия в браузере.
//	@Tags			images
//	@Produce		octet-stream
//	@Security		BearerAuth
//	@Param			id	path		string				true	"UUID изображения"
//	@Success		200	{file}		binary				"Изображение"
//	@Failure		401	{object}	restapi.ErrorResp	"Требуется аутентификация"
//	@Failure		404	{object}	restapi.ErrorResp	"Не найдено"
//	@Router			/api/v1/images/download/{id} [get]
func (h *Handler) DownloadImage() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := restapi.UserID(c)
		if err != nil {
			return restapi.SendError(c, err)
		}

		imageID := strings.SplitN(c.Params("id"), ".", 2)[0] // if id provided with image extension (for example, *.png)

		log.Info("image download", slog.String("id", imageID))
		body, contentType, err := h.imageUC.Download(c.Context(), userID, imageID)
		if err != nil {
			log.Error("image download error", slog.Any("error", err))
			return restapi.SendError(c, err)
		}

		if !(strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")) {
			return restapi.SendError(c, fmt.Errorf("incorrect content type: %v", contentType))
		}

		c.Set("Content-Type", contentType)
		return c.Status(fiber.StatusOK).SendStream(body)
	}
}
