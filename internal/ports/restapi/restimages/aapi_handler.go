package restimages

import (
	"context"
	"io"
	"log/slog"

	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/pkg/slogdiscard"

	"github.com/gofiber/fiber/v3"
)

type ImageUsecase interface {
	Upload(ctx context.Context, userID int, file domain.File, extension string) (*domain.Image, error)
	GetImage(ctx context.Context, userID int, imageID string) (*domain.Image, error)
	GetUserImages(ctx context.Context, userID int) ([]domain.Image, error)
	Download(ctx context.Context, userID int, imageID string) (io.ReadCloser, string, error)
	DeleteImage(ctx context.Context, userID int, imageID string) error

	AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error
	RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error
	SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error)
	SearchImagesByTagPaths(ctx context.Context, userID int, paths []string) ([]domain.Image, error)
}

type Handler struct {
	log *slog.Logger

	imageUC ImageUsecase
}

func NewHandler(log *slog.Logger, imageUC ImageUsecase) *Handler {
	log = slogdiscard.LoggerIfNil(log)

	return &Handler{
		log:     log,
		imageUC: imageUC,
	}
}

func (h *Handler) SetupRotes(parent *fiber.App, authMiddleware fiber.Handler) {
	api := parent.Group("/api/v1/images", authMiddleware)

	api.Post("/", h.GetImages())
	api.Get("/:id", h.GetImage())
	api.Post("/upload", h.UploadImage())
	api.Get("/download/:id", h.DownloadImage())
	api.Delete("/:id", h.DeleteImage())
	api.Post("/:id/tags", h.AddTagsToImage())
	api.Delete("/:id/tags", h.RemoveTagsFromImage())
}
