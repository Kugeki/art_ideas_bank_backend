package taguc

import (
	"art_ideas_bank_backend/internal/domain"
	"context"
)

type TagRepo interface {
	CreateTag(ctx context.Context, userID int, fullPath string) (*domain.Tag, error)
	ListByUser(ctx context.Context, userID int) ([]domain.Tag, error)
	GetTagsForImage(ctx context.Context, userID int, imageID string) ([]domain.Tag, error)
	DeleteTag(ctx context.Context, userID int, tagID string) error
	UpdateTag(ctx context.Context, userID int, tagID string, newName string, newParentID *string) (*domain.Tag, error)
	AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error
	RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error
	SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error)
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

func (uc *TagUC) AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	return uc.tagRepo.AddTagsToImage(ctx, userID, imageID, tagIDs)
}

func (uc *TagUC) RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	return uc.tagRepo.RemoveTagsFromImage(ctx, userID, imageID, tagIDs)
}

func (uc *TagUC) SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error) {
	return uc.tagRepo.SearchImagesByTags(ctx, userID, tagIDs)
}
