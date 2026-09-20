package authentication

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
)

var (
	ErrTokenIsInvalid    = errors.New("token is not valid")
	ErrNoUserLoginExists = errors.New("user login is not exists in token claims")
	ErrExpiredToken      = errors.New("token is expired")
	ErrInvalidToken      = errors.New("token is invalid")
)

type Parser struct {
	SecretKey []byte
}

func NewUserIDParser(secretKey []byte) *Parser {
	return &Parser{
		SecretKey: secretKey,
	}
}

func (p Parser) ValidateUserID(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, ErrTokenIsInvalid
			}
			return p.SecretKey, nil
		})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.UUID{}, ErrExpiredToken
		}
		return uuid.UUID{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.UUID{}, ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return uuid.UUID{}, ErrExpiredToken
	}
	if claims.UserID == uuid.Nil {
		return uuid.UUID{}, ErrNoUserLoginExists
	}
	return claims.UserID, nil
}
