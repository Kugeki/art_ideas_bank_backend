package postgres

import (
	"art_ideas_bank_backend/internal/domain"
	"art_ideas_bank_backend/pkg/slogdiscard"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type UserRepoPG struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewUserRepo(db *pgxpool.Pool, log *slog.Logger) *UserRepoPG {
	log = slogdiscard.LoggerIfNil(log)
	return &UserRepoPG{db: db, log: log.With(slog.String("repository", "user"))}
}

func (r *UserRepoPG) CreateUser(ctx context.Context, u *domain.User) error {
	q1 := `insert into passwords
		(hash_base64, argon2_version, argon2_type, salt_base64, argon2_time, argon2_memory, argon2_threads, argon2_keylen) 
		values($1, $2, $3, $4, $5, $6, $7, $8) returning id`

	q2 := "insert into users(email, password_id) values($1, $2) returning id"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		if errors.Is(err, pgx.ErrTxClosed) {
			r.log.Debug("transaction is closed", slog.Any("error", err))
			return
		}
		if err := tx.Rollback(ctx); err != nil {
			r.log.Error("transaction rollback error", slog.Any("error", err))
		}
	}(tx, ctx)

	p := u.Password

	var passwordID int
	err = r.db.QueryRow(ctx, q1, p.HashBase64, p.Argon2Version, p.Argon2Type,
		p.SaltBase64, p.Time, p.Memory, p.Threads, p.KeyLen).Scan(&passwordID)
	if err != nil {
		return DomainCreateError(err)
	}

	var userID int
	err = r.db.QueryRow(ctx, q2, u.Email, passwordID).Scan(&userID)
	if err != nil {
		return DomainCreateError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	u.ID = userID
	u.Password.ID = passwordID

	return nil
}

func (r *UserRepoPG) GetUser(ctx context.Context, email string) (*domain.User, error) {
	q := `select p.hash_base64, p.argon2_version, p.argon2_type, 
       		p.salt_base64, p.argon2_time, p.argon2_memory, p.argon2_threads, p.argon2_keylen, u.id 
		from users u
			join passwords p on p.id = u.password_id 
         where u.email = $1`

	u := domain.User{Email: email}
	p := domain.Password{}

	err := r.db.QueryRow(ctx, q, email).
		Scan(&p.HashBase64, &p.Argon2Version, &p.Argon2Type,
			&p.SaltBase64, &p.Time, &p.Memory, &p.Threads, &p.KeyLen, &u.ID)
	if err != nil {
		return nil, DomainGetError(err)
	}

	u.Password = p

	return &u, nil
}
