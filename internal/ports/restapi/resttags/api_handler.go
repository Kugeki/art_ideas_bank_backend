package resttags

import (
	"context"
	"log/slog"

	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/pkg/slogdiscard"

	"github.com/gofiber/fiber/v3"
)

type TagUsecase interface {
	CreateTag(ctx context.Context, userID int, fullPath string) (*domain.Tag, error)
	ListTags(ctx context.Context, userID int) ([]domain.Tag, error)
	DeleteTag(ctx context.Context, userID int, tagID string) error
	UpdateTag(ctx context.Context, userID int, tagID string, newName string, newParentID *string) (*domain.Tag, error)
	SuggestTags(ctx context.Context, userID int, prefix string, limit int) ([]domain.Tag, error)
}

type Handler struct {
	log *slog.Logger

	tagUC TagUsecase
}

func NewHandler(log *slog.Logger, tagUC TagUsecase) *Handler {
	log = slogdiscard.LoggerIfNil(log)

	return &Handler{
		log:   log,
		tagUC: tagUC,
	}
}

func (h *Handler) SetupRotes(parent *fiber.App, authMiddleware fiber.Handler) {
	api := parent.Group("/api/v1/tags", authMiddleware)

	api.Get("/", h.GetTags())
	api.Post("/", h.CreateTag())
	api.Delete("/:id", h.DeleteTag())
	api.Put("/:id", h.UpdateTag())
	api.Get("/suggest", h.SuggestTags())
}
