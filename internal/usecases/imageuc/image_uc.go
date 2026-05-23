package imageuc

import (
	"art_ideas_bank_backend/internal/adapters/s3client"
	"art_ideas_bank_backend/internal/domain"
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
	"time"
)

type ImageRepo interface {
	Create(ctx context.Context, img *domain.Image) error
	GetByID(ctx context.Context, userID int, imageID string) (*domain.Image, error)
	ListByUser(ctx context.Context, userID int) ([]domain.Image, error)
	GetByKey(ctx context.Context, key string) (*domain.Image, error)
}

type TagRepo interface {
	GetTagsForImage(ctx context.Context, userID int, imageID string) ([]domain.Tag, error)
}

type ImageUC struct {
	imageRepo ImageRepo
	tagRepo   TagRepo
	s3        *s3client.S3Client
}

func New(repo ImageRepo, s3 *s3client.S3Client) (*ImageUC, error) {
	return &ImageUC{imageRepo: repo, s3: s3}, nil
}

func (uc *ImageUC) Upload(ctx context.Context, userID int, file io.Reader, contentType, extension string) (*domain.Image, error) {
	key := fmt.Sprintf("%d/%s%s", userID, uuid.New().String(), extension)

	if err := uc.s3.Upload(ctx, key, file, contentType); err != nil {
		return nil, fmt.Errorf("s3 upload: %w", err)
	}

	img := &domain.Image{
		UserID:     userID,
		S3Key:      key,
		UploadedAt: time.Now(),
	}
	if err := uc.imageRepo.Create(ctx, img); err != nil {
		// Откат загрузки? Можно просто удалить из S3, но для простоты оставим так
		return nil, err
	}

	return img, nil
}

func (uc *ImageUC) GetImage(ctx context.Context, userID int, imageID string) (*domain.Image, error) {
	img, err := uc.imageRepo.GetByID(ctx, userID, imageID)
	if err != nil {
		return nil, err
	}
	tags, err := uc.tagRepo.GetTagsForImage(ctx, userID, imageID)
	if err != nil {
		return nil, err
	}
	img.Tags = tags

	return img, nil
}

func (uc *ImageUC) GetUserImages(ctx context.Context, userID int) ([]domain.Image, error) {
	return uc.imageRepo.ListByUser(ctx, userID)
}

func (uc *ImageUC) Download(ctx context.Context, userID int, imageID string) (io.ReadCloser, string, error) {
	img, err := uc.imageRepo.GetByID(ctx, userID, imageID)
	if err != nil || img.UserID != userID {
		return nil, "", domain.ErrNotFound
	}
	return uc.s3.Download(ctx, img.S3Key)
}
