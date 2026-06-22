package restusers

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/pkg/slogdiscard"

	"github.com/gofiber/fiber/v3"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, u *domain.User, password string) error
	VerifyUser(ctx context.Context, email string, password string) (*domain.User, error)
}

type JwtAuth interface {
	GenerateToken(userID int) (string, error)
	ParseToken(tokenStr string) (*domain.Claims, error)
}

type Handler struct {
	log      *slog.Logger
	validate *validator.Validate

	userUC  UserUsecase
	jwtAuth JwtAuth
}

func NewHandler(log *slog.Logger, v *validator.Validate, userUC UserUsecase, jwtAuth JwtAuth) *Handler {
	log = slogdiscard.LoggerIfNil(log)

	return &Handler{
		log:      log,
		validate: v,
		userUC:   userUC,
		jwtAuth:  jwtAuth,
	}
}

func (h *Handler) SetupRotes(parent *fiber.App) {
	api := parent.Group("/api/v1/users")

	api.Post("/create", h.CreateUser())
	api.Post("/login", h.UserLogin())
}
