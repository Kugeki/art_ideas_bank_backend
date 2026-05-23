package postgres

import (
	"art_ideas_bank_backend/internal/domain"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strings"
)

type TagRepoPG struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewTagRepo(db *pgxpool.Pool, log *slog.Logger) *TagRepoPG {
	return &TagRepoPG{db: db, log: log.With(slog.String("repository", "tag"))}
}

func (r *TagRepoPG) CreateTag(ctx context.Context, userID int, fullPath string) (*domain.Tag, error) {
	parts := strings.Split(fullPath, ".")
	if len(parts) == 0 || fullPath == "" {
		return nil, fmt.Errorf("путь не может быть пустым")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Последовательно создаём каждый уровень
	var currentPath string
	for i, part := range parts {
		name := sanitizeTagName(part)
		if name == "" {
			return nil, fmt.Errorf("некорректное имя '%s' в позиции %d", part, i+1)
		}
		if i == 0 {
			currentPath = name
		} else {
			currentPath = currentPath + "." + name
		}

		// Пытаемся вставить, игнорируем конфликт (уже существует)
		_, err = tx.Exec(ctx,
			`INSERT INTO tags (user_id, path, name) VALUES ($1, $2, $3) ON CONFLICT (user_id, path) DO NOTHING`,
			userID, currentPath, name,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания тега '%s': %w", currentPath, err)
		}
	}

	// Получаем последний созданный тег
	var tag domain.Tag
	err = tx.QueryRow(ctx,
		`SELECT id, user_id, path::text, name FROM tags WHERE user_id = $1 AND path = $2`,
		userID, currentPath,
	).Scan(&tag.ID, &tag.UserID, &tag.Path, &tag.Name)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить созданный тег: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepoPG) ListByUser(ctx context.Context, userID int) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, path::text, name FROM tags WHERE user_id = $1 ORDER BY path`,
		userID,
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
	return tags, rows.Err()
}

func (r *TagRepoPG) GetTagsForImage(ctx context.Context, userID int, imageID string) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.user_id, t.path::text, t.name
         FROM tags t
         JOIN image_tags it ON t.id = it.tag_id
         JOIN images i ON i.id = it.image_id
         WHERE i.id = $1 AND i.user_id = $2
         ORDER BY t.path`,
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
	return tags, rows.Err()
}

// DeleteTag удаляет тег и все его поддеревья, только если на них нет ссылок из image_tags.
func (r *TagRepoPG) DeleteTag(ctx context.Context, userID int, tagID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Получаем путь тега
	var tagPath string
	err = tx.QueryRow(ctx,
		`SELECT path::text FROM tags WHERE id = $1 AND user_id = $2`,
		tagID, userID,
	).Scan(&tagPath)
	if err != nil {
		return fmt.Errorf("тег не найден: %w", err)
	}

	// Проверяем, есть ли изображения у этого тега или его потомков
	var count int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM image_tags it
         JOIN tags t ON it.tag_id = t.id
         WHERE t.path <@ $1 AND t.user_id = $2`,
		tagPath, userID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("нельзя удалить тег, так как к нему или его потомкам привязаны изображения")
	}

	// Удаляем само поддерево
	_, err = tx.Exec(ctx,
		`DELETE FROM tags WHERE path <@ $1 AND user_id = $2`,
		tagPath, userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdateTag обновляет имя и/или родителя тега, перемещая всё поддерево.
func (r *TagRepoPG) UpdateTag(ctx context.Context, userID int, tagID string, newName string, newParentID *string) (*domain.Tag, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Текущий путь
	var oldPath, oldName string
	err = tx.QueryRow(ctx,
		`SELECT path::text, name FROM tags WHERE id = $1 AND user_id = $2`,
		tagID, userID,
	).Scan(&oldPath, &oldName)
	if err != nil {
		return nil, fmt.Errorf("тег не найден: %w", err)
	}

	finalName := oldName
	if newName != "" {
		finalName = sanitizeTagName(newName) // функция очистки имени (ниже)
	}

	// Вычисляем новый родительский путь
	var newParentPath string
	if newParentID != nil {
		err = tx.QueryRow(ctx,
			`SELECT path::text FROM tags WHERE id = $1 AND user_id = $2`,
			*newParentID, userID,
		).Scan(&newParentPath)
		if err != nil {
			return nil, fmt.Errorf("родительский тег не найден: %w", err)
		}
		// Проверка на цикл: новый родитель не должен быть потомком текущего тега
		isDescendant := false
		err = tx.QueryRow(ctx,
			`SELECT $1 <@ (SELECT path FROM tags WHERE id = $2)`,
			newParentPath, tagID,
		).Scan(&isDescendant)
		if err != nil {
			return nil, err
		}
		if isDescendant {
			return nil, fmt.Errorf("нельзя переместить тег в своего потомка")
		}
	} else {
		// Без смены родителя путь вычисляется из старого, заменой последнего компонента
		parts := strings.Split(oldPath, ".")
		if len(parts) == 1 {
			newParentPath = ""
		} else {
			newParentPath = strings.Join(parts[:len(parts)-1], ".")
		}
	}

	var newSelfPath string
	if newParentPath == "" {
		newSelfPath = finalName
	} else {
		newSelfPath = newParentPath + "." + finalName
	}

	// Проверяем, не существует ли уже такой путь у другого тега
	if newSelfPath != oldPath {
		var exists bool
		err = tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM tags WHERE user_id = $1 AND path = $2 AND id != $3)`,
			userID, newSelfPath, tagID,
		).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("тег с путём '%s' уже существует", newSelfPath)
		}
	}

	// Обновляем пути поддерева: заменяем старый префикс на новый
	_, err = tx.Exec(ctx,
		`UPDATE tags SET path = $1 || subpath(path, nlevel($2)-1)
    	 WHERE path <@ $2 AND user_id = $3`,
		newSelfPath, oldPath, userID,
	)
	if err != nil {
		return nil, err
	}

	var updated domain.Tag
	err = tx.QueryRow(ctx,
		`SELECT id, user_id, path::text, name FROM tags WHERE id = $1`,
		tagID,
	).Scan(&updated.ID, &updated.UserID, &updated.Path, &updated.Name)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &updated, nil
}

