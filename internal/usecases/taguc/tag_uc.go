package taguc

import (
	"context"
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
)

type TagRepo interface {
	CreateTag(ctx context.Context, userID int, fullPath string) (*domain.Tag, error)
	ListByUser(ctx context.Context, userID int) ([]domain.Tag, error)
	GetTagsByPaths(ctx context.Context, userID int, paths []string) ([]domain.Tag, error)
	DeleteTag(ctx context.Context, userID int, tagID string) error
	UpdateTag(ctx context.Context, userID int, tagID string, newName string, newParentID *string) (*domain.Tag, error)
	SuggestTags(ctx context.Context, userID int, prefix string, limit int) ([]domain.Tag, error)
}

type TagUC struct {
	tagRepo TagRepo
}

func New(repo TagRepo) (*TagUC, error) {
	return &TagUC{tagRepo: repo}, nil
}

func (uc *TagUC) CreateTag(ctx context.Context, userID int, fullPath string) (*domain.Tag, error) {
	return uc.tagRepo.CreateTag(ctx, userID, fullPath)
}

func (uc *TagUC) ListTags(ctx context.Context, userID int) ([]domain.Tag, error) {
	return uc.tagRepo.ListByUser(ctx, userID)
}

func (uc *TagUC) DeleteTag(ctx context.Context, userID int, tagID string) error {
	return uc.tagRepo.DeleteTag(ctx, userID, tagID)
}

func (uc *TagUC) UpdateTag(ctx context.Context, userID int, tagID string, newName string, newParentID *string) (*domain.Tag, error) {
	return uc.tagRepo.UpdateTag(ctx, userID, tagID, newName, newParentID)
}

func (uc *TagUC) SuggestTags(ctx context.Context, userID int, prefix string, limit int) ([]domain.Tag, error) {
	return uc.tagRepo.SuggestTags(ctx, userID, prefix, limit)
}
