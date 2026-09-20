package authentication

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	ErrExpiredToken  = errors.New("token is expired")
	ErrInvalidToken  = errors.New("token is invalid")
	ErrMissingUserID = errors.New("user id is missing from token claims")
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
				return nil, ErrInvalidToken
			}
			return p.SecretKey, nil
		})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.UUID{}, ErrExpiredToken
		}
		return uuid.UUID{}, ErrInvalidToken
	}

	if token == nil || !token.Valid {
		return uuid.UUID{}, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.UUID{}, ErrInvalidToken
	}

	if claims.UserID == uuid.Nil {
		return uuid.UUID{}, ErrMissingUserID
	}
	return claims.UserID, nil
}
