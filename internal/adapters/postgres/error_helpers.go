package postgres

import (
	"art_ideas_bank_backend/internal/domain"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgErrUniqueViolation https://www.postgresql.org/docs/16/errcodes-appendix.html
const PgErrUniqueViolation = "23505"

func DomainCreateError(dbErr error) error {
	var pgErr *pgconn.PgError
	if errors.As(dbErr, &pgErr) {
		if pgErr.Code == PgErrUniqueViolation {
			return domain.ErrAlreadyExists
		}
	}

	return dbErr
}

func DomainGetError(dbErr error) error {
	if errors.Is(dbErr, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	return dbErr
}
