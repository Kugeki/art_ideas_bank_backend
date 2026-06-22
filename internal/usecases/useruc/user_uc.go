package useruc

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"golang.org/x/crypto/argon2"
)

type UserRepo interface {
	CreateUser(ctx context.Context, u *domain.User) error
	GetUser(ctx context.Context, email string) (*domain.User, error)
}

type UserUC struct {
	userRepo UserRepo
}

func New(repo UserRepo) (*UserUC, error) {
	return &UserUC{userRepo: repo}, nil
}

const (
	DefaultSaltSize      = 16 // SaltSize recommended is 16: https://datatracker.ietf.org/doc/html/rfc9106#name-argon2-inputs-and-outputs
	DefaultArgon2Time    = 1
	DefaultArgon2Memory  = 64 * 1024
	DefaultArgon2Threads = 4
	DefaultArgon2KeyLen  = 32
)

func (uc *UserUC) CreateUser(ctx context.Context, u *domain.User, password string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	salt := make([]byte, DefaultSaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	pwHash := argon2.IDKey([]byte(password), salt, DefaultArgon2Time,
		DefaultArgon2Memory, DefaultArgon2Threads, DefaultArgon2KeyLen)

	pwHashBase64 := base64.StdEncoding.EncodeToString(pwHash)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)

	u.Password = domain.Password{
		HashBase64:    pwHashBase64,
		Argon2Version: argon2.Version,
		Argon2Type:    domain.Argon2idType,
		SaltBase64:    saltBase64,
		Time:          DefaultArgon2Time,
		Memory:        DefaultArgon2Memory,
		Threads:       DefaultArgon2Threads,
		KeyLen:        DefaultArgon2KeyLen,
	}

	err = uc.userRepo.CreateUser(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UserUC) VerifyUser(ctx context.Context, email string, password string) (*domain.User, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	u, err := uc.userRepo.GetUser(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrWrongCredentials
	}
	if err != nil {
		return nil, err
	}

	pw := u.Password

	wantHash, err := base64.StdEncoding.DecodeString(pw.HashBase64)
	if err != nil {
		return nil, err
	}

	salt, err := base64.StdEncoding.DecodeString(pw.SaltBase64)
	if err != nil {
		return nil, err
	}

	gotHash := argon2.IDKey([]byte(password), salt, pw.Time, pw.Memory, pw.Threads, pw.KeyLen)

	if subtle.ConstantTimeCompare(wantHash, gotHash) != 1 {
		return nil, domain.ErrWrongCredentials
	}

	return u, nil
}
