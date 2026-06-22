package postgres

import (
	"context"
	"fmt"
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
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
	tagErr := &domain.TagError{Path: fullPath}

	parts := strings.Split(fullPath, ".")
	if len(parts) == 0 || fullPath == "" {
		return nil, domain.ErrEmptyTagPath
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var currentPath string
	for i, part := range parts {
		name := sanitizeTagName(part)
		if name == "" {
			return nil, tagErr.With(fmt.Errorf("%w '%s' in poisition %d (from 1)", domain.ErrTagIncorrectName, part, i+1))
		}
		if i == 0 {
			currentPath = name
		} else {
			currentPath = currentPath + "." + name
		}

		_, err = tx.Exec(ctx,
			`insert into tags (user_id, path, name) values ($1, $2, $3) on conflict (user_id, path) do nothing`,
			userID, currentPath, name,
		)
		if err != nil {
			return nil, tagErr.With(fmt.Errorf("tag creation error: %w", err))
		}
	}

	tagErr.Path = currentPath

	var tag domain.Tag
	err = tx.QueryRow(ctx,
		`select id, user_id, path::text, name from tags where user_id = $1 and path = $2`,
		userID, currentPath,
	).Scan(&tag.ID, &tag.UserID, &tag.Path, &tag.Name)
	if err != nil {
		return nil, tagErr.With(fmt.Errorf("cannot get created tag: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &tag, nil
}

func (r *TagRepoPG) ListByUser(ctx context.Context, userID int) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx,
		`select id, user_id, path::text, name from tags where user_id = $1 order by path`,
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
	err = rows.Err()
	if err != nil {
		return tags, rows.Err()
	}

	return tags, nil
}

// GetTagsByPaths возвращает теги по точным путям (ltree) в рамках пользователя.
func (r *TagRepoPG) GetTagsByPaths(ctx context.Context, userID int, paths []string) ([]domain.Tag, error) {
	if len(paths) <= 0 {
		return []domain.Tag{}, nil
	}

	rows, err := r.db.Query(ctx,
		`select id, user_id, path::text, name from tags
         where user_id = $1 and path::text = any($2)`,
		userID, paths,
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
	if len(tags) != len(paths) {
		return nil, domain.ErrSomeTagsNotFound
	}

	err = rows.Err()
	if err != nil {
		return tags, rows.Err()
	}

	return tags, nil
}

// DeleteTag удаляет тег и все его поддеревья, только если на них нет ссылок из image_tags.
func (r *TagRepoPG) DeleteTag(ctx context.Context, userID int, tagID string) error {
	tagErr := &domain.TagError{ID: tagID}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var tagPath string
	err = tx.QueryRow(ctx,
		`select path::text from tags where id = $1 and user_id = $2`,
		tagID, userID,
	).Scan(&tagPath)
	if err != nil {
		return tagErr.With(DomainGetError(err))
	}

	var count int
	err = tx.QueryRow(ctx,
		`select count(*) from image_tags it
         join tags t on it.tag_id = t.id
         where t.path <@ $1 and t.user_id = $2`,
		tagPath, userID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return tagErr.With(domain.ErrTagHasAssociatedImages)
	}

	_, err = tx.Exec(ctx,
		`delete from tags where path <@ $1 AND user_id = $2`,
		tagPath, userID,
	)
	if err != nil {
		return tagErr.With(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// UpdateTag обновляет имя и/или родителя тега, перемещая всё поддерево.
func (r *TagRepoPG) UpdateTag(ctx context.Context, userID int, tagID string, newName string, newParentID *string) (*domain.Tag, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var oldPath, oldName string
	err = tx.QueryRow(ctx,
		`select path::text, name from tags where id = $1 and user_id = $2`,
		tagID, userID,
	).Scan(&oldPath, &oldName)
	if err != nil {
		return nil, &domain.TagError{Err: DomainGetError(err)}
	}

	finalName := oldName
	if newName != "" {
		finalName = sanitizeTagName(newName)
	}

	var newParentPath string
	if newParentID != nil {
		err = tx.QueryRow(ctx,
			`select path::text from tags where id = $1 and user_id = $2`,
			*newParentID, userID,
		).Scan(&newParentPath)
		if err != nil {
			return nil, fmt.Errorf("new parent tag: %w", DomainGetError(err))
		}
		isDescendant := true
		err = tx.QueryRow(ctx,
			`select $1 <@ (select path from tags where id = $2)`,
			newParentPath, tagID,
		).Scan(&isDescendant)
		if err != nil {
			return nil, err
		}
		if isDescendant {
			return nil, domain.ErrCannotMoveTagToDescendant
		}
	} else {
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

	if newSelfPath != oldPath {
		var exists bool
		err = tx.QueryRow(ctx,
			`select exists(select 1 from tags where user_id = $1 and path = $2 and id != $3)`,
			userID, newSelfPath, tagID,
		).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("tag with path '%s': %w", newSelfPath, domain.ErrAlreadyExists)
		}
	}

	r.log.Info("updating tag path", slog.String("new_self_path", newSelfPath),
		slog.String("old_path", oldPath), slog.Int("user_id", userID))

	_, err = tx.Exec(ctx,
		`update tags set path = $1, name = $2 where id = $3 and user_id = $4`,
		newSelfPath, finalName, tagID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("tag update error: %w", err)
	}

	// Обновляем всех потомков тега
	_, err = tx.Exec(ctx,
		`update tags
     			set path = $1 || subpath(path, nlevel($2))
     			where path <@ $2 and id != $3 and user_id = $4`,
		newSelfPath, oldPath, tagID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("tag descendants update error: %w", err)
	}

	var updated domain.Tag
	err = tx.QueryRow(ctx,
		`select id, user_id, path::text, name from tags where id = $1`,
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

func sanitizeTagName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

func (r *TagRepoPG) SuggestTags(ctx context.Context, userID int, prefix string, limit int) ([]domain.Tag, error) {
	if limit <= 0 {
		return nil, domain.ErrLimitEqualOrLessThanZero
	}

	rows, err := r.db.Query(ctx,
		`select id, user_id, path::text, name from tags
         where user_id = $1 and path::text like $2
         order by path
         limit $3`,
		userID, prefix+"%", limit,
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
		return tags, rows.Err()
	}

	return tags, nil
}
