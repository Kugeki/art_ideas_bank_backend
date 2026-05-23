package postgres

import (
	"art_ideas_bank_backend/internal/domain"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type ImageRepoPG struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewImageRepo(db *pgxpool.Pool, log *slog.Logger) *ImageRepoPG {
	return &ImageRepoPG{db: db, log: log.With(slog.String("repository", "image"))}
}

func (r *ImageRepoPG) Create(ctx context.Context, img *domain.Image) error {
	q := `insert into images (user_id, s3_key) values ($1, $2) returning id, uploaded_at`
	err := r.db.QueryRow(ctx,
		q,
		img.UserID, img.S3Key,
	).Scan(&img.ID, &img.UploadedAt)
	return DomainCreateError(err)
}

func (r *ImageRepoPG) GetByID(ctx context.Context, userID int, imageID string) (*domain.Image, error) {
	var img domain.Image
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, s3_key, uploaded_at FROM images WHERE id = $1 AND user_id = $2`,
		imageID, userID,
	).Scan(&img.ID, &img.UserID, &img.S3Key, &img.UploadedAt)
	if err != nil {
		return nil, DomainGetError(err)
	}
	return &img, nil
}

func (r *ImageRepoPG) ListByUser(ctx context.Context, userID int) ([]domain.Image, error) {
	q := `select id, user_id, s3_key, uploaded_at from images where user_id = $1 order by uploaded_at desc`

	rows, err := r.db.Query(ctx,
		q,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.Image
	for rows.Next() {
		var img domain.Image
		if err := rows.Scan(&img.ID, &img.UserID, &img.S3Key, &img.UploadedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, DomainGetError(rows.Err())
}

func (r *ImageRepoPG) GetByKey(ctx context.Context, key string) (*domain.Image, error) {
	var img domain.Image
	err := r.db.QueryRow(ctx,
		`select id, user_id, s3_key, uploaded_at from images where s3_key = $1`,
		key,
	).Scan(&img.ID, &img.UserID, &img.S3Key, &img.UploadedAt)
	if err != nil {
		return nil, DomainGetError(err)
	}
	return &img, nil
}