// Вспомогательная очистка имени
func sanitizeTagName(name string) string {
	name = strings.TrimSpace(name)
	// Оставляем только буквы, цифры, -, _
	filtered := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, name)
	return filtered
}

// AddTagsToImage привязывает несколько тегов к изображению (с проверкой владения)
func (r *TagRepoPG) AddTagsToImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	// Проверяем, что изображение принадлежит пользователю
	var ownerID int
	err := r.db.QueryRow(ctx, `SELECT user_id FROM images WHERE id = $1`, imageID).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("изображение не найдено")
	}
	if ownerID != userID {
		return fmt.Errorf("доступ запрещён")
	}

	// Проверяем, что все теги принадлежат пользователю
	var count int
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tags WHERE id = ANY($1) AND user_id = $2`,
		tagIDs, userID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count != len(tagIDs) {
		return fmt.Errorf("один из тегов не принадлежит вам или не существует")
	}

	// Вставляем связи, игнорируя существующие
	for _, tagID := range tagIDs {
		_, err = r.db.Exec(ctx,
			`INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			imageID, tagID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveTagsFromImage удаляет связи с указанными тегами у изображения
func (r *TagRepoPG) RemoveTagsFromImage(ctx context.Context, userID int, imageID string, tagIDs []string) error {
	var ownerID int
	err := r.db.QueryRow(ctx, `SELECT user_id FROM images WHERE id = $1`, imageID).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("изображение не найдено")
	}
	if ownerID != userID {
		return fmt.Errorf("доступ запрещён")
	}

	_, err = r.db.Exec(ctx,
		`DELETE FROM image_tags WHERE image_id = $1 AND tag_id = ANY($2)`,
		imageID, tagIDs,
	)
	return err
}

// SearchImagesByTags выполняет конъюнктивный поиск изображений пользователя.
// tagIDs – список ID тегов; изображение должно иметь хотя бы один тег из поддерева каждого из указанных.
func (r *TagRepoPG) SearchImagesByTags(ctx context.Context, userID int, tagIDs []string) ([]domain.Image, error) {
	if len(tagIDs) == 0 {
		return nil, fmt.Errorf("необходимо указать хотя бы один тег")
	}

	var query strings.Builder
	query.WriteString(`
        SELECT i.id, i.user_id, i.s3_key, i.uploaded_at
        FROM images i
        WHERE i.user_id = $1
    `)
	args := []interface{}{userID}
	argIdx := 2

	for _, tagID := range tagIDs {
		query.WriteString(fmt.Sprintf(`
            AND EXISTS (
                SELECT 1
                FROM image_tags it
                JOIN tags t ON it.tag_id = t.id
                WHERE it.image_id = i.id
                  AND t.path <@ (SELECT path FROM tags WHERE id = $%d AND user_id = $1)
            )`, argIdx))
		args = append(args, tagID)
		argIdx++
	}
	query.WriteString(` ORDER BY i.uploaded_at DESC`)

	rows, err := r.db.Query(ctx, query.String(), args...)
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
	return images, rows.Err()
}
