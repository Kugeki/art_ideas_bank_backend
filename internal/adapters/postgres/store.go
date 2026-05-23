package postgres

import (
	"art_ideas_bank_backend/pkg/slogdiscard"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"log/slog"
)

type Store struct {
	db  *pgxpool.Pool
	log *slog.Logger

	userRepo  *UserRepoPG
	imageRepo *ImageRepoPG
	tagRepo   *TagRepoPG
}

func NewStore(ctx context.Context, log *slog.Logger, dbURL string) (*Store, error) {
	s := &Store{log: slogdiscard.LoggerIfNil(log)}

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newLogger(s.log),
		LogLevel: fromSlogLevel(getSlogLevel(ctx, s.log)),
	}

	db, err := connect(ctx, cfg)
	if err != nil {
		return nil, err
	}

	s.db = db

	return s, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) UserRepo() *UserRepoPG {
	if s.userRepo == nil {
		s.userRepo = NewUserRepo(s.db, s.log)
	}
	return s.userRepo
}

func (s *Store) ImageRepo() *ImageRepoPG {
	if s.imageRepo == nil {
		s.imageRepo = NewImageRepo(s.db, s.log)
	}
	return s.imageRepo
}

func (s *Store) TagRepo() *TagRepoPG {
	if s.tagRepo == nil {
		s.tagRepo = NewTagRepo(s.db, s.log)
	}
	return s.tagRepo
}

func connect(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
