package jwtoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID string

	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
}

func NewTokenService(secret []byte) *TokenService {
	return &TokenService{
		secret: secret,
	}
}

func (t *TokenService) GenerateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,

		RegisteredClaims: jwt.RegisteredClaims{
			// Выдаём пока что на год
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 365)),

			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *TokenService) ValidateToken(tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected jwt signing method")
		}
		return t.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil

}
