package postgres

import (
	"context"
	"fmt"
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strings"
)

type ImageRepoPG struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewImageRepo(db *pgxpool.Pool, log *slog.Logger) *ImageRepoPG {
	return &ImageRepoPG{db: db, log: log.With(slog.String("repository", "image"))}
}

func (r *ImageRepoPG) Create(ctx context.Context, img *domain.Image) error {
	q := `insert into images (user_id, ext, s3_key) values ($1, $2, $3) returning id, uploaded_at`
	err := r.db.QueryRow(ctx,
		q,
		img.UserID, img.Extension, img.S3Key,
	).Scan(&img.ID, &img.UploadedAt)
	if err != nil {
		return &domain.ImageError{Err: DomainCreateError(err)}
	}

	return nil
}

func (r *ImageRepoPG) DeleteImage(ctx context.Context, userID int, imageID string) error {
	_, err := r.db.Exec(ctx,
		`delete from images where id = $1 and user_id = $2`,
		imageID, userID,
	)
	if err != nil {
		return &domain.ImageError{Err: err, ID: imageID}
	}

	return nil
}

func (r *ImageRepoPG) GetByID(ctx context.Context, userID int, imageID string) (*domain.Image, error) {
	var img domain.Image
	err := r.db.QueryRow(ctx,
		`select id, user_id, ext, s3_key, uploaded_at from images where id = $1 and user_id = $2`,
		imageID, userID,
	).Scan(&img.ID, &img.UserID, &img.Extension, &img.S3Key, &img.UploadedAt)
	if err != nil {
		return nil, &domain.ImageError{ID: imageID, Err: DomainGetError(err)}
	}

	return &img, nil
}

func (r *ImageRepoPG) ListByUser(ctx context.Context, userID int) ([]domain.Image, error) {
	q := `select id, user_id, ext, s3_key, uploaded_at from images where user_id = $1 order by uploaded_at desc`

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
		if err := rows.Scan(&img.ID, &img.UserID, &img.Extension, &img.S3Key, &img.UploadedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	err = rows.Err()
	if err != nil {
		return images, DomainGetError(rows.Err())
	}

	return images, nil
}

func (r *ImageRepoPG) GetByKey(ctx context.Context, key string) (*domain.Image, error) {
	var img domain.Image
	err := r.db.QueryRow(ctx,
		`select id, user_id, ext, s3_key, uploaded_at from images where s3_key = $1`,
		key,
	).Scan(&img.ID, &img.UserID, &img.Extension, &img.S3Key, &img.UploadedAt)
	if err != nil {
		return nil, DomainGetError(err)
	}

	return &img, nil
}

func (r *ImageRepoPG) GetTagsForImage(ctx context.Context, userID int, imageID string) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx,
		`select t.id, t.user_id, t.path::text, t.name
         from tags t
         join image_tags it ON t.id = it.tag_id
         join images i ON i.id = it.image_id
         where i.id = $1 AND i.user_id = $2
         order by t.path`,
		imageID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Path, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	err = rows.Err()
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (r *ImageRepoPG) AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	var ownerID int
	err := r.db.QueryRow(ctx, `select user_id from images where id = $1`, imageID).Scan(&ownerID)
	if err != nil {
		return &domain.ImageError{ID: imageID, Err: DomainGetError(err)}
	}
	if ownerID != userID {
		return &domain.ImageError{ID: imageID, Err: domain.ErrNotFound}
	}

	var count int
	err = r.db.QueryRow(ctx,
		`select count(*) from tags where id = any ($1) and user_id = $2`,
		tagIDs, userID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count != len(tagIDs) {
		return domain.ErrSomeTagsNotFound
	}

	for _, tagID := range tagIDs {
		_, err = r.db.Exec(ctx,
			`insert into image_tags (image_id, tag_id) values ($1, $2) on conflict do nothing`,
			imageID, tagID,
		)
		if err != nil {
			return err
		}
		log.Info("tag inserted into image", slog.String("image_id", imageID), slog.String("tag_id", tagID))
	}

	return nil
}

func (r *ImageRepoPG) RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	var ownerID int
	err := r.db.QueryRow(ctx, `select user_id from images where id = $1`, imageID).Scan(&ownerID)
	if err != nil {
		return &domain.ImageError{ID: imageID, Err: DomainGetError(err)}
	}
	if ownerID != userID {
		return &domain.ImageError{ID: imageID, Err: domain.ErrNotFound}
	}

	_, err = r.db.Exec(ctx,
		`delete from image_tags where image_id = $1 and tag_id = any($2)`,
		imageID, tagIDs,
	)
	if err != nil {
		return err
	}

	return nil
}

// SearchImagesByTags выполняет конъюнктивный поиск изображений пользователя.
// tagIDs – список ID тегов; изображение должно иметь хотя бы один тег из поддерева каждого из указанных.
func (r *ImageRepoPG) SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error) {
	var query strings.Builder
	query.WriteString(`
        select i.id, i.user_id, i.s3_key, i.uploaded_at
        from images i
        where i.user_id = $1
    `)
	args := []interface{}{userID}
	argIdx := 2

	for _, tagID := range tagIDs {
		query.WriteString(fmt.Sprintf(`
            and exists (
                select 1
                from image_tags it
                join tags t on it.tag_id = t.id
                where it.image_id = i.id
                  and t.path <@ (select path from tags where id = $%d and user_id = $1)
            )`, argIdx))
		args = append(args, tagID)
		argIdx++
	}
	query.WriteString(` order by i.uploaded_at desc`)

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]domain.Image, 0)
	for rows.Next() {
		var img domain.Image
		if err := rows.Scan(&img.ID, &img.UserID, &img.S3Key, &img.UploadedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	err = rows.Err()
	if err != nil {
		return images, rows.Err()

	}

	return images, nil
}
