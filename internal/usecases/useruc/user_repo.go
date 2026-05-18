package useruc

import (
	"art_ideas_bank_backend/internal/domain"
	"context"
)

type UserRepo interface {
	CreateUser(ctx context.Context, u *domain.User) error
	GetUser(ctx context.Context, email string) (*domain.User, error)
}
