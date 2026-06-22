package jwtauth

import (
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWT struct {
	secretKey []byte
}

func NewJWT(secretKey string) *JWT {
	return &JWT{secretKey: []byte(secretKey)}
}

func (j *JWT) GenerateToken(userID int) (string, error) {
	now := time.Now()

	claims := domain.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWT) ParseToken(tokenStr string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &domain.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if claims, ok := token.Claims.(*domain.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, domain.ErrInvalidToken
}
