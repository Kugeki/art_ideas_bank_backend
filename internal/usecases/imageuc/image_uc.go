package imageuc

import (
	"context"
	"fmt"
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"io"
	"strings"
	"time"
)

type ImageRepo interface {
	Create(ctx context.Context, img *domain.Image) error
	DeleteImage(ctx context.Context, userID int, imageID string) error
	GetByID(ctx context.Context, userID int, imageID string) (*domain.Image, error)
	ListByUser(ctx context.Context, userID int) ([]domain.Image, error)
	GetByKey(ctx context.Context, key string) (*domain.Image, error)

	GetTagsForImage(ctx context.Context, userID int, imageID string) ([]domain.Tag, error)
	AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error
	RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error
	SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error)
}

type TagRepo interface {
	GetTagsByPaths(ctx context.Context, userID int, paths []string) ([]domain.Tag, error)
}

type S3 interface {
	Test(ctx context.Context) (*s3.ListBucketsOutput, error)
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
}

type ContentTypeDetector interface {
	DetectFileContentType(src domain.File) (contentType string, newSrc domain.File, err error)
	DetectReaderContentType(src io.ReadCloser) (contentType string, newSrc io.ReadCloser, err error)
	GetExtensionContentType(ext string) (contentType string, err error)
}

type ImageUC struct {
	imageRepo ImageRepo
	tagRepo   TagRepo
	s3        S3
	detector  ContentTypeDetector
}

func New(imageRepo ImageRepo, tagRepo TagRepo, s3 S3, detector ContentTypeDetector) (*ImageUC, error) {
	return &ImageUC{imageRepo: imageRepo, tagRepo: tagRepo, s3: s3, detector: detector}, nil
}

func (uc *ImageUC) Upload(ctx context.Context, userID int, file domain.File, extension string) (*domain.Image, error) {
	extension = strings.ToLower(extension)

	newFile, contentType, err := uc.verifyFileContentType(file, extension)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%d/%s%s", userID, uuid.New().String(), extension)
	if err := uc.s3.Upload(ctx, key, newFile, contentType); err != nil {
		return nil, fmt.Errorf("s3 upload: %w", err)
	}

	img := &domain.Image{
		UserID:     userID,
		Extension:  extension,
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
	tags, err := uc.imageRepo.GetTagsForImage(ctx, userID, imageID)
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

	file, contentType, err := uc.s3.Download(ctx, img.S3Key)
	if err != nil {
		return nil, "", err
	}

	newFile, _, err := uc.verifyReaderContentType(file, img.Extension)
	if err != nil {
		return nil, "", err
	}

	return newFile, contentType, nil
}

func (uc *ImageUC) DeleteImage(ctx context.Context, userID int, imageID string) error {
	img, err := uc.imageRepo.GetByID(ctx, userID, imageID)
	if err != nil {
		return err
	}

	err = uc.s3.Delete(ctx, img.S3Key)
	if err != nil {
		return err
	}

	err = uc.imageRepo.DeleteImage(ctx, userID, imageID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *ImageUC) AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	return uc.imageRepo.AddTagsToImage(ctx, userID, imageID, tagIDs)
}

func (uc *ImageUC) RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	return uc.imageRepo.RemoveTagsFromImage(ctx, userID, imageID, tagIDs)
}

func (uc *ImageUC) SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error) {
	return uc.imageRepo.SearchImagesByTags(ctx, userID, tagIDs)
}

func (uc *ImageUC) SearchImagesByTagPaths(ctx context.Context, userID int, paths []string) ([]domain.Image, error) {
	tags, err := uc.tagRepo.GetTagsByPaths(ctx, userID, paths)
	if err != nil {
		return nil, err
	}
	tagIDs := make([]string, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}
	return uc.imageRepo.SearchImagesByTags(ctx, userID, tagIDs)
}

func (uc *ImageUC) verifyExtensionContentType(extension string) (string, error) {
	extContentType, err := uc.detector.GetExtensionContentType(extension)
	if err != nil {
		return "", err
	}
	if !domain.IsTypeImageOrVideo(extContentType) {
		return "", fmt.Errorf("extension (%v), ext content type (%v): %w",
			extension, extContentType, domain.ErrIncorrectContentType)
	}

	return extContentType, nil
}

func (uc *ImageUC) verifyFileContentType(file domain.File, extension string) (newFile domain.File, contentType string, err error) {
	_, err = uc.verifyExtensionContentType(extension)
	if err != nil {
		return nil, "", err
	}

	detectedContentType, newFile, err := uc.detector.DetectFileContentType(file)
	if err != nil {
		return nil, "", err
	}
	if !domain.IsTypeImageOrVideo(detectedContentType) {
		return nil, detectedContentType, fmt.Errorf("detected content type (%v): %w",
			detectedContentType, domain.ErrIncorrectContentType)
	}

	return newFile, detectedContentType, nil
}

func (uc *ImageUC) verifyReaderContentType(file io.ReadCloser, extension string) (newFile io.ReadCloser, contentType string, err error) {
	_, err = uc.verifyExtensionContentType(extension)
	if err != nil {
		return nil, "", err
	}

	detectedContentType, newFile, err := uc.detector.DetectReaderContentType(file)
	if err != nil {
		return nil, "", err
	}
	if !domain.IsTypeImageOrVideo(detectedContentType) {
		return nil, detectedContentType, fmt.Errorf("detected content type (%v): %w",
			detectedContentType, domain.ErrIncorrectContentType)
	}

	return newFile, detectedContentType, nil
}
