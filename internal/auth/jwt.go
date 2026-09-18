package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrTokenIsInvalid    = errors.New("token is not valid")
	ErrNoUserLoginExists = errors.New("user login is not exists in token claims")
)

type TokenBuilder interface {
	BuildJWTString(userID string, tokenExp time.Duration) (string, error)
}

type JWTBuilder struct {
	SecretKey []byte
}

func NewJWTBuilder(secretKey []byte) *JWTBuilder {
	return &JWTBuilder{
		SecretKey: secretKey,
	}
}

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserLogin
type Claims struct {
	jwt.RegisteredClaims
	UserLogin string
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (b JWTBuilder) BuildJWTString(userLogin string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		// собственное утверждение
		UserLogin: userLogin,
	})

	// создаём строку токена
	tokenString, err := token.SignedString(b.SecretKey)
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}
